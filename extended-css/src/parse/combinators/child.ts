import { Step } from '../types';

export class Child implements Step {
  run(input: Element[]) {
    const res = [];
    for (const el of input) {
      res.push(...el.children);
    }
    return res;
  }
}
