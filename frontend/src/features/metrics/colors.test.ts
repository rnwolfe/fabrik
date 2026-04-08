import { describe, it, expect } from 'vitest';
import {
  oversubColor,
  oversubBgColor,
  utilizationColor,
  utilizationBgColor,
} from './colors';

describe('oversubColor', () => {
  it('returns green for ratio <= 2', () => {
    expect(oversubColor(1)).toContain('green');
    expect(oversubColor(2)).toContain('green');
  });

  it('returns amber for ratio > 2 and <= 4', () => {
    expect(oversubColor(2.01)).toContain('amber');
    expect(oversubColor(3)).toContain('amber');
    expect(oversubColor(4)).toContain('amber');
  });

  it('returns red for ratio > 4', () => {
    expect(oversubColor(4.01)).toContain('red');
    expect(oversubColor(8)).toContain('red');
  });

  it('handles boundary at exactly 2', () => {
    expect(oversubColor(2)).toContain('green');
    expect(oversubColor(2.001)).toContain('amber');
  });

  it('handles boundary at exactly 4', () => {
    expect(oversubColor(4)).toContain('amber');
    expect(oversubColor(4.001)).toContain('red');
  });
});

describe('oversubBgColor', () => {
  it('returns green background for ratio <= 2', () => {
    expect(oversubBgColor(1)).toBe('bg-green-500');
    expect(oversubBgColor(2)).toBe('bg-green-500');
  });

  it('returns amber background for ratio > 2 and <= 4', () => {
    expect(oversubBgColor(3)).toBe('bg-amber-500');
    expect(oversubBgColor(4)).toBe('bg-amber-500');
  });

  it('returns red background for ratio > 4', () => {
    expect(oversubBgColor(5)).toBe('bg-red-500');
  });
});

describe('utilizationColor', () => {
  it('returns green for utilization <= 75%', () => {
    expect(utilizationColor(0)).toContain('green');
    expect(utilizationColor(50)).toContain('green');
    expect(utilizationColor(75)).toContain('green');
  });

  it('returns amber for utilization > 75% and <= 90%', () => {
    expect(utilizationColor(75.01)).toContain('amber');
    expect(utilizationColor(80)).toContain('amber');
    expect(utilizationColor(90)).toContain('amber');
  });

  it('returns red for utilization > 90%', () => {
    expect(utilizationColor(90.01)).toContain('red');
    expect(utilizationColor(100)).toContain('red');
  });

  it('handles boundary at exactly 75', () => {
    expect(utilizationColor(75)).toContain('green');
    expect(utilizationColor(75.001)).toContain('amber');
  });

  it('handles boundary at exactly 90', () => {
    expect(utilizationColor(90)).toContain('amber');
    expect(utilizationColor(90.001)).toContain('red');
  });
});

describe('utilizationBgColor', () => {
  it('returns green background for pct <= 75', () => {
    expect(utilizationBgColor(50)).toBe('bg-green-500');
    expect(utilizationBgColor(75)).toBe('bg-green-500');
  });

  it('returns amber background for pct > 75 and <= 90', () => {
    expect(utilizationBgColor(80)).toBe('bg-amber-500');
    expect(utilizationBgColor(90)).toBe('bg-amber-500');
  });

  it('returns red background for pct > 90', () => {
    expect(utilizationBgColor(95)).toBe('bg-red-500');
  });
});
