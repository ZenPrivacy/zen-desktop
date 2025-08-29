import * as CSSTree from 'css-tree';

/**
 * Intermediate representation token.
 */
export type IRToken = RawToken | CombToken | ExtToken;

/**
 * Parses the selector into an intermediate token representation.
 */
export function tokenize(selector: string): IRToken[] {
  const ast = CSSTree.parse(selector, { context: 'selector', positions: true });

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
        if (ExtToken.classes.has(name)) {
          flushRaw();

          const arg = node.children?.first;
          if (arg == undefined) {
            throw new Error(`:${name}: expected an argument, got null/undefined`);
          }

          const argValue = getLiteral(arg);

          out.push(new ExtToken(name, argValue));
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
  /**
   * Names of supported extended pseudo classes.
   */
  static readonly classes = new Set(['contains', 'matches-css', 'matches-path', 'upward', 'has']);

  public kind: 'ext' = 'ext';
  constructor(
    public name: string,
    public args: string,
  ) {}

  get requiresContext() {
    switch (this.name) {
      case 'contains':
        return true;
      case 'matches-css':
        return true;
      case 'matches-path':
        return false;
      case 'upward':
        return true;
      case 'has':
        return true;
      default:
        return false;
    }
  }

  toString() {
    return `ExtTok(:${this.name}(${this.args}))`;
  }
}
