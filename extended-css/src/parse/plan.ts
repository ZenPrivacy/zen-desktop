import { Child, Descendant, NextSibling, SubsequentSibling } from './combinators';
import { Contains, MatchesCSS, MatchesPath, Upward } from './extendedPseudoClasses';
import { RawQuery } from './raw';
import { IRToken, ExtToken, CombToken, Query } from './types';

/**
 * Builds a final, optimized query out of intermediate representation tokens.
 */
export function plan(tokens: IRToken[]): Query {
  const steps: Query = [];
  let cssBuilder = '';
  let havePrevStep = false;

  const flushRaw = () => {
    const raw = cssBuilder.trim();
    if (!raw) return;
    const prefix = havePrevStep ? ':scope ' : '';
    steps.push(new RawQuery(prefix + raw));
    cssBuilder = '';
    havePrevStep = true;
  };

  const emitBridge = (comb: CombToken) => {
    switch (comb.literal) {
      case ' ':
        steps.push(new Descendant());
        break;
      case '+':
        steps.push(new NextSibling());
        break;
      case '~':
        steps.push(new SubsequentSibling());
        break;
      case '>':
        steps.push(new Child());
        break;
    }
    havePrevStep = true;
  };

  for (let i = 0; i < tokens.length; i++) {
    const t = tokens[i];

    switch (t.kind) {
      case 'raw':
        cssBuilder += t.literal;
        break;
      case 'comb': {
        const next = tokens[i + 1];

        if (!next) {
          throw new Error('Last token is a dangling combinator');
        }

        switch (next.kind) {
          case 'raw':
            // Bridge declaratively
            cssBuilder += ` ${t.literal} `;
            break;
          case 'ext':
            // Bridge imperatively
            flushRaw();
            emitBridge(t);
            break;
          case 'comb':
            throw new Error('Multiple subsequent combinator tokens');
        }
        break;
      }
      case 'ext':
        flushRaw();
        if (t.requiresContext && !havePrevStep) {
          steps.push(new RawQuery('*'));
          havePrevStep = true;
        }
        steps.push(makeExtended(t));
        havePrevStep = true;
        break;
    }
  }

  flushRaw();
  return steps;
}

function makeExtended(token: ExtToken) {
  switch (token.name) {
    case 'matches-path':
      return new MatchesPath(token.args);
    case 'contains':
      return new Contains(token.args);
    case 'upward':
      return new Upward(token.args);
    case 'matches-css':
      return new MatchesCSS(token.args);
  }
  throw new Error(`Unknown extended pseudo class ${token.name}`);
}

/**
 * Determines whether a selector needs a scope, based on whether it has a leading combinator.
 */
// function needsScope(sel: string) {
//   return /^[ ~+>]/.test(sel);
// }
