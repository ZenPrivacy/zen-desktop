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
