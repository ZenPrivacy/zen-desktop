import * as CSSTree from 'css-tree';

import { Child } from './combinators/child';
import { Descendant } from './combinators/descendant';
import { NextSibling } from './combinators/nextSibling';
import { SubsequentSibling } from './combinators/subsequentSibling';
import { Contains, MatchesPath } from './extendedPseudoClasses';
import { Upward } from './extendedPseudoClasses/upward';
import { Mixed } from './mixed';
import { RawQuery } from './raw';
import { Selector } from './types';

export function parse(rule: string): Selector[] {
  const ast = CSSTree.parse(rule, { context: 'selector', positions: true });

  const res: Selector[] = [];

  CSSTree.walk(ast, (node) => {
    switch (node.type) {
      case 'Selector':
        return;
      case 'IdSelector':
      case 'ClassSelector':
      case 'TypeSelector':
      case 'AttributeSelector': {
        const literal = rule.slice(node.loc!.start.offset, node.loc!.end.offset);
        if (res.length === 0) {
          res.push(new RawQuery(literal));
        } else if (res[res.length - 1] instanceof Descendant) {
          res[res.length - 1] = new RawQuery(literal);
        } else {
          res.push(new Mixed(literal));
        }

        if (node.type === 'AttributeSelector') {
          return CSSTree.walk.skip;
        }
        break;
      }
      case 'Combinator':
        res.push(parseCombinator(node.name));
        break;
      case 'PseudoClassSelector': {
        const { selector, skipChildren, requiresQueryInFront } = parsePseudoClass(node);
        if (requiresQueryInFront && res.length === 0) {
          res.push(new RawQuery('*'));
        }
        res.push(selector);
        if (skipChildren) {
          return CSSTree.walk.skip;
        }
        break;
      }
      default:
        throw new Error(`Unsupported node type ${node.type}`);
    }
  });

  return res;
}

function parseCombinator(literal: string): Selector {
  switch (literal) {
    case '>':
      return new Child();
    case ' ':
      return new Descendant();
    case '+':
      return new NextSibling();
    case '~':
      return new SubsequentSibling();
    default:
      throw new Error('Combinator not supported');
  }
}

function parsePseudoClass(node: CSSTree.PseudoClassSelector): {
  selector: Selector;
  skipChildren: boolean;
  requiresQueryInFront?: boolean;
} {
  switch (node.name) {
    case 'matches-path':
      return { selector: createPseudoWithArgument(node, MatchesPath, 'matches-path'), skipChildren: true };
    case 'contains':
      return {
        selector: createPseudoWithArgument(node, Contains, 'contains'),
        skipChildren: true,
        requiresQueryInFront: true,
      };
    case 'upward':
      return { selector: createPseudoWithArgument(node, Upward, 'upward'), skipChildren: true };
    case 'matches-css':
      return {
        selector: createPseudoWithArgument(node, Upward, 'matches-css'),
        skipChildren: true,
        requiresQueryInFront: true,
      };
    case 'has':
    case 'has-text':
    case 'matches-attr':
    case 'matches-css-before':
    case 'matches-css-after':
    case 'matches-media':
    case 'matches-prop':
    case 'min-text-length':
    case 'not':
    case 'others':
    case 'watch-attr':
    case 'xpath':
      throw new Error(`Unsupported pseudoclass ${node.name}`);
    default:
      throw new Error(`Unknown pseudoclass ${node.name}`);
  }
}

function createPseudoWithArgument<T extends Selector>(
  node: CSSTree.PseudoClassSelector,
  SelectorClass: new (value: string) => T,
  pseudoName: string,
) {
  const child = node.children?.first;
  if (child?.type !== 'Raw') {
    throw new Error(`Bad ${pseudoName}: expected first child type to be 'Raw', got ${child?.type}`);
  }
  return new SelectorClass(child.value);
}
