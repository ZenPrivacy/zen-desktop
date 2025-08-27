import { Step } from '../types';

export class NextSibling implements Step {
  run(input: Element[]) {
    const result = [];
    for (const element of input) {
      const nextSibling = element.nextElementSibling;
      if (nextSibling) {
        result.push(nextSibling);
      }
    }
    return result;
  }
}
