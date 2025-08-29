import { Query } from '../parse/types';

export class QueryRunner {
  constructor(private query: Query) {}

  run(input: Element[]): Element[] {
    let els: Element[] = input;
    for (const step of this.query) {
      els = step.run(els);
      if (els.length === 0) {
        return [];
      }
    }

    return els;
  }
}
