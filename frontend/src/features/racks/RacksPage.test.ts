import { describe, it, expect } from 'vitest';
import type { RackWarning } from '@/models';
import { getRuOverflowDelta } from './RacksPage';

/**
 * Unit tests for the getRuOverflowDelta helper exported from RacksPage.
 *
 * Covers:
 * - No violation (healthy rack): returns null
 * - RU violation with delta: returns the delta in U
 * - Missing / empty warnings: graceful null fallback (no crash)
 */

// ── Fixtures ──────────────────────────────────────────────────────────────────

const ruWarning: RackWarning = {
  kind: 'ru',
  detail: 'rack is over capacity: 46U used / 42U total (overflow by 4U)',
  delta_u: 4,
};

const powerWarning: RackWarning = {
  kind: 'power',
  detail: 'power over limit',
  delta_w: 1000,
};

// ── getRuOverflowDelta ────────────────────────────────────────────────────────

describe('getRuOverflowDelta', () => {
  it('returns null when warnings is undefined', () => {
    expect(getRuOverflowDelta({ used_u: 46, height_u: 42, warnings: undefined })).toBeNull();
  });

  it('returns null when warnings is an empty array', () => {
    expect(getRuOverflowDelta({ used_u: 46, height_u: 42, warnings: [] })).toBeNull();
  });

  it('returns null when no RU warning is present', () => {
    expect(getRuOverflowDelta({ used_u: 46, height_u: 42, warnings: [powerWarning] })).toBeNull();
  });

  it('returns null when RU warning has no delta_u', () => {
    const w: RackWarning = { kind: 'ru', detail: 'RU warning', delta_u: undefined };
    expect(getRuOverflowDelta({ used_u: 46, height_u: 42, warnings: [w] })).toBeNull();
  });

  it('returns null when RU warning has delta_u of 0', () => {
    const w: RackWarning = { kind: 'ru', detail: 'RU warning', delta_u: 0 };
    expect(getRuOverflowDelta({ used_u: 42, height_u: 42, warnings: [w] })).toBeNull();
  });

  it('returns 4 when rack has +4U RU overflow', () => {
    expect(getRuOverflowDelta({ used_u: 46, height_u: 42, warnings: [ruWarning] })).toBe(4);
  });

  it('returns the delta from the RU warning even when mixed with other warnings', () => {
    expect(getRuOverflowDelta({ used_u: 46, height_u: 42, warnings: [powerWarning, ruWarning] })).toBe(4);
  });

  it('returns the exact delta_u value', () => {
    const w: RackWarning = { kind: 'ru', detail: 'over', delta_u: 2 };
    expect(getRuOverflowDelta({ used_u: 44, height_u: 42, warnings: [w] })).toBe(2);
  });
});
