import { describe, test, beforeAll, afterAll } from '@jest/globals';

import { Not } from './not';

describe(':not()', () => {
  let originalBody: string;

  beforeAll(() => {
    originalBody = document.body.innerHTML;

    document.body.innerHTML = `
        <div id="div1">
          <h2 id="h2">header</h2>
          <p id="p1">Text <span id="span">txet</span></p>
        </div>
        <div id="div2">
          <p id="p2"></p>
        </div>
      `;
  });

  afterAll(() => {
    document.body.innerHTML = originalBody;
  });

  test('matches descendants not matching a simple selector', () => {
    const selector = new Not('h2');
    const input = [document.querySelector('#div1')!];
    expect(selector.run(input)).toEqual([document.querySelector('#p1'), document.querySelector('#span')]);
  });

  test('supports selector list separated by commas', () => {
    const selector = new Not('h2, #span');
    const input = [document.querySelector('#div1')!];
    expect(selector.run(input)).toEqual([document.querySelector('#p1')]);
  });

  test('supports descendant selectors', () => {
    const selector = new Not('p span');
    const input = [document.querySelector('#div1')!];
    expect(selector.run(input)).toEqual([document.querySelector('#h2'), document.querySelector('#p1')]);
  });

  test('returns all descendants when none match the negated selector', () => {
    const selector = new Not('.does-not-exist');
    const input = [document.querySelector('#div1')!];
    expect(selector.run(input)).toEqual([
      document.querySelector('#h2'),
      document.querySelector('#p1'),
      document.querySelector('#span'),
    ]);
  });

  test('returns empty array if all descendants match (star selector)', () => {
    const selector = new Not('*');
    const input = [document.querySelector('#div1')!];
    expect(selector.run(input)).toEqual([]);
  });

  test('returns empty array for elements without descendants', () => {
    const selector = new Not('p');
    const input = [document.querySelector('#p2')!];
    expect(selector.run(input)).toEqual([]);
  });

  test('handles multiple input elements', () => {
    const selector = new Not('p');
    const input = [document.querySelector('#div1')!, document.querySelector('#div2')!];
    expect(selector.run(input)).toEqual([document.querySelector('#h2'), document.querySelector('#span')]);
  });
});
