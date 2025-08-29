/**
 * Part of a final Query. Takes a set of elements as its input and returns another set based on internal semantics.
 */
export interface Step {
  run(input: Element[]): Element[];
}

export type Query = Step[];
