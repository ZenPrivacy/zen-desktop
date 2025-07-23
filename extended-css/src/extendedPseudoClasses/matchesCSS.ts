import { Selector } from '../types';
import { parseRegexpLiteral } from '../utils/parseRegexp';

export class MatchesCSS implements Selector {
  private pseudoElement?: string;
  private property: string;
  private valueRe?: RegExp;
  private valueSearch?: string;

  constructor(args: string) {
    const parsed = this.parseArgs(args);
    this.pseudoElement = parsed.pseudoElement;
    this.property = parsed.property;

    const re = parseRegexpLiteral(parsed.value);
    if (re !== null) {
      this.valueRe = re;
      return;
    }
    this.valueSearch = parsed.value;
  }

  private parseArgs(args: string): { pseudoElement?: string; property: string; value: string } {
    // Handle pseudo-element syntax: "before, property: value" or "property: value"
    const parts = args.split(',').map((s) => s.trim());

    let pseudoElement: string | undefined;
    let propertyValue: string;

    if (parts.length === 2) {
      // Has pseudo-element
      pseudoElement = parts[0];
      propertyValue = parts[1];
    } else {
      // No pseudo-element
      propertyValue = parts[0];
    }

    // Parse "property: value"
    const colonIndex = propertyValue.indexOf(':');
    if (colonIndex === -1) {
      throw new Error('Invalid matches-css syntax: missing colon separator');
    }

    const property = propertyValue.substring(0, colonIndex).trim();
    const value = propertyValue.substring(colonIndex + 1).trim();

    if (!property || !value) {
      throw new Error('Invalid matches-css syntax: empty property or value');
    }

    return { pseudoElement, property, value };
  }

  select(input: Element[]): Element[] {
    return input.filter((element) => this.matchesElement(element));
  }

  private matchesElement(element: Element): boolean {
    try {
      const computedStyle = window.getComputedStyle(element, this.pseudoElement);
      const actualValue = computedStyle.getPropertyValue(this.property);

      if (this.valueRe) {
        return this.valueRe.test(actualValue);
      } else if (this.valueSearch) {
        // Handle wildcard matching with case-insensitive comparison
        return this.matchesPattern(actualValue.toLowerCase(), this.valueSearch.toLowerCase());
      }

      return false;
    } catch (error) {
      // Handle cases where getComputedStyle fails
      return false;
    }
  }

  private matchesPattern(actualValue: string, pattern: string): boolean {
    // Simple wildcard matching - convert * to regex
    if (pattern.includes('*')) {
      const regexPattern = pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&').replace(/\\\*/g, '.*');
      const regex = new RegExp(`^${regexPattern}$`, 'i');
      return regex.test(actualValue);
    }

    // Exact match (case-insensitive)
    return actualValue === pattern;
  }
}
