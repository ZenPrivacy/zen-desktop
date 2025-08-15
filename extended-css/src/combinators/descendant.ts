import { Selector } from '../types';

export class Descendant implements Selector {
  select(input: Element[]) {
    const descendantSet = new Set<Element>();
    for (const el of input) {
      const descendants = el.querySelectorAll('*');
      for (const el of descendants) {
        descendantSet.add(el);
      }
    }
    return Array.from(descendantSet);
  }
}
