/**
 * Shared color helpers for metrics — used by MetricsPage and MetricsStrip.
 * Thresholds: ≤2:1 = excellent (green), ≤4:1 = acceptable (amber), >4:1 = high (red)
 */

export function oversubColor(ratio: number): string {
  if (ratio <= 2) return 'text-green-600 dark:text-green-400';
  if (ratio <= 4) return 'text-amber-600 dark:text-amber-400';
  return 'text-red-600 dark:text-red-400';
}

export function oversubBgColor(ratio: number): string {
  if (ratio <= 2) return 'bg-green-500';
  if (ratio <= 4) return 'bg-amber-500';
  return 'bg-red-500';
}

/**
 * Color for power / port utilization percentages.
 * >90% = red, >75% = amber, else default (green)
 */
export function utilizationColor(pct: number): string {
  if (pct > 90) return 'text-red-600 dark:text-red-400';
  if (pct > 75) return 'text-amber-600 dark:text-amber-400';
  return 'text-green-600 dark:text-green-400';
}

export function utilizationBgColor(pct: number): string {
  if (pct > 90) return 'bg-red-500';
  if (pct > 75) return 'bg-amber-500';
  return 'bg-green-500';
}
