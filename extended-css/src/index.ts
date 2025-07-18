import * as CSSTree from 'css-tree';

export default function (rules: string) {
  const parsedRules = parse(rules);
}

function parse(rules: string): CSSTree.CssNode[] {
  return rules.split('\n').map((r) => CSSTree.parse(r, { context: 'selector', positions: true }));
}
