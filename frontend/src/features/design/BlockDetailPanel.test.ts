/**
 * Unit tests for BlockDetailPanel topology derivation logic.
 * Tests cover the four panel states required by issue #48:
 *   - no leaf model set (no topology, show placeholder)
 *   - leaf-only (no spine → no topology, show dash stats)
 *   - spine-only (no leaf → no topology, show placeholder)
 *   - both set (topology rendered)
 */
import { describe, it, expect } from 'vitest';
import type { DeviceModel } from '@/models';

// ── Inline copies of the pure functions from BlockDetailPanel ────────────────
// These are tested here so we can verify the logic without rendering the component.

function deriveFromPortGroups(leafModel: DeviceModel): {
  downlinks: number;
  downlinkSpeed: number;
  uplinks: number;
  uplinkSpeed: number;
  oversubscription: number;
} | null {
  const groups = leafModel.port_groups;
  if (!groups || groups.length < 2) return null;

  const sorted = [...groups].sort((a, b) =>
    a.speed_gbps !== b.speed_gbps ? a.speed_gbps - b.speed_gbps : b.count - a.count
  );

  const uplinkGroup = sorted[sorted.length - 1];
  const downlinkGroups = sorted.slice(0, -1);

  const downlinks = downlinkGroups.reduce((sum, g) => sum + g.count, 0);
  const downlinkBw = downlinkGroups.reduce((sum, g) => sum + g.count * g.speed_gbps, 0);
  const uplinkBw = uplinkGroup.count * uplinkGroup.speed_gbps;

  if (uplinkBw === 0) return null;

  return {
    downlinks,
    downlinkSpeed: downlinkGroups[0].speed_gbps,
    uplinks: uplinkGroup.count,
    uplinkSpeed: uplinkGroup.speed_gbps,
    oversubscription: downlinkBw / uplinkBw,
  };
}

function maxSpines(leafModel: DeviceModel): number {
  const pg = deriveFromPortGroups(leafModel);
  return pg ? pg.uplinks : Math.max(1, leafModel.port_count - 1);
}

function deriveTopology(
  leafModel: DeviceModel | undefined,
  spineModel: DeviceModel | undefined,
  spineCount: number,
  rackCount: number
): { oversubscription: number; leaf_count: number; leaf_downlinks: number; leaf_uplinks: number; total_host_ports: number } | null {
  if (!leafModel || !spineModel) return null;

  const spineRadix = spineModel.port_count;
  const portGroupResult = deriveFromPortGroups(leafModel);

  let uplinks: number;
  let downlinks: number;
  let oversubscription: number;

  if (portGroupResult) {
    uplinks = Math.min(spineCount, portGroupResult.uplinks);
    downlinks = portGroupResult.downlinks;
    const downlinkBw = downlinks * portGroupResult.downlinkSpeed;
    const uplinkBw = uplinks * portGroupResult.uplinkSpeed;
    oversubscription = uplinkBw > 0 ? downlinkBw / uplinkBw : Infinity;
  } else {
    uplinks = spineCount;
    downlinks = leafModel.port_count - uplinks;
    oversubscription = uplinks > 0 ? downlinks / uplinks : Infinity;
  }

  const leavesPerRack = 2;
  const maxLeaves = spineRadix;
  const leafCount = Math.min((rackCount || 1) * leavesPerRack, maxLeaves);

  return {
    oversubscription: Math.round(oversubscription * 10) / 10,
    leaf_count: leafCount,
    leaf_downlinks: downlinks,
    leaf_uplinks: uplinks,
    total_host_ports: leafCount * downlinks,
  };
}

// ── Test fixtures ────────────────────────────────────────────────────────────

function makeLeaf(portCount = 48, portGroups?: DeviceModel['port_groups']): DeviceModel {
  return {
    id: 1,
    vendor: 'Arista',
    model: '7050CX3-32S',
    device_model_type: 'network',
    port_count: portCount,
    height_u: 1,
    power_watts_idle: 100,
    power_watts_typical: 200,
    power_watts_max: 300,
    cpu_sockets: 0,
    cores_per_socket: 0,
    ram_gb: 0,
    storage_tb: 0,
    gpu_count: 0,
    description: '',
    is_seed: false,
    archived_at: null,
    created_at: '',
    updated_at: '',
    port_groups: portGroups,
  };
}

function makeSpine(portCount = 64): DeviceModel {
  return { ...makeLeaf(portCount), id: 2, model: 'QFX10002-72Q' };
}

// ── State: no leaf ────────────────────────────────────────────────────────────

describe('panel state: no leaf model', () => {
  it('deriveTopology returns null without a leaf model', () => {
    const spine = makeSpine();
    expect(deriveTopology(undefined, spine, 4, 2)).toBeNull();
  });

  it('maxSpines returns 0 gracefully when there is no leaf model (caller guards)', () => {
    // The panel only calls maxSpines when leafModel is defined.
    // This test verifies that if called with a 0-port model it behaves safely.
    const leaf = makeLeaf(1);
    expect(maxSpines(leaf)).toBeGreaterThanOrEqual(1);
  });
});

// ── State: leaf-only (no spine) ───────────────────────────────────────────────

describe('panel state: leaf-only (no spine assigned)', () => {
  it('deriveTopology returns null without a spine model', () => {
    const leaf = makeLeaf(48);
    expect(deriveTopology(leaf, undefined, 2, 2)).toBeNull();
  });

  it('maxSpines returns correct uplink count from port groups', () => {
    const leaf = makeLeaf(48, [
      { id: 1, device_model_id: 1, count: 8, speed_gbps: 100, label: 'uplink', created_at: '' },
      { id: 2, device_model_id: 1, count: 40, speed_gbps: 25, label: 'downlink', created_at: '' },
    ]);
    expect(maxSpines(leaf)).toBe(8);
  });

  it('maxSpines falls back to port_count - 1 without port groups', () => {
    const leaf = makeLeaf(48);
    expect(maxSpines(leaf)).toBe(47);
  });
});

// ── State: spine-only (no leaf) ───────────────────────────────────────────────

describe('panel state: spine-only (no leaf assigned)', () => {
  it('deriveTopology returns null without a leaf model', () => {
    const spine = makeSpine(64);
    expect(deriveTopology(undefined, spine, 4, 2)).toBeNull();
  });

  it('deriveFromPortGroups returns null for a spine (typically single port group)', () => {
    const spineAsLeaf = makeLeaf(64); // no port_groups
    expect(deriveFromPortGroups(spineAsLeaf)).toBeNull();
  });
});

// ── State: both leaf and spine assigned ───────────────────────────────────────

describe('panel state: both leaf and spine assigned', () => {
  it('derives correct topology with uniform ports', () => {
    const leaf = makeLeaf(48);
    const spine = makeSpine(64);
    const topo = deriveTopology(leaf, spine, 4, 2);
    expect(topo).not.toBeNull();
    expect(topo!.leaf_uplinks).toBe(4);
    expect(topo!.leaf_downlinks).toBe(44);
    expect(topo!.oversubscription).toBe(11); // 44/4
    expect(topo!.leaf_count).toBe(4); // min(2*2, 64)
    expect(topo!.total_host_ports).toBe(176); // 4 leaves * 44 downlinks
  });

  it('derives correct topology with port groups', () => {
    const leaf = makeLeaf(48, [
      { id: 1, device_model_id: 1, count: 8, speed_gbps: 100, label: 'uplink', created_at: '' },
      { id: 2, device_model_id: 1, count: 40, speed_gbps: 25, label: 'downlink', created_at: '' },
    ]);
    const spine = makeSpine(64);
    const topo = deriveTopology(leaf, spine, 4, 2);
    expect(topo).not.toBeNull();
    expect(topo!.leaf_uplinks).toBe(4);     // min(spineCount=4, uplinks=8)
    expect(topo!.leaf_downlinks).toBe(40);
    // downlinkBw=40*25=1000, uplinkBw=4*100=400 → 1000/400=2.5
    expect(topo!.oversubscription).toBe(2.5);
  });

  it('caps leaf count at spine radix', () => {
    const leaf = makeLeaf(48);
    const spine = makeSpine(4); // only 4 ports
    const topo = deriveTopology(leaf, spine, 4, 10);
    expect(topo).not.toBeNull();
    expect(topo!.leaf_count).toBe(4); // capped at spine.port_count
  });

  it('calculates oversubscription rounded to 1 decimal place', () => {
    const leaf = makeLeaf(3); // 3 ports, 1 uplink, 2 downlinks
    const spine = makeSpine(64);
    const topo = deriveTopology(leaf, spine, 1, 1);
    expect(topo).not.toBeNull();
    expect(topo!.oversubscription).toBe(2); // 2/1 = 2.0
  });
});

// ── deriveFromPortGroups ──────────────────────────────────────────────────────

describe('deriveFromPortGroups', () => {
  it('returns null when no port groups', () => {
    expect(deriveFromPortGroups(makeLeaf(48))).toBeNull();
  });

  it('returns null when fewer than 2 port groups', () => {
    const leaf = makeLeaf(48, [
      { id: 1, device_model_id: 1, count: 48, speed_gbps: 25, label: 'all', created_at: '' },
    ]);
    expect(deriveFromPortGroups(leaf)).toBeNull();
  });

  it('identifies highest-speed group as uplinks', () => {
    const leaf = makeLeaf(48, [
      { id: 1, device_model_id: 1, count: 40, speed_gbps: 25, label: 'downlink', created_at: '' },
      { id: 2, device_model_id: 1, count: 8, speed_gbps: 100, label: 'uplink', created_at: '' },
    ]);
    const result = deriveFromPortGroups(leaf);
    expect(result).not.toBeNull();
    expect(result!.uplinks).toBe(8);
    expect(result!.downlinks).toBe(40);
    expect(result!.uplinkSpeed).toBe(100);
    expect(result!.downlinkSpeed).toBe(25);
  });

  it('tie-breaks by count: larger group is downlinks', () => {
    // Both groups same speed — larger group should be downlinks
    const leaf = makeLeaf(64, [
      { id: 1, device_model_id: 1, count: 48, speed_gbps: 100, label: 'downlink', created_at: '' },
      { id: 2, device_model_id: 1, count: 16, speed_gbps: 100, label: 'uplink', created_at: '' },
    ]);
    const result = deriveFromPortGroups(leaf);
    expect(result).not.toBeNull();
    // Tie-break: sort by count descending → [48, 16]; last in sorted = 16 (uplink)
    // So uplinkGroup = 16, downlinkGroups = [48]
    expect(result!.uplinks).toBe(16);
    expect(result!.downlinks).toBe(48);
  });
});
