import { parse } from '..';
import { QueryRunner } from '../../engine/queryRunner';
import { Step } from '../types';

export class Not implements Step {
  static requiresContext = true;

  private runners: QueryRunner[] = [];

  constructor(arg: string) {
    // arg may include multiple selectors in a selector list.
    this.runners = arg.split(',').map((selector) => {
      return new QueryRunner(parse(selector));
    });
  }

  run(input: Element[]): Element[] {
    return input
      .map((element) => {
        const matched = new Set();
        for (const runner of this.runners) {
          runner.run([element]).forEach((el) => matched.add(el));
        }
        const notInMatched: Element[] = [];
        element.querySelectorAll('*').forEach((el) => {
          if (!matched.has(el)) notInMatched.push(el);
        });
        return notInMatched;
      })
      .flat();
  }
}
