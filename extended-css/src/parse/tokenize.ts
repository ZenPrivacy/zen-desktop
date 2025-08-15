import * as CSSTree from 'css-tree';

import { CombToken, ExtToken, IRToken, RawToken } from './ir';

/**
 * Maps extended selector names to whether they require a context (a raw query) in front.
 */
const EXTENDED_CONTEXT: Record<string, boolean> = {
  contains: true,
  'matches-css': true,
  'matches-path': false,
  upward: true,
};

/**
 * Parses the selector into an intermediate token representation.
 */
export function tokenizeSelector(selector: string): IRToken[] {
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
        if (name in EXTENDED_CONTEXT) {
          flushRaw();

          const rawArg = node.children?.first;
          if (rawArg?.type !== 'Raw') {
            throw new Error(`:${name}(): expected Raw argument`);
          }

          out.push(new ExtToken(name, rawArg.value, EXTENDED_CONTEXT[name]));
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
