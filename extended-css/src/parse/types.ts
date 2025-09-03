/**
 * Part of a final {@link Selector}. Takes a set of elements as its input and returns another set based on internal semantics.
 */
export interface Step {
  run(input: Element[]): Element[];
}

/**
 * Parsed selector.
 */
export type Selector = Step[];

/**
 * Parsed selector list.
 */
export type SelectorList = Selector[];

export enum ParseContext {
  Selector = 'selector',
  SelectorList = 'selectorList',
}
