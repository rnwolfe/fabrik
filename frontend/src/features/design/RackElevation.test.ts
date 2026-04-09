import { describe, it, expect } from 'vitest';
import type { RackWarning } from '@/models';
import { getPowerViolationText, getRuViolationText } from './RackElevation';

/**
 * Unit tests for inline rack violation text helpers in RackElevation.
 *
 * Covers:
 * - No violation (healthy rack): returns null
 * - Power violation: returns "+NNN W over limit"
 * - RU violation: returns "+NU over height"
 * - Missing / empty warnings: graceful null fallback (no crash)
 */

// ── Fixtures ──────────────────────────────────────────────────────────────────

const powerWarning: RackWarning = {
  kind: 'power',
  detail: 'power utilization at 110% of capacity (11000W typical / 10000W capacity)',
  delta_w: 1000,
};

const ruWarning: RackWarning = {
  kind: 'ru',
  detail: 'rack is over capacity: 46U used / 42U total (overflow by 4U)',
  delta_u: 4,
};

const portWarning: RackWarning = {
  kind: 'port',
  detail: 'leaf port exhaustion: 0 downlink ports available',
};

// ── getPowerViolationText ─────────────────────────────────────────────────────

describe('getPowerViolationText', () => {
  it('returns null when warnings is undefined', () => {
    expect(getPowerViolationText(undefined)).toBeNull();
  });

  it('returns null when warnings is an empty array', () => {
    expect(getPowerViolationText([])).toBeNull();
  });

  it('returns null when no power warning is present', () => {
    expect(getPowerViolationText([ruWarning, portWarning])).toBeNull();
  });

  it('returns null when power warning has no delta_w', () => {
    const w: RackWarning = { kind: 'power', detail: 'power warning', delta_w: undefined };
    expect(getPowerViolationText([w])).toBeNull();
  });

  it('returns null when power warning has delta_w of 0', () => {
    const w: RackWarning = { kind: 'power', detail: 'power warning', delta_w: 0 };
    expect(getPowerViolationText([w])).toBeNull();
  });

  it('returns "+1000 W over limit" for a 1000 W overage', () => {
    expect(getPowerViolationText([powerWarning])).toBe('+1000 W over limit');
  });

  it('returns correct text when power warning is mixed with other warnings', () => {
    expect(getPowerViolationText([ruWarning, powerWarning, portWarning])).toBe('+1000 W over limit');
  });

  it('returns violation text with exact delta_w value', () => {
    const w: RackWarning = { kind: 'power', detail: 'over', delta_w: 250 };
    expect(getPowerViolationText([w])).toBe('+250 W over limit');
  });
});

// ── getRuViolationText ────────────────────────────────────────────────────────

describe('getRuViolationText', () => {
  it('returns null when warnings is undefined', () => {
    expect(getRuViolationText(undefined)).toBeNull();
  });

  it('returns null when warnings is an empty array', () => {
    expect(getRuViolationText([])).toBeNull();
  });

  it('returns null when no RU warning is present', () => {
    expect(getRuViolationText([powerWarning, portWarning])).toBeNull();
  });

  it('returns null when RU warning has no delta_u', () => {
    const w: RackWarning = { kind: 'ru', detail: 'RU warning', delta_u: undefined };
    expect(getRuViolationText([w])).toBeNull();
  });

  it('returns null when RU warning has delta_u of 0', () => {
    const w: RackWarning = { kind: 'ru', detail: 'RU warning', delta_u: 0 };
    expect(getRuViolationText([w])).toBeNull();
  });

  it('returns "+4U over height" for a 4U overage', () => {
    expect(getRuViolationText([ruWarning])).toBe('+4U over height');
  });

  it('returns correct text when RU warning is mixed with other warnings', () => {
    expect(getRuViolationText([powerWarning, ruWarning, portWarning])).toBe('+4U over height');
  });

  it('returns violation text with exact delta_u value', () => {
    const w: RackWarning = { kind: 'ru', detail: 'over', delta_u: 2 };
    expect(getRuViolationText([w])).toBe('+2U over height');
  });
});
