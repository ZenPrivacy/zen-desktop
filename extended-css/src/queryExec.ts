import { Query } from './parse/types';

export class QueryRunner {
  constructor(private query: Query) {}

  run() {
    let els: Element[] = [document.documentElement];
    for (const step of this.query) {
      els = step.run(els);
      if (els.length === 0) {
        return;
      }
    }

    for (const el of els) {
      el.remove();
    }
  }
}
