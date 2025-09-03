import * as CSSTree from 'css-tree';

import { extPseudoClasses } from './extendedPseudoClasses';

/**
 * Intermediate representation token.
 */
export type IRToken = RawToken | CombToken | ExtToken;

/**
 * Token representation of a selector.
 */
export type SelectorTokens = IRToken[];

/**
 * Parses a selector/selector list into an intermediate token representation.
 */
export function tokenize(selectorList: string): SelectorTokens[] {
  const ast = CSSTree.parse(selectorList, { context: 'selectorList', positions: true });

  const result: IRToken[][] = [];

  (ast as CSSTree.SelectorList).children.forEach((selectorNode) => {
    if (selectorNode.type === 'Selector') {
      result.push(parseTokens(selectorNode, selectorList));
    }
  });

  return result;
}

function parseTokens(ast: CSSTree.CssNode, selector: string): IRToken[] {
  const out: IRToken[] = [];
  let cssBuf = '';

  const flushRaw = () => {
    const t = cssBuf.trim();
    if (t.length > 0) {
      out.push(new RawToken(t));
    }
    cssBuf = '';
  };

  const getLiteral = (node: CSSTree.CssNode) => selector.slice(node.loc!.start.offset, node.loc!.end.offset);

  CSSTree.walk(ast, (node) => {
    switch (node.type) {
      case 'Selector':
        return;

      case 'IdSelector':
      case 'ClassSelector':
      case 'TypeSelector':
      case 'AttributeSelector':
        cssBuf += getLiteral(node);
        if (node.type === 'AttributeSelector') return CSSTree.walk.skip;
        return;

      case 'Combinator':
        flushRaw();
        out.push(new CombToken(node.name));
        return;

      case 'PseudoClassSelector': {
        const name = node.name.toLowerCase();
        if (name in extPseudoClasses) {
          flushRaw();

          const arg = node.children?.first;
          if (arg == undefined) {
            throw new Error(`:${name}: expected an argument, got null/undefined`);
          }

          const argValue = getLiteral(arg);

          out.push(new ExtToken(name as keyof typeof extPseudoClasses, argValue));
        } else {
          cssBuf += getLiteral(node);
        }
        return CSSTree.walk.skip;
      }

      default:
        throw new Error(`Unexpected node type: ${node.type}`);
    }
  });

  flushRaw();

  return out;
}

/**
 * Raw query token.
 */
export class RawToken {
  public kind: 'raw' = 'raw';
  constructor(public literal: string) {}
  toString() {
    return `RawTok(${this.literal})`;
  }
}

/**
 * Combinator token.
 */
export class CombToken {
  public kind: 'comb' = 'comb';
  constructor(public literal: string) {}
  toString() {
    return `CombTok(${this.literal})`;
  }
}

/**
 * Extended pseudo class token.
 */
export class ExtToken {
  public kind: 'ext' = 'ext';
  constructor(
    public name: keyof typeof extPseudoClasses,
    public args: string,
  ) {}

  toString() {
    return `ExtTok(:${this.name}(${this.args}))`;
  }
}
