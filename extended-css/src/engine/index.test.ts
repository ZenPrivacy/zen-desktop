import { describe, test, beforeEach, afterEach, expect } from '@jest/globals';

import { Engine } from '.';

describe('Engine', () => {
  let originalBody: string;
  const startedEngines: Engine[] = [];

  beforeEach(() => {
    originalBody = document.body.innerHTML;
  });

  afterEach(() => {
    // Stop all engines to disconnect MutationObservers and event handlers
    for (const engine of startedEngines) {
      engine.stop();
    }
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
    return Array.from(document.documentElement.querySelectorAll(selector)).filter((el) => !isElementHidden(el));
  };

  const startEngine = (rules: string): Engine => {
    const engine = new Engine(rules);
    engine.start();
    startedEngines.push(engine);
    return engine;
  };

  describe('basic selector parsing and execution', () => {
    test('hides elements matching simple class selector', () => {
      createTestDOM(`
        <div class="hide-me">Should be hidden</div>
        <div class="keep-me">Should remain visible</div>
        <span class="hide-me">Should also be hidden</span>
      `);

      startEngine('.hide-me');

      expect(getVisibleElements('.hide-me')).toHaveLength(0);
      expect(getVisibleElements('.keep-me')).toHaveLength(1);
    });

    test('hides elements matching ID selector', () => {
      createTestDOM(`
        <div id="target">Should be hidden</div>
        <div id="other">Should remain visible</div>
      `);

      startEngine('#target');

      expect(getVisibleElements('#target')).toHaveLength(0);
      expect(getVisibleElements('#other')).toHaveLength(1);
    });

    test('hides elements matching tag selector', () => {
      createTestDOM(`
        <span>Should be hidden</span>
        <div>Should remain visible</div>
        <span>Should also be hidden</span>
      `);

      startEngine('span');

      expect(getVisibleElements('span')).toHaveLength(0);
      expect(getVisibleElements('div')).toHaveLength(1);
    });

    test('hides elements matching attribute selector', () => {
      createTestDOM(`
        <div data-ad="true">Should be hidden</div>
        <div data-content="true">Should remain visible</div>
        <span data-ad="banner">Should also be hidden</span>
      `);

      startEngine('[data-ad]');

      expect(getVisibleElements('[data-ad]')).toHaveLength(0);
      expect(getVisibleElements('[data-content]')).toHaveLength(1);
    });

    test('hides all elements with a universal selector', () => {
      createTestDOM(`
        <div>Should be hidden</div>
        <span>Should also be hidden</span>
        <h3>Should also be hidden></h3>
        <p>
          <span>Should also be hidden</span>
          Should also be hidden
        </p>
      `);

      startEngine('*');

      expect(getVisibleElements('*')).toHaveLength(0);
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

      startEngine('div:has(.ad-marker)');

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

      startEngine('div:has(> .marker)');

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

      startEngine('div:has(span, .marker)');

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

      startEngine(':is(.target, #special)');

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

      startEngine('.container :is(.first, .last)');

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

      startEngine(rules);

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

      startEngine(':is(.container:has(.ad))');

      expect(getVisibleElements('#complex1')).toHaveLength(0);
      expect(getVisibleElements('#complex2')).toHaveLength(1);
    });

    test('handles dynamic content updates', () => {
      jest.useFakeTimers();
      createTestDOM(`
        <div class="dangerous">Ad</div>
      `);

      startEngine('div:has-text(Ad)');

      expect(getVisibleElements('div')).toHaveLength(0);

      for (let i = 0; i < 100; i++) {
        const div = document.createElement('div');
        div.textContent = 'Ad ' + i;
        document.body.appendChild(div);
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

      expect(() => startEngine('')).not.toThrow();
      expect(getVisibleElements('.test')).toHaveLength(1);
    });

    test('handles invalid CSS syntax gracefully', () => {
      createTestDOM(`
        <div class="test">Should remain visible</div>
      `);

      expect(() => new Engine('~~~~invalid css syntax~~~~~')).not.toThrow();
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

      expect(() => new Engine(rules).start()).not.toThrow();
      expect(getVisibleElements('.valid')).toHaveLength(0);
      expect(getVisibleElements('.other')).toHaveLength(1);
    });

    test('handles deeply nested structures', () => {
      const createNestedStructure = (depth: number): string => {
        if (depth === 0) return '<span class="deep-target">Deep content</span>';
        return `<div class="level-${depth}">${createNestedStructure(depth - 1)}</div>`;
      };

      createTestDOM(createNestedStructure(100));

      startEngine('div:has(.deep-target)');

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

      startEngine('.even:has-text(Item)');

      expect(getVisibleElements('.even')).toHaveLength(0);
      expect(getVisibleElements('.odd')).toHaveLength(5000);
    });

    test('handles multiple engine instances independently', () => {
      createTestDOM(`
        <div class="target1">Target 1</div>
        <div class="target2">Target 2</div>
        <div class="safe">Safe content</div>
      `);

      startEngine('.target1');

      expect(getVisibleElements('.target1')).toHaveLength(0);
      expect(getVisibleElements('.target2')).toHaveLength(1);

      startEngine('.target2');

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

      startEngine(rules);

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

      startEngine('.social-widget:has(iframe[src])');

      expect(getVisibleElements('.social-widget')).toHaveLength(0);
      expect(getVisibleElements('.content')).toHaveLength(1);
    });
  });
});
