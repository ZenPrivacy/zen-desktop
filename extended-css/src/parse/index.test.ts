import { describe, test, expect } from '@jest/globals';

import { parse } from './index';

describe('parse', () => {
  test.each<[string, string]>([
    ['div', 'Raw(div)'],
    ['a[href^="http"]', 'Raw(a[href^="http"])'],
    ['div:not(.ad)', 'Raw(div:not(.ad))'],

    // Pure CSS with combinators is bridged into a single Raw
    ['div>.x+span~a', 'Raw(div > .x + span ~ a)'],

    // Extended pseudo classes split into steps
    ['div:contains(ad)', 'Raw(div) :Contains(ad)'],
    ['div.banner:matches-css(color: red)', 'Raw(div.banner) :MatchesCSS(color: red)'],
    [':matches-path(/^\\/shop/) .card', ':MatchesPath(/^\\/shop/) Raw(:scope .card)'],
    ['div:upward(3)', 'Raw(div) :Upward(3)'],

    // Imperative bridging when a combinator is adjacent to an extended step
    ['div:upward(3)~:contains(ad)', 'Raw(div) :Upward(3) SubsSiblComb :Contains(ad)'],

    // Leading combinator in raw followed by extended
    ['.x:contains(y)', 'Raw(.x) :Contains(y)'],

    // Context bootstrap for leading extended step and imperative bridge
    [':upward(1)+:upward(2)', 'Raw(*) :Upward(1) NextSiblComb :Upward(2)'],

    // Non-extended pseudo remains in raw
    ['section:where(.x, .y)', 'Raw(section:where(.x, .y))'],

    // Imperative bridging between raw and extended
    ['.x > :contains(y)', 'Raw(.x) ChildComb :Contains(y)'],
  ])('parse %j', (input, expected) => {
    const got = parse(input)
      .map((s) => s.toString())
      .join(' ');
    expect(got).toEqual(expected);
  });

  test('throws on dangling combinator', () => {
    expect(() => parse('div >')).toThrow(/dangling combinator/i);
  });
});
