import { parse } from '../parse';
import { createLogger } from '../utils/logger';
import { throttle } from '../utils/throttle';

import { SelectorExecutor } from './selectorExecutor';

const logger = createLogger('engine');

export class Engine {
  private readonly executors: SelectorExecutor[];
  private readonly target = document.documentElement;

  /**
   * Tracks original inline display values for elements we hide.
   */
  private readonly hiddenOriginalDisplay = new Map<Element, { value: string; important: boolean }>();
  private observer: MutationObserver | null = null;

  constructor(rules: string) {
    logger.debug('Initializing engine');
    this.executors = this.parseRules(rules);
  }

  start(): void {
    logger.debug(`Starting with ${this.executors.length} rules`);

    this.applyQueries();

    if (document.readyState !== 'complete') {
      document.addEventListener(
        'DOMContentLoaded',
        () => {
          this.applyQueries();
        },
        { once: true },
      );
    }

    this.registerObserver();
  }

  /**
   * Tears down the mutation observer and restores styles of all affected elements to their original state.
   */
  stop(): void {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }

    for (const el of this.hiddenOriginalDisplay.keys()) {
      this.restoreElement(el);
    }

    logger.debug('Engine stopped');
  }

  private parseRules(rules: string): SelectorExecutor[] {
    const lines = rules.split('\n');

    const executors = [];
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed.length === 0) continue;

      try {
        const selectorList = parse(trimmed);
        executors.push(new SelectorExecutor(selectorList));
      } catch (ex) {
        logger.error(`Failed to parse rule: "${line}"`, ex);
      }
    }

    return executors;
  }

  private applyQueries(): void {
    const start = performance.now();

    // Compute the union of all current matches.
    const currentMatches = new Set<Element>();
    for (const ex of this.executors) {
      try {
        const els = ex.match(this.target);
        for (const el of els) currentMatches.add(el);
      } catch (ex) {
        logger.error(`Failed to apply rule`, ex);
      }
    }

    let restored = 0;
    for (const el of this.hiddenOriginalDisplay.keys()) {
      if (!currentMatches.has(el)) {
        this.restoreElement(el);
        restored++;
      }
    }

    // Hide all currently matched elements, recording original inline display once.
    let newlyHidden = 0;
    for (const el of currentMatches) {
      if (!(el instanceof HTMLElement)) continue;
      if (this.hiddenOriginalDisplay.has(el)) continue;

      this.hiddenOriginalDisplay.set(el, {
        value: el.style.getPropertyValue('display'),
        important: el.style.getPropertyPriority('display') === 'important',
      });

      el.style.setProperty('display', 'none', 'important');

      newlyHidden++;
    }

    const end = performance.now();
    logger.debug(`Hidden ${newlyHidden} elements, restored ${restored} in ${(end - start).toFixed(2)}ms`);
  }

  private restoreElement(el: Element): void {
    if (!(el instanceof HTMLElement)) return;
    const original = this.hiddenOriginalDisplay.get(el);
    if (original === undefined) return;

    if (original.value) {
      el.style.setProperty('display', original.value, original.important ? 'important' : undefined);
    } else {
      el.style.removeProperty('display');
    }

    this.hiddenOriginalDisplay.delete(el);
  }

  private registerObserver(): void {
    const options: MutationObserverInit = {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['id', 'class'],
    };

    const cb = throttle((observer: MutationObserver) => {
      observer.disconnect();
      this.applyQueries();
      observer.observe(this.target, options);
    }, 100);

    this.observer = new MutationObserver((mutations, observer) => {
      if (mutations.length === 0) return;
      if (mutations.every((m) => m.type === 'attributes')) return;

      cb(observer);
    });

    this.observer.observe(this.target, options);
  }
}
