import { describe, test, beforeEach, afterEach, expect } from '@jest/globals';

import { Engine } from '.';

describe('Engine', () => {
  let originalBody: string;

  beforeEach(() => {
    originalBody = document.body.innerHTML;
  });

  afterEach(() => {
    document.body.innerHTML = originalBody;
  });

  // Helper function to create test DOM structure
  const createTestDOM = (html: string) => {
    document.body.innerHTML = html;
  };

  // Helper function to check if element is hidden
  const isElementHidden = (element: Element): boolean => {
    const style = getComputedStyle(element);
    return style.display === 'none';
  };

  // Helper function to get visible elements by selector
  const getVisibleElements = (selector: string): Element[] => {
    return Array.from(document.querySelectorAll(selector)).filter((el) => !isElementHidden(el));
  };

  describe('basic selector parsing and execution', () => {
    test('hides elements matching simple class selector', () => {
      createTestDOM(`
        <div class="hide-me">Should be hidden</div>
        <div class="keep-me">Should remain visible</div>
        <span class="hide-me">Should also be hidden</span>
      `);

      const engine = new Engine('.hide-me');
      engine.start();

      expect(getVisibleElements('.hide-me')).toHaveLength(0);
      expect(getVisibleElements('.keep-me')).toHaveLength(1);
    });

    test('hides elements matching ID selector', () => {
      createTestDOM(`
        <div id="target">Should be hidden</div>
        <div id="other">Should remain visible</div>
      `);

      const engine = new Engine('#target');
      engine.start();

      expect(getVisibleElements('#target')).toHaveLength(0);
      expect(getVisibleElements('#other')).toHaveLength(1);
    });

    test('hides elements matching tag selector', () => {
      createTestDOM(`
        <span>Should be hidden</span>
        <div>Should remain visible</div>
        <span>Should also be hidden</span>
      `);

      const engine = new Engine('span');
      engine.start();

      expect(getVisibleElements('span')).toHaveLength(0);
      expect(getVisibleElements('div')).toHaveLength(1);
    });

    test('hides elements matching attribute selector', () => {
      createTestDOM(`
        <div data-ad="true">Should be hidden</div>
        <div data-content="true">Should remain visible</div>
        <span data-ad="banner">Should also be hidden</span>
      `);

      const engine = new Engine('[data-ad]');
      engine.start();

      expect(getVisibleElements('[data-ad]')).toHaveLength(0);
      expect(getVisibleElements('[data-content]')).toHaveLength(1);
    });
  });

  describe(':has() pseudo-class functionality', () => {
    test('hides parent elements containing specific children', () => {
      createTestDOM(`
        <div id="container1">
          <span class="ad-marker">Advertisement</span>
          <p>Some content</p>
        </div>
        <div id="container2">
          <p>Clean content</p>
        </div>
        <div id="container3">
          <div class="ad-marker">Another ad</div>
        </div>
      `);

      const engine = new Engine('div:has(.ad-marker)');
      engine.start();

      expect(getVisibleElements('#container1')).toHaveLength(0);
      expect(getVisibleElements('#container2')).toHaveLength(1);
      expect(getVisibleElements('#container3')).toHaveLength(0);
    });

    test('handles :has() with direct child combinator', () => {
      createTestDOM(`
        <div id="direct">
          <span class="marker">Direct child</span>
        </div>
        <div id="nested">
          <div>
            <span class="marker">Nested child</span>
          </div>
        </div>
      `);

      const engine = new Engine('div:has(> .marker)');
      engine.start();

      expect(getVisibleElements('#direct')).toHaveLength(0);
      expect(getVisibleElements('#nested')).toHaveLength(1);
    });

    test('handles :has() with selector list (OR semantics)', () => {
      createTestDOM(`
        <div id="hasSpan"><span>Has span</span></div>
        <div id="hasP"><p class="marker">Has p.marker</p></div>
        <div id="hasBoth">
          <span>Has span</span>
          <p class="marker">Has p.marker</p>
        </div>
        <div id="hasNeither">Has neither</div>
      `);

      const engine = new Engine('div:has(span, .marker)');
      engine.start();

      expect(getVisibleElements('#hasSpan')).toHaveLength(0);
      expect(getVisibleElements('#hasP')).toHaveLength(0);
      expect(getVisibleElements('#hasBoth')).toHaveLength(0);
      expect(getVisibleElements('#hasNeither')).toHaveLength(1);
    });
  });

  describe(':is() pseudo-class functionality', () => {
    test('hides elements matching any selector in list', () => {
      createTestDOM(`
        <div class="target">Should be hidden</div>
        <span id="special">Should be hidden</span>
        <p class="safe">Should remain visible</p>
        <div id="other">Should remain visible</div>
      `);

      const engine = new Engine(':is(.target, #special)');
      engine.start();

      expect(getVisibleElements('.target')).toHaveLength(0);
      expect(getVisibleElements('#special')).toHaveLength(0);
      expect(getVisibleElements('.safe')).toHaveLength(1);
      expect(getVisibleElements('#other')).toHaveLength(1);
    });

    test('handles :is() with complex selectors', () => {
      createTestDOM(`
        <div class="container">
          <span class="item first">Should be hidden</span>
          <span class="item">Should remain visible</span>
          <p class="item last">Should be hidden</p>
        </div>
      `);

      const engine = new Engine('.container :is(.first, .last)');
      engine.start();

      expect(getVisibleElements('.first')).toHaveLength(0);
      expect(getVisibleElements('.last')).toHaveLength(0);
      expect(getVisibleElements('.item:not(.first):not(.last)')).toHaveLength(1);
    });
  });

  describe('multiple rules and complex scenarios', () => {
    test('applies multiple rules independently', () => {
      createTestDOM(`
        <div class="ad">Ad content</div>
        <span class="tracker">Tracking pixel</span>
        <p class="content">Good content</p>
        <div class="popup">Popup</div>
      `);

      const rules = `
        .ad
        .tracker
        .popup
      `;

      const engine = new Engine(rules);
      engine.start();

      expect(getVisibleElements('.ad')).toHaveLength(0);
      expect(getVisibleElements('.tracker')).toHaveLength(0);
      expect(getVisibleElements('.popup')).toHaveLength(0);
      expect(getVisibleElements('.content')).toHaveLength(1);
    });

    test('handles nested :has() and :is() selectors', () => {
      createTestDOM(`
        <div id="complex1" class="container">
          <div class="ad-wrapper">
            <span class="ad">Advertisement</span>
          </div>
        </div>
        <div id="complex2" class="container">
          <div class="content-wrapper">
            <span class="content">Clean content</span>
          </div>
        </div>
      `);

      const engine = new Engine(':is(.container:has(.ad))');
      engine.start();

      expect(getVisibleElements('#complex1')).toHaveLength(0);
      expect(getVisibleElements('#complex2')).toHaveLength(1);
    });

    test('handles dynamic content updates', () => {
      jest.useFakeTimers();
      createTestDOM(`
        <div class="dangerous">Ad</div>
      `);

      const engine = new Engine('div:has-text(Ad)');
      engine.start();

      expect(getVisibleElements('div')).toHaveLength(0);

      for (let i = 0; i < 100; i++) {
        const div = document.createElement('div');
        div.textContent = 'Ad ' + i;
        document.documentElement.appendChild(div);
      }

      jest.runAllTimers();

      expect(getVisibleElements('.dangerous')).toHaveLength(0);
      jest.useRealTimers();
    });
  });

  describe('edge cases and error handling', () => {
    test('handles empty rules gracefully', () => {
      createTestDOM(`
        <div class="test">Should remain visible</div>
      `);

      const engine = new Engine('');
      expect(() => engine.start()).not.toThrow();
      expect(getVisibleElements('.test')).toHaveLength(1);
    });

    test('handles invalid CSS syntax gracefully', () => {
      createTestDOM(`
        <div class="test">Should remain visible</div>
      `);

      const engine = new Engine('~~~~invalid css syntax~~~~~');
      expect(() => engine.start()).not.toThrow();
      expect(getVisibleElements('.test')).toHaveLength(1);
    });

    test('handles malformed selectors in forgiving mode', () => {
      createTestDOM(`
        <div class="valid">Should be hidden</div>
        <div class="other">Should remain visible</div>
      `);

      const rules = `
        .valid
        :invalid-pseudo
      `;

      const engine = new Engine(rules);
      expect(() => engine.start()).not.toThrow();
      expect(getVisibleElements('.valid')).toHaveLength(0);
      expect(getVisibleElements('.other')).toHaveLength(1);
    });

    test('handles deeply nested structures', () => {
      const createNestedStructure = (depth: number): string => {
        if (depth === 0) return '<span class="deep-target">Deep content</span>';
        return `<div class="level-${depth}">${createNestedStructure(depth - 1)}</div>`;
      };

      createTestDOM(createNestedStructure(100));

      const engine = new Engine('div:has(.deep-target)');
      engine.start();

      expect(getVisibleElements('.level-1')).toHaveLength(0);
      expect(document.querySelectorAll('.deep-target')).toHaveLength(1);
    });
  });

  describe('performance and optimization', () => {
    test('handles large number of elements', () => {
      const elements = Array.from(
        { length: 10000 },
        (_, i) => `<div class="${i % 2 === 0 ? 'even' : 'odd'}" id="item-${i}">Item ${i}</div>`,
      ).join('');

      createTestDOM(elements);

      const engine = new Engine('.even:has-text(Item)');
      engine.start();

      expect(getVisibleElements('.even')).toHaveLength(0);
      expect(getVisibleElements('.odd')).toHaveLength(5000);
    });

    test('handles multiple engine instances independently', () => {
      createTestDOM(`
        <div class="target1">Target 1</div>
        <div class="target2">Target 2</div>
        <div class="safe">Safe content</div>
      `);

      const engine1 = new Engine('.target1');
      const engine2 = new Engine('.target2');

      engine1.start();
      expect(getVisibleElements('.target1')).toHaveLength(0);
      expect(getVisibleElements('.target2')).toHaveLength(1);

      engine2.start();
      expect(getVisibleElements('.target1')).toHaveLength(0);
      expect(getVisibleElements('.target2')).toHaveLength(0);
      expect(getVisibleElements('.safe')).toHaveLength(1);
    });
  });

  describe('real-world use cases', () => {
    test('blocks common ad patterns', () => {
      createTestDOM(`
        <div class="advertisement">Ad banner</div>
        <div data-ad-type="banner">Another ad</div>
        <div class="content">
          <div class="ad-container">
            <span class="ad-label">Sponsored</span>
            <div class="ad-content">Ad content</div>
          </div>
        </div>
        <article class="post">Clean content</article>
      `);

      const rules = `
        .advertisement
        [data-ad-type]
        div:has(.ad-label)
      `;

      const engine = new Engine(rules);
      engine.start();

      expect(getVisibleElements('.advertisement')).toHaveLength(0);
      expect(getVisibleElements('[data-ad-type]')).toHaveLength(0);
      expect(getVisibleElements('.ad-container')).toHaveLength(0);
      expect(getVisibleElements('.post')).toHaveLength(1);
    });

    test('handles social media widgets', () => {
      createTestDOM(`
        <div class="social-widget" data-platform="facebook">
          <iframe src="https://facebook.com/plugins/panopticon"></iframe>
        </div>
        <div class="social-widget" data-platform="twitter">
          <iframe src="https://twitter.com/plugins/panopticon"></iframe>
        </div>
        <div class="content">
          <p>Article content</p>
        </div>
      `);

      const engine = new Engine('.social-widget:has(iframe[src])');
      engine.start();

      expect(getVisibleElements('.social-widget')).toHaveLength(0);
      expect(getVisibleElements('.content')).toHaveLength(1);
    });
  });
});
