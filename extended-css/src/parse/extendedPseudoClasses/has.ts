import { parse } from '..';
import { SelectorExecutor } from '../../engine/selectorExecutor';
import { Step } from '../types';

export class Has implements Step {
  static requiresContext = true;

  private executor: SelectorExecutor;

  constructor(selector: string) {
    this.executor = new SelectorExecutor(parse(selector));
  }

  run(input: Element[]): Element[] {
    // For every element in "input", check if any runner returns at least a single result.
    return input.filter((element) => this.executor.match(element).length > 0);
  }

  toString() {
    // A complete implementation would store the "arg" passed to the
    // constructor. However, since it's unused in production methods, we avoid
    // the memory overhead at the cost of slightly less thorough testing.
    return ':Has(...)';
  }
}
