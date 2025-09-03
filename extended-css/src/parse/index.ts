import { plan } from './plan';
import { tokenize } from './tokenize';
import { SelectorList } from './types';

export function parse(rule: string): SelectorList {
  const tokens = tokenize(rule);
  return tokens.map(plan);
}
