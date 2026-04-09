import { describe, it, expect } from 'vitest';
import type { RackWarning, RackSummary } from '@/models';

/**
 * Unit tests for rack warning badge rendering logic.
 *
 * These tests cover the data-side contract for RackWarningBadges:
 * - 0 warnings: nothing rendered
 * - 1 warning: single badge with correct kind label
 * - multiple warnings: one badge per violation
 * - legacy string fallback for backward compatibility
 */

// ── Pure helpers (mirror logic in RacksPage) ─────────────────────────────────

const warningKindLabel: Record<RackWarning['kind'], string> = {
  power: 'Power',
  ru: 'RU',
  port: 'Port',
};

function getWarningBadges(rack: Pick<RackSummary, 'warnings' | 'warning'>): Array<{ label: string; detail: string }> {
  if (rack.warnings && rack.warnings.length > 0) {
    return rack.warnings.map((w) => ({
      label: warningKindLabel[w.kind] ?? w.kind,
      detail: w.detail,
    }));
  }
  if (rack.warning) {
    return [{ label: 'warning', detail: rack.warning }];
  }
  return [];
}

// ── Fixtures ──────────────────────────────────────────────────────────────────

const powerWarning: RackWarning = {
  kind: 'power',
  detail: 'power utilization at 110% of capacity (11000W typical / 10000W capacity)',
  delta_w: 1000,
};

const ruWarning: RackWarning = {
  kind: 'ru',
  detail: 'rack is over capacity: 44U used / 42U total (overflow by 2U)',
  delta_u: 2,
};

const portWarning: RackWarning = {
  kind: 'port',
  detail: 'leaf port exhaustion: 0 downlink ports available',
};

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('getWarningBadges (rack warning badge data)', () => {
  it('returns empty array when rack has no warnings', () => {
    const badges = getWarningBadges({ warnings: [], warning: undefined });
    expect(badges).toHaveLength(0);
  });

  it('returns empty array when warnings is undefined and no legacy warning', () => {
    const badges = getWarningBadges({ warnings: undefined, warning: undefined });
    expect(badges).toHaveLength(0);
  });

  it('returns one badge for a single power warning', () => {
    const badges = getWarningBadges({ warnings: [powerWarning] });
    expect(badges).toHaveLength(1);
    expect(badges[0].label).toBe('Power');
    expect(badges[0].detail).toContain('110%');
  });

  it('returns one badge for a single RU warning', () => {
    const badges = getWarningBadges({ warnings: [ruWarning] });
    expect(badges).toHaveLength(1);
    expect(badges[0].label).toBe('RU');
    expect(badges[0].detail).toContain('overflow by 2U');
  });

  it('returns one badge for a single port warning', () => {
    const badges = getWarningBadges({ warnings: [portWarning] });
    expect(badges).toHaveLength(1);
    expect(badges[0].label).toBe('Port');
  });

  it('returns multiple badges for multiple violations', () => {
    const badges = getWarningBadges({ warnings: [powerWarning, ruWarning] });
    expect(badges).toHaveLength(2);
    expect(badges.map((b) => b.label)).toEqual(['Power', 'RU']);
  });

  it('returns badges for all three violation kinds', () => {
    const badges = getWarningBadges({ warnings: [powerWarning, ruWarning, portWarning] });
    expect(badges).toHaveLength(3);
    expect(badges.map((b) => b.label)).toEqual(['Power', 'RU', 'Port']);
  });

  it('falls back to legacy warning string when warnings is an empty array', () => {
    // When no structured warnings exist but a legacy string is present, show it.
    const badges = getWarningBadges({ warnings: [], warning: 'legacy warning text' });
    expect(badges).toHaveLength(1);
    expect(badges[0].label).toBe('warning');
    expect(badges[0].detail).toBe('legacy warning text');
  });

  it('falls back to legacy warning string when warnings is undefined', () => {
    const badges = getWarningBadges({ warnings: undefined, warning: 'legacy warning text' });
    expect(badges).toHaveLength(1);
    expect(badges[0].label).toBe('warning');
    expect(badges[0].detail).toBe('legacy warning text');
  });

  it('prefers structured warnings over legacy warning string', () => {
    const badges = getWarningBadges({
      warnings: [powerWarning],
      warning: 'legacy warning text',
    });
    expect(badges).toHaveLength(1);
    expect(badges[0].label).toBe('Power');
  });
});
