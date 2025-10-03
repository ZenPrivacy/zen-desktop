import { describe, test } from '@jest/globals';

import { tokenize } from './tokenize';

describe('tokenize', () => {
  test.each<[string, string]>([
    ['div', 'RawTok(div)'],
    ['*', 'RawTok(*)'],
    ['a[href^="http"]', 'RawTok(a[href^="http"])'],
    ['div:not(.ad)', 'RawTok(div) ExtTok(:not(.ad))'],

    ['div>.x+span~a', 'RawTok(div) CombTok(>) RawTok(.x) CombTok(+) RawTok(span) CombTok(~) RawTok(a)'],

    ['div :not(.ad)', 'RawTok(div) CombTok( ) ExtTok(:not(.ad))'],

    ['div:contains(ad)', 'RawTok(div) ExtTok(:contains(ad))'],
    ['div.banner:matches-css(color: red)', 'RawTok(div.banner) ExtTok(:matches-css(color: red))'],
    [':matches-path(/^\\/shop/) .card', 'ExtTok(:matches-path(/^\\/shop/)) CombTok( ) RawTok(.card)'],
    ['div:upward(3)', 'RawTok(div) ExtTok(:upward(3))'],

    ['div:upward(3)~:contains(ad)', 'RawTok(div) ExtTok(:upward(3)) CombTok(~) ExtTok(:contains(ad))'],

    ['> .x:contains(y)', 'CombTok(>) RawTok(.x) ExtTok(:contains(y))'],

    ['div >', 'RawTok(div) CombTok(>)'],

    [':upward(1)+:upward(2)', 'ExtTok(:upward(1)) CombTok(+) ExtTok(:upward(2))'],

    // Selector lists in classes/pseudo-classes
    ['section:where(.x, .y)', 'RawTok(section:where(.x, .y))'],
    ['div:has(span, strong)', 'RawTok(div) ExtTok(:has(span, strong))'],

    ['div, .banner', 'RawTok(div), RawTok(.banner)'],
  ])('tokenize selector %j', (input, expected) => {
    const got = tokenize(input)
      .map((t) => t.map((t) => t.toString()).join(' '))
      .join(', ');
    expect(got).toEqual(expected);
  });
});
