import { Child, Descendant, NextSibling, SubsequentSibling } from './combinators';
import { extPseudoClasses } from './extendedPseudoClasses';
import { RawMatches, RawQuery } from './raw';
import { CombToken, IRToken } from './tokenize';
import { Query } from './types';

/**
 * Builds a final, optimized query out of intermediate representation tokens.
 */
export function plan(tokens: IRToken[]): Query {
  if (tokens.length === 0) return [];

  const steps: Query = [];
  let haveContextualStep = false; // true after we've emitted anything that can serve as context for requiresContext

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
      default:
        throw new Error(`Unknown combinator: "${comb.literal}"`);
    }
    haveContextualStep = true;
  };

  const emitExt = (name: keyof typeof extPseudoClasses, args: string) => {
    const extClass = extPseudoClasses[name];
    if (!extClass) {
      throw new Error(`Unknown extended pseudo-class ":${name}"`);
    }
    if (extClass.requiresContext && !haveContextualStep) {
      steps.push(new RawQuery('*'));
      haveContextualStep = true;
    }
    steps.push(new extClass(args));
    haveContextualStep = true;
  };

  // ---------- Relative selector (starts with a combinator) ----------
  const startsWithCombinator = tokens[0].kind === 'comb';
  if (startsWithCombinator) {
    /**
     * Special case: if the selector start with a combinator, then it's a relative selector.
     * In this case we cannot optimize by merging adjacent "raw" runs; instead, we must rely on filtering via
     * .matches at each step.
     * This behavior is required to support the :has() pseudo-class.
     * See: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_selectors/Selector_structure#relative_selector
     */
    for (let i = 0; i < tokens.length; i++) {
      const t = tokens[i];

      switch (t.kind) {
        case 'comb': {
          // Mirror default path invariants
          const next = tokens[i + 1];
          if (!next) {
            throw new Error('Relative selector ends with a dangling combinator');
          }
          if (next.kind === 'comb') {
            throw new Error('Multiple subsequent combinator tokens in relative selector');
          }
          emitBridge(t);
          break;
        }
        case 'raw':
          // In relative mode we avoid merging raw chunks and rely on stepwise matching semantics.
          steps.push(new RawMatches(t.literal));
          haveContextualStep = true;
          break;
        case 'ext':
          emitExt(t.name, t.args);
          break;
      }
    }
    return steps;
  }

  // ---------- Default path (non-relative selector) ----------
  // Accumulates CSS for coalesced raw runs; trims on flush.
  let cssBuilder = '';
  const flushRaw = () => {
    const raw = cssBuilder.trim();
    if (raw) {
      steps.push(new RawQuery(raw));
      haveContextualStep = true;
    }
    cssBuilder = '';
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
        if (next.kind === 'comb') {
          throw new Error('Multiple subsequent combinator tokens');
        }

        if (next.kind === 'raw') {
          // If the next token is raw, prefer declarative bridging for performance.
          cssBuilder += ` ${t.literal} `;
        } else {
          // Next is ext: end the merged raw run and bridge imperatively.
          flushRaw();
          emitBridge(t);
        }
        break;
      }

      case 'ext':
        flushRaw();
        emitExt(t.name, t.args);
        break;
    }
  }

  flushRaw();
  return steps;
}
