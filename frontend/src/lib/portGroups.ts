import type { DeviceModel } from '@/models';

/**
 * Returns a human-readable summary of a device model's port groups.
 * e.g. "48×25G + 6×100G"
 *
 * Returns "—" when no port groups are defined.
 * Falls back to "<count> ports" when port_count > 0 but no port_groups.
 */
export function portGroupSummary(dm: DeviceModel): string {
  const groups = dm.port_groups;
  if (!groups || groups.length === 0) {
    return '—';
  }
  return groups
    .map((g) => `${g.count}×${g.speed_gbps}G`)
    .join(' + ');
}
