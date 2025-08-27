import { plan } from './plan';
import { tokenize } from './tokenize';
import { Query } from './types';

export function parse(rule: string): Query {
  const tokens = tokenize(rule);
  const query = plan(tokens);
  return query;
}
