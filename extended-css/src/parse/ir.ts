/**
 * Raw query token.
 */
export class RawToken {
  public kind: 'raw' = 'raw';
  constructor(public literal: string) {}
  toString() {
    return `Raw(${this.literal})`;
  }
}

/**
 * Combinator token.
 */
export class CombToken {
  public kind: 'comb' = 'comb';
  constructor(public literal: string) {}
  toString() {
    return `Comb(${this.literal})`;
  }
}

/**
 * Extended pseudo class token.
 */
export class ExtToken {
  public kind: 'ext' = 'ext';
  constructor(
    public name: string,
    public args: string,
    public requiresContext: boolean,
  ) {}
  toString() {
    return `Ext(:${this.name}(${this.args}))`;
  }
}

/**
 * Intermediate representation token.
 */
export type IRToken = RawToken | CombToken | ExtToken;
