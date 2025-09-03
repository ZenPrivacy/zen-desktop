import { Contains } from './contains';
import { Has } from './has';
import { MatchesCSS } from './matchesCSS';
import { MatchesPath } from './matchesPath';
import { Upward } from './upward';

/**
 * Maps pseudo-class names and aliases to their respective implementations.
 */
export const extPseudoClasses = {
  '-abp-contains': Contains,
  'has-text': Contains,
  contains: Contains,
  '-abp-has': Has,
  has: Has,
  'matches-css': MatchesCSS,
  'matches-path': MatchesPath,
  upward: Upward,
};
