import { Selector } from '../types';

export class Child implements Selector {
  select(input: Element[]) {
    const res = [];
    for (const el of input) {
      res.push(...el.children);
    }
    return res;
  }
}
