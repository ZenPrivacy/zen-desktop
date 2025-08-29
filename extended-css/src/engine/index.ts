import { parse } from '../parse';
import { Query } from '../parse/types';

export class Engine {
  private queries: Query[] = [];

  private readonly target = document.documentElement;

  constructor(rules: string) {
    console.log('here');
    this.parseRules(rules);
  }

  start() {
    this.applyQueries();

    if (document.readyState !== 'complete') {
      document.addEventListener('DOMContentLoaded', () => {
        this.applyQueries();
      });
    }

    this.registerObserver();
  }

  private parseRules(rules: string): void {
    const lines = rules.split('\n');

    for (let line of lines) {
      line = line.trim();
      if (line.length === 0) {
        continue;
      }

      const query = parse(line);
      this.queries.push(query);
    }
  }

  private applyQueries(): void {
    for (const query of this.queries) {
      let els: Element[] = [this.target];
      for (const step of query) {
        els = step.run(els);
        if (els.length === 0) {
          break;
        }
      }

      if (els.length > 0) {
        console.log('REMOVING ELEMENTS', els);
      }
      for (const el of els) {
        el.remove();
      }
    }
  }

  private registerObserver(): void {
    const options: MutationObserverInit = {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['id', 'class'],
    };

    const observer = new MutationObserver((mutations, observer) => {
      if (mutations.length === 0) {
        return;
      }
      if (mutations.every((m) => m.type === 'attributes')) {
        return;
      }

      // Avoid infinite looping
      observer.disconnect();

      this.applyQueries();

      observer.observe(this.target, options);
    });

    observer.observe(this.target, options);
  }
}
