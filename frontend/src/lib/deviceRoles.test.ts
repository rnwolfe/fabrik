import { describe, it, expect } from 'vitest';
import { suggestedRoles } from './deviceRoles';
import type { DeviceModel } from '@/models';

function pg(count: number, speed_gbps: number) {
  return { id: 1, device_model_id: 1, count, speed_gbps, label: '', created_at: '' };
}

function makeDevice(
  overrides: Partial<Omit<DeviceModel, 'port_groups'>> & {
    port_groups?: { count: number; speed_gbps: number }[];
  },
): DeviceModel {
  return {
    id: 1,
    vendor: 'Acme',
    model: 'Test',
    device_model_type: 'network',
    port_count: 0,
    height_u: 1,
    power_watts_idle: 0,
    power_watts_typical: 0,
    power_watts_max: 0,
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
    ...overrides,
    port_groups: (overrides.port_groups ?? []).map((p) => pg(p.count, p.speed_gbps)),
  };
}

describe('suggestedRoles', () => {
  it('returns spine and super-spine for a 32×400G switch', () => {
    const device = makeDevice({
      port_count: 32,
      port_groups: [{ count: 32, speed_gbps: 400 }],
    });
    const roles = suggestedRoles(device);
    expect(roles).toContain('spine');
    expect(roles).toContain('super-spine');
    expect(roles).not.toContain('leaf');
  });

  it('returns spine and super-spine for a 64×800G super-spine', () => {
    const device = makeDevice({
      port_count: 64,
      port_groups: [{ count: 64, speed_gbps: 800 }],
    });
    const roles = suggestedRoles(device);
    expect(roles).toContain('spine');
    expect(roles).toContain('super-spine');
    expect(roles).not.toContain('leaf');
  });

  it('returns leaf for a 48×25G + 6×100G switch', () => {
    const device = makeDevice({
      port_count: 54,
      port_groups: [
        { count: 48, speed_gbps: 25 },
        { count: 6, speed_gbps: 100 },
      ],
    });
    const roles = suggestedRoles(device);
    expect(roles).toContain('leaf');
    expect(roles).not.toContain('spine');
    expect(roles).not.toContain('super-spine');
  });

  it('returns no roles for a server', () => {
    const device = makeDevice({
      device_model_type: 'server',
      port_count: 2,
      port_groups: [{ count: 2, speed_gbps: 25 }],
    });
    expect(suggestedRoles(device)).toEqual([]);
  });

  it('returns no roles for a storage device', () => {
    const device = makeDevice({
      device_model_type: 'storage',
      port_count: 4,
      port_groups: [{ count: 4, speed_gbps: 100 }],
    });
    expect(suggestedRoles(device)).toEqual([]);
  });

  it('returns no roles for a 16×100G switch (borderline — below 32-port threshold)', () => {
    // Edge case: stable documented answer is [] because total ports < 32.
    const device = makeDevice({
      port_count: 16,
      port_groups: [{ count: 16, speed_gbps: 100 }],
    });
    expect(suggestedRoles(device)).toEqual([]);
  });

  it('returns no roles when port_groups is empty', () => {
    const device = makeDevice({ port_count: 48, port_groups: [] });
    expect(suggestedRoles(device)).toEqual([]);
  });

  it('returns no roles for a non-network type with matching port shape', () => {
    const device = makeDevice({
      device_model_type: 'other',
      port_count: 32,
      port_groups: [{ count: 32, speed_gbps: 400 }],
    });
    expect(suggestedRoles(device)).toEqual([]);
  });

  it('returns spine+super-spine for exactly 32 ports at 100G (boundary)', () => {
    const device = makeDevice({
      port_count: 32,
      port_groups: [{ count: 32, speed_gbps: 100 }],
    });
    const roles = suggestedRoles(device);
    expect(roles).toContain('spine');
    expect(roles).toContain('super-spine');
  });

  it('does not assign leaf role when downlink and uplink counts are equal', () => {
    // 8×100G + 8×400G — equal counts → not a majority downlink → not a leaf
    const device = makeDevice({
      port_count: 16,
      port_groups: [
        { count: 8, speed_gbps: 100 },
        { count: 8, speed_gbps: 400 },
      ],
    });
    const roles = suggestedRoles(device);
    expect(roles).not.toContain('leaf');
  });

  it('returns leaf for 36×100G + 4×400G pattern', () => {
    const device = makeDevice({
      port_count: 40,
      port_groups: [
        { count: 36, speed_gbps: 100 },
        { count: 4, speed_gbps: 400 },
      ],
    });
    const roles = suggestedRoles(device);
    expect(roles).toContain('leaf');
    expect(roles).not.toContain('spine');
    expect(roles).not.toContain('super-spine');
  });
});
