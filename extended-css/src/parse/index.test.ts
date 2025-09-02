import { describe, test, expect } from '@jest/globals';

import { parse } from './index';

describe('parse', () => {
  test.each<[string, string]>([
    ['div', 'RawQuery(div)'],
    ['a[href^="http"]', 'RawQuery(a[href^="http"])'],
    ['div:not(.ad)', 'RawQuery(div:not(.ad))'],

    // Pure CSS with combinators is bridged into a single Raw
    ['div>.x+span~a', 'RawQuery(div > .x + span ~ a)'],

    // Extended pseudo classes split into steps
    ['div:contains(ad)', 'RawQuery(div) :Contains(ad)'],
    ['div.banner:matches-css(color: red)', 'RawQuery(div.banner) :MatchesCSS(color: red)'],
    [':matches-path(/^\\/shop/) .card', ':MatchesPath(/^\\/shop/) RawQuery(.card)'],
    ['div:upward(3)', 'RawQuery(div) :Upward(3)'],

    // Imperative bridging when a combinator is adjacent to an extended step
    ['div:upward(3)~:contains(ad)', 'RawQuery(div) :Upward(3) SubsSiblComb :Contains(ad)'],

    // Leading combinator in raw followed by extended
    ['.x:contains(y)', 'RawQuery(.x) :Contains(y)'],

    // Context bootstrap for leading extended step and imperative bridge
    [':upward(1)+:upward(2)', 'RawQuery(*) :Upward(1) NextSiblComb :Upward(2)'],

    // Non-extended pseudo remains in raw
    ['section:where(.x, .y)', 'RawQuery(section:where(.x, .y))'],

    // Relative selector
    ['> .x + p', 'ChildComb RawMatches(.x) NextSiblComb RawMatches(p)'],
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
