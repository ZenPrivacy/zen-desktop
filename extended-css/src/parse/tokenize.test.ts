import { describe, test } from '@jest/globals';

import { tokenize } from './tokenize';

describe('tokenizeSelector → IR tokens', () => {
  test.each<[string, string]>([
    ['div', 'Raw(div)'],
    ['a[href^="http"]', 'Raw(a[href^="http"])'],
    ['div:not(.ad)', 'Raw(div:not(.ad))'],

    ['div>.x+span~a', 'Raw(div) Comb(>) Raw(.x) Comb(+) Raw(span) Comb(~) Raw(a)'],

    ['div :not(.ad)', 'Raw(div) Comb( ) Raw(:not(.ad))'],

    ['div:contains(ad)', 'Raw(div) Ext(:contains(ad))'],
    ['div.banner:matches-css(color: red)', 'Raw(div.banner) Ext(:matches-css(color: red))'],
    [':matches-path(/^\\/shop/) .card', 'Ext(:matches-path(/^\\/shop/)) Comb( ) Raw(.card)'],
    ['div:upward(3)', 'Raw(div) Ext(:upward(3))'],

    ['div:upward(3)~:contains(ad)', 'Raw(div) Ext(:upward(3)) Comb(~) Ext(:contains(ad))'],

    ['> .x:contains(y)', 'Comb(>) Raw(.x) Ext(:contains(y))'],

    ['div >', 'Raw(div) Comb(>)'],

    [':upward(1)+:upward(2)', 'Ext(:upward(1)) Comb(+) Ext(:upward(2))'],

    ['section:where(.x, .y)', 'Raw(section:where(.x, .y))'],
  ])('tokenize %j', (input, expected) => {
    const got = tokenize(input)
      .map((t) => t.toString())
      .join(' ');
    expect(got).toEqual(expected);
  });
});
