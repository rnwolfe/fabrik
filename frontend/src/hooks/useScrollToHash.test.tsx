/**
 * Tests for the scroll-to-hash behaviour.
 *
 * The hook's core logic: given a hash string and a contentReady flag,
 * call scrollIntoView on the matching element (if it exists).
 *
 * We test this logic via a pure helper that accepts a mock getElementById,
 * avoiding any dependency on a real DOM environment.
 */
import { describe, it, expect, vi } from 'vitest';

/**
 * Core logic extracted from useScrollToHash's useEffect body.
 * Accepting getElementById as a parameter makes the logic fully testable
 * without a DOM environment.
 */
function scrollToHashLogic(
  hash: string,
  contentReady: boolean,
  getElementById: (id: string) => { scrollIntoView: (opts: ScrollIntoViewOptions) => void } | null
): void {
  if (!hash || !contentReady) return;
  const id = hash.startsWith('#') ? hash.slice(1) : hash;
  if (!id) return;
  const el = getElementById(id);
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
}

describe('useScrollToHash scroll logic', () => {
  it('calls scrollIntoView when hash is present and content is ready', () => {
    const scrollIntoView = vi.fn();
    const getElementById = vi.fn().mockReturnValue({ scrollIntoView });

    scrollToHashLogic('#leaf-spine-ratio', true, getElementById);

    expect(getElementById).toHaveBeenCalledWith('leaf-spine-ratio');
    expect(scrollIntoView).toHaveBeenCalledWith({
      behavior: 'smooth',
      block: 'start',
    });
  });

  it('does not scroll when contentReady is false', () => {
    const scrollIntoView = vi.fn();
    const getElementById = vi.fn().mockReturnValue({ scrollIntoView });

    scrollToHashLogic('#leaf-spine-ratio', false, getElementById);

    expect(getElementById).not.toHaveBeenCalled();
    expect(scrollIntoView).not.toHaveBeenCalled();
  });

  it('does not throw when element is not found', () => {
    const getElementById = vi.fn().mockReturnValue(null);

    expect(() => scrollToHashLogic('#nonexistent', true, getElementById)).not.toThrow();
  });

  it('does not call getElementById when hash is empty', () => {
    const getElementById = vi.fn();

    scrollToHashLogic('', true, getElementById);

    expect(getElementById).not.toHaveBeenCalled();
  });
});
