import { Child } from './child';
import { Descendant } from './descendant';
import { NextSibling } from './nextSibling';
import { SubsequentSibling } from './subsequentSibling';

export type Combinator = Child | Descendant | NextSibling | SubsequentSibling;

export { Child, Descendant, NextSibling, SubsequentSibling };
