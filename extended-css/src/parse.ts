import * as CSSTree from 'css-tree';

import { Child } from './combinators/child';
import { Descendant } from './combinators/descendant';
import { Contains, MatchesPath } from './extendedPseudoClasses';
import { Upward } from './extendedPseudoClasses/upward';
import { Mixed } from './mixed';
import { RawQuery } from './raw';
import { Selector } from './types';

export function parse(rule: string): Selector[] {
  const ast = CSSTree.parse(rule, { context: 'selector', positions: true });

  const res: Selector[] = [];

  CSSTree.walk(ast, (node) => {
    let selector: Selector;
    let action;
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
          action = CSSTree.walk.skip;
        }
        break;
      }
      case 'Combinator':
        res.push(parseCombinator(node.name));
        break;
      case 'PseudoClassSelector': {
        let skip;
        ({ selector, skip } = parsePseudoClass(node));
        if (skip) {
          action = CSSTree.walk.skip;
        }
        res.push(selector);
        break;
      }
      default:
        throw new Error(`Unsupported node type ${node.type}`);
    }
    if (action) {
      return action;
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
    default:
      throw new Error('Combinator not supported');
  }
}

function parsePseudoClass(node: CSSTree.PseudoClassSelector): { selector: Selector; skip: boolean } {
  switch (node.name) {
    case 'matches-path': {
      const child = node.children?.first;
      if (child?.type !== 'Raw') {
        throw new Error('Bad matches-path');
      }
      return { selector: new MatchesPath(child.value), skip: true };
    }
    case 'contains': {
      const child = node.children?.first;
      if (child?.type !== 'Raw') {
        throw new Error('Bad contains');
      }
      return { selector: new Contains(child.value), skip: true };
    }
    case 'upward': {
      const child = node.children?.first;
      if (child?.type !== 'Raw') {
        throw new Error('Bad upward');
      }
      return { selector: new Upward(child.value), skip: true };
    }
    default:
      throw new Error('Unsupported pseudoclass');
  }
}
