import * as CSSTree from 'css-tree';

import { Child } from './combinators/child';
import { Descendant } from './combinators/descendant';
import { MatchesPath } from './extendedPseudoClasses';
import { Mixed } from './mixed';
import { RawQuery } from './raw';
import { Selector } from './types';

export function select(selectors: Selector[]) {
  let nodes: Element[] = [document.body];
  for (const selector of selectors) {
    nodes = selector.select(nodes);
    console.log(selector, nodes);
    if (nodes.length === 0) {
      return nodes;
    }
  }
  return nodes;
}

export function parse(rule: string): Selector[] {
  const ast = CSSTree.parse(rule, { context: 'selector', positions: true });

  const res: Selector[] = [];

  CSSTree.walk(ast, (node) => {
    const literal = rule.slice(node.loc!.start.offset, node.loc!.end.offset);

    let selector: Selector;
    let action;
    switch (node.type) {
      case 'Selector':
        return;
      case 'IdSelector':
      case 'ClassSelector':
        if (res.length > 0) {
          selector = new Mixed(literal);
        } else {
          selector = new RawQuery(literal);
        }
        break;
      case 'Combinator':
        selector = parseCombinator(node.name);
        break;
      case 'PseudoClassSelector': {
        let skip;
        ({ selector, skip } = parsePseudoClass(node));
        if (skip) {
          action = CSSTree.walk.skip;
        }
        break;
      }
      default:
        throw new Error(`Unsupported node type ${node.type}`);
    }
    res.push(selector);
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
    default:
      throw new Error('Unsupported pseudoclass');
  }
}
