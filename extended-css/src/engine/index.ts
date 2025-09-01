import { parse } from '../parse';
import { Query } from '../parse/types';
import { createLogger } from '../utils/logger';
import { throttle } from '../utils/throttle';

const logger = createLogger('engine');

export class Engine {
  private readonly queries: Query[];
  private readonly target = document.documentElement;

  constructor(rules: string) {
    logger.debug('Initializing engine');
    this.queries = this.parseRules(rules);
  }

  start(): void {
    logger.debug(`Starting with ${this.queries.length} queries`);

    this.applyQueries();

    if (document.readyState !== 'complete') {
      document.addEventListener('DOMContentLoaded', () => {
        this.applyQueries();
      });
    }

    this.registerObserver();
  }

  private parseRules(rules: string): Query[] {
    return rules
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
      .map((line) => {
        try {
          return parse(line);
        } catch (ex) {
          logger.error(`Failed to parse rule: "${line}"`, ex);
          return undefined;
        }
      })
      .filter((q): q is Query => q !== undefined);
  }

  private applyQueries(): void {
    const start = performance.now();
    let removedCnt = 0;

    for (const query of this.queries) {
      let els: Element[] = [this.target];
      for (const step of query) {
        els = step.run(els);
        if (els.length === 0) break;
      }
      for (const el of els) el.remove();
      removedCnt += els.length;
    }

    const end = performance.now();
    logger.debug(`Removed ${removedCnt} elements in ${(end - start).toFixed(2)}ms`);
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
