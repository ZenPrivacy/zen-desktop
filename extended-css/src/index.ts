import { parse } from './parse';
import { select } from './select';
import { Selector } from './types';

export default function (rules: string) {
  const parsed: Selector[][] = [];
  for (const rule of rules.split('\n')) {
    parsed.push(parse(rule));
  }

  document.addEventListener('DOMContentLoaded', () => {
    const selectedSet = new Set<Element>();
    for (const selector of parsed) {
      const elements = select(selector);
      for (const el of elements) {
        selectedSet.add(el);
      }
    }

    selectedSet.forEach((el) => el.remove());
  });
}
