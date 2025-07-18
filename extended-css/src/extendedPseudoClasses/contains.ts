import { Selector } from '../types';
import { parseRegexpLiteral } from '../utils/parseRegexp';

export class Contains implements Selector {
  private textRe?: RegExp;
  private textSearch?: string;

  constructor(text: string) {
    const re = parseRegexpLiteral(text);
    if (re !== null) {
      this.textRe = re;
      return;
    }
    this.textSearch = text;
  }

  select(input: Element[]) {
    if (this.textRe) {
      return input.filter((e) => this.textRe!.test(e.innerHTML));
    } else if (this.textSearch) {
      return input.filter((e) => e.innerHTML.includes(this.textSearch!));
    }
    return [];
  }
}
