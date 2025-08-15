import { Selector } from './types';

export class RawQuery implements Selector {
  constructor(private query: string) {}

  select(input: Element[]) {
    const res = [];
    for (const el of input) {
      const selected = el.querySelectorAll(this.query);
      for (const el of selected) {
        res.push(el);
      }
    }
    return res;
  }
}
