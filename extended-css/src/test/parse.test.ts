import { describe, expect, test } from '@jest/globals';

import { parse, select } from '..';

describe('parse', () => {
  /**
   * @jest-environment jsdom
   * @jest-environment-options {"url": "https://example.com/test"}
   */

  test('test', () => {
    document.body.innerHTML = `
      <div id="id">
        <span class="class">
          yo
        </span>
      </div>
    `;
    // window.location.pathname = '/test';
    window.history.pushState({}, 'Test Title', '/test.html?query=true');
    console.log(window.location.pathname);

    const selectors = parse('#id > .class:matches-path(/test)');
    console.log(select(selectors));
  });
});
