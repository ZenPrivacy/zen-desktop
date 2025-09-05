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
    return input
      .map((element) => {
        const matched = new Set();
        this.executor.match(element).forEach((el) => matched.add(el));

        const notInMatched: Element[] = [];
        element.querySelectorAll('*').forEach((el) => {
          if (!matched.has(el)) notInMatched.push(el);
        });
        return notInMatched;
      })
      .flat();
  }
}
