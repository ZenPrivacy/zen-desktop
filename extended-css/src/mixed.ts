import { Selector } from './types';

export class Mixed implements Selector {
  constructor(private query: string) {}

  select(input: Element[]) {
    return input.filter((el) => el.matches(this.query));
  }
}
