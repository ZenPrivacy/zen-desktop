import { Step } from '../types';

export class SubsequentSibling implements Step {
  run(input: Element[]) {
    const result = new Set<Element>();

    for (const element of input) {
      const parent = element.parentElement;

      if (!parent) continue;

      let foundCurrent = false;

      for (const child of Array.from(parent.children)) {
        if (child === element) {
          foundCurrent = true;
          continue;
        }

        if (foundCurrent) {
          result.add(child);
        }
      }
    }

    return Array.from(result);
  }

  toString() {
    return 'SubsSiblComb';
  }
}
