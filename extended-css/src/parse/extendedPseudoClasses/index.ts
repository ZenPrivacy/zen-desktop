import { Contains } from './contains';
import { MatchesCSS } from './matchesCSS';
import { MatchesPath } from './matchesPath';
import { Upward } from './upward';

export type ExtendedPseudoClass = Contains | MatchesCSS | MatchesPath | Upward;

export { Contains, MatchesCSS, MatchesPath, Upward };
