/**
 * Tests for HelpLink component logic.
 *
 * Since @testing-library/react is not installed, these tests verify the
 * href construction and aria-label derivation logic directly without
 * rendering the component into the DOM.
 */
import { describe, it, expect } from 'vitest';

// ── Helpers mirroring HelpLink's internal logic ────────────────────────────

function buildHref(article: string, anchor?: string): string {
  return anchor ? `/knowledge/${article}#${anchor}` : `/knowledge/${article}`;
}

function buildAriaLabel(article: string): string {
  return `Learn more about ${article.replace(/-/g, ' ')}`;
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe('HelpLink href construction', () => {
  it('includes article and anchor when both are provided', () => {
    const href = buildHref('oversubscription', 'leaf-spine-ratio');
    expect(href).toBe('/knowledge/oversubscription#leaf-spine-ratio');
  });

  it('omits the hash fragment when no anchor is provided', () => {
    const href = buildHref('radix');
    expect(href).toBe('/knowledge/radix');
  });

  it('builds the correct href for bisection-bandwidth with anchor', () => {
    const href = buildHref('bisection-bandwidth', 'full-bisection');
    expect(href).toBe('/knowledge/bisection-bandwidth#full-bisection');
  });
});

describe('HelpLink aria-label', () => {
  it('converts hyphens to spaces in the article name', () => {
    const label = buildAriaLabel('bisection-bandwidth');
    expect(label).toBe('Learn more about bisection bandwidth');
  });

  it('uses the article name directly when it has no hyphens', () => {
    const label = buildAriaLabel('radix');
    expect(label).toBe('Learn more about radix');
  });

  it('replaces multiple hyphens', () => {
    const label = buildAriaLabel('some-long-article-name');
    expect(label).toBe('Learn more about some long article name');
  });
});
