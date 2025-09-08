import { parse } from '..';
import { SelectorExecutor } from '../../engine/selectorExecutor';
import { Step } from '../types';

export class Not implements Step {
  static requiresContext = true;

  private executor: SelectorExecutor;

  constructor(selector: string) {
    this.executor = new SelectorExecutor(parse(selector));
  }

  run(input: Element[]): Element[] {
    const matched: Set<Element> = new Set();

    const matchedEls = this.executor.match(document.documentElement);
    for (const element of matchedEls) {
      matched.add(element);
    }

    const filtered = [];
    for (const element of input) {
      if (!matched.has(element)) {
        filtered.push(element);
      }
    }
    return filtered;
  }
}
