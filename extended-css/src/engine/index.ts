import { parse } from '../parse';
import { createLogger } from '../utils/logger';
import { throttle } from '../utils/throttle';

import { SelectorExecutor } from './selectorExecutor';

const logger = createLogger('engine');

export class Engine {
  private readonly executors: SelectorExecutor[];
  private readonly target = document.documentElement;

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
    let hiddenCnt = 0;

    for (const ex of this.executors) {
      try {
        const els = ex.match(this.target);
        for (const el of els) {
          if (!(el instanceof HTMLElement)) continue;
          el.style.setProperty('display', 'none', 'important');
        }
        hiddenCnt += els.length;
      } catch (ex) {
        logger.error(`Failed to apply rule`, ex);
      }
    }

    const end = performance.now();
    logger.debug(`Hidden ${hiddenCnt} elements in ${(end - start).toFixed(2)}ms`);
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

    const observer = new MutationObserver((mutations, observer) => {
      if (mutations.length === 0) return;
      if (mutations.every((m) => m.type === 'attributes')) return;

      cb(observer);
    });

    observer.observe(this.target, options);
  }
}
