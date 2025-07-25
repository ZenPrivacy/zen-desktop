import { parse } from './parse';
import { select } from './select';
import { Selector } from './types';

export default function (rules: string): void {
  const parsed: Selector[][] = [];
  for (const rule of rules.split('\n')) {
    try {
      parsed.push(parse(rule));
    } catch (err) {
      // eslint-disable-next-line no-console
      console.debug(`Zen (extended-css): failed to parse rule ${rule}: ${err}`);
    }
  }

  document.addEventListener('DOMContentLoaded', () => {
    const selectedSet = new Set<Element>();
    for (const selector of parsed) {
      const elements = select(selector);
      for (const el of elements) {
        selectedSet.add(el);
      }
    }

    console.log(selectedSet);
    selectedSet.forEach((el) => el.remove());
  });

  window.addEventListener('popstate', () => {
    const selectedSet = new Set<Element>();
    for (const selector of parsed) {
      const elements = select(selector);
      for (const el of elements) {
        selectedSet.add(el);
      }
    }

    console.log(selectedSet);
    selectedSet.forEach((el) => el.remove());
  });
}
