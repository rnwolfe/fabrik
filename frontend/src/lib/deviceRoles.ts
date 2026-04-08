import type { DeviceModel } from '@/models';

export type Role = 'leaf' | 'spine' | 'super-spine';

/**
 * Suggest fabric tier roles for a device based on its port-group shape.
 *
 * Derivation rules (v1):
 * - Non-network device types → [] (servers/storage have no fabric role)
 * - No port groups → [] (cannot classify without speed info)
 * - Single port group (or all same speed), ≥32 ports total, all at ≥100 GbE
 *     → ['spine', 'super-spine']
 * - Single port group, ≥32 ports total, at ≥400 GbE
 *     → ['spine', 'super-spine'] (same rule; both roles apply at high radix)
 * - Mixed port groups: lower-speed downlink majority + smaller higher-speed uplink minority
 *     → ['leaf']
 * - A device whose port groups fit both shapes gets both roles.
 *
 * Edge case: 16×100G (borderline, < 32 ports threshold) → [] (no role assigned).
 * This is intentional: 16-port devices are too small to classify confidently.
 */
export function suggestedRoles(device: DeviceModel): Role[] {
  if (device.device_model_type !== 'network') return [];

  const groups = device.port_groups;
  if (!groups || groups.length === 0) return [];

  const roles: Role[] = [];

  if (isSpineCandidate(groups)) {
    roles.push('spine', 'super-spine');
  }

  if (isLeafCandidate(groups)) {
    roles.push('leaf');
  }

  return roles;
}

interface PortGroupLike {
  count: number;
  speed_gbps: number;
}

/**
 * Returns true if the device looks like a spine or super-spine:
 * - All port groups have the same speed (or only one group), and
 * - Total port count ≥ 32, and
 * - Port speed ≥ 100 GbE.
 */
function isSpineCandidate(groups: PortGroupLike[]): boolean {
  const speeds = new Set(groups.map((g) => g.speed_gbps));
  if (speeds.size > 1) return false; // mixed speeds → not a uniform spine

  const speed = groups[0].speed_gbps;
  if (speed < 100) return false;

  const totalPorts = groups.reduce((sum, g) => sum + g.count, 0);
  return totalPorts >= 32;
}

/**
 * Returns true if the device looks like a leaf:
 * - Has ≥ 2 port groups with different speeds, and
 * - The group(s) with the lowest speed are the majority (downlinks), and
 * - There is at least one smaller higher-speed group (uplinks).
 */
function isLeafCandidate(groups: PortGroupLike[]): boolean {
  if (groups.length < 2) return false;

  const speeds = Array.from(new Set(groups.map((g) => g.speed_gbps))).sort((a, b) => a - b);
  if (speeds.length < 2) return false; // all same speed — not a leaf shape

  const lowestSpeed = speeds[0];
  const highestSpeed = speeds[speeds.length - 1];

  const downlinkCount = groups
    .filter((g) => g.speed_gbps === lowestSpeed)
    .reduce((sum, g) => sum + g.count, 0);

  const uplinkCount = groups
    .filter((g) => g.speed_gbps === highestSpeed)
    .reduce((sum, g) => sum + g.count, 0);

  // Downlinks must outnumber uplinks (majority rule).
  return downlinkCount > uplinkCount;
}

/**
 * Human-readable label for a role badge.
 */
export const roleBadgeLabel: Record<Role, string> = {
  leaf: 'Leaf',
  spine: 'Spine',
  'super-spine': 'Super-spine',
};

/**
 * Tailwind CSS classes for each role badge variant.
 */
export const roleBadgeVariant: Record<Role, string> = {
  leaf: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
  spine: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  'super-spine': 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
};
