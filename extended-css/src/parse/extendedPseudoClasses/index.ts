import { Contains } from './contains';
import { Has } from './has';
import { MatchesAttr } from './matchesAttr';
import { MatchesCSS } from './matchesCSS';
import { MatchesPath } from './matchesPath';
import { MinTextLength } from './minTextLength';
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
  'matches-attr': MatchesAttr,
  'min-text-length': MinTextLength,
  upward: Upward,
};
