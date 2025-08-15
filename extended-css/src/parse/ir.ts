export class RawSegment {
  public kind = 'raw';
  constructor(public text: string) {}
  toString() {
    return `Raw(${this.text})`;
  }
}

export class Combinator {
  public kind = 'comb';
  constructor(public token: string) {}
  toString() {
    return `Comb(${this.token})`;
  }
}

export class Extended {
  public kind = 'ext';
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
export type IRToken = RawSegment | Combinator | Extended;
