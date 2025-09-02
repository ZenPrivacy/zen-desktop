import { Step } from '../types';

/**
 * Imperative (JS) implementation of the subsequent-sibling (~) CSS combinator.
 *
 * @see {@link https://developer.mozilla.org/en-US/docs/Web/CSS/Subsequent-sibling_combinator}
 */
export class SubsequentSibling implements Step {
  run(input: Element[]) {
    const result = [];

    for (const element of input) {
      const parent = element.parentElement;

      if (!parent) continue;

      const children = Array.from(parent.children);
      for (let i = 0; i < children.length - 1; i++) {
        if (children[i] === element) {
          result.push(children[i + 1]);
          break;
        }
      }
    }

    return result;
  }

  toString() {
    return 'SubsSiblComb';
  }
}
