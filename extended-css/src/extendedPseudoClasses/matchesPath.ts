import { Selector } from '../types';
import { parseRegexpLiteral } from '../utils/parseRegexp';

export class MatchesPath implements Selector {
  private pathRe?: RegExp;
  private pathSearch?: string;

  constructor(path: string) {
    const re = parseRegexpLiteral(path);
    if (re !== null) {
      this.pathRe = re;
      return;
    }
    this.pathSearch = path;
  }

  select(input: Element[]) {
    const path = window.location.pathname;
    if (this.pathRe) {
      return this.pathRe.test(path) ? input : [];
    } else if (this.pathSearch) {
      return path.includes(this.pathSearch) ? input : [];
    }
    return [];
  }
}
