import { parse } from '..';
import { QueryRunner } from '../../engine/queryRunner';
import { Step } from '../types';

export class Has implements Step {
  private runners: QueryRunner[] = [];

  constructor(arg: string) {
    // arg may include multiple selectors in a selector list.
    this.runners = arg.split(',').map((selector) => {
      if (!selector.startsWith(':scope')) {
        selector = ':scope ' + selector;
      }
      return new QueryRunner(parse(selector));
    });
  }

  run(input: Element[]): Element[] {
    // For every element in "input", check if any runner returns at least a single result.
    return input.filter((element) => this.runners.some((runner) => runner.run([element]).length > 0));
  }
}
