import { Contains } from './contains';
import { Has } from './has';
import { MatchesCSS } from './matchesCSS';
import { MatchesPath } from './matchesPath';
import { Upward } from './upward';

export const extPseudoClasses = {
  contains: Contains,
  has: Has,
  'matches-css': MatchesCSS,
  'matches-path': MatchesPath,
  upward: Upward,
};
