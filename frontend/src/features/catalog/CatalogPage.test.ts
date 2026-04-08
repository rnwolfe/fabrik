import { describe, it, expect } from 'vitest';
import type { DeviceModel } from '@/models';
import {
  parseSpeedsParam,
  collectSpeedOptions,
  deviceMatchesSpeed,
} from './CatalogPage';

// Minimal DeviceModel fixture
const makeDevice = (overrides: Partial<DeviceModel> = {}): DeviceModel => ({
  id: 1,
  vendor: 'Cisco',
  model: 'Nexus 9336C',
  device_model_type: 'network',
  port_count: 36,
  height_u: 1,
  power_watts_idle: 200,
  power_watts_typical: 350,
  power_watts_max: 500,
  cpu_sockets: 0,
  cores_per_socket: 0,
  ram_gb: 0,
  storage_tb: 0,
  gpu_count: 0,
  description: 'Test device',
  is_seed: false,
  archived_at: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  port_groups: [
    { id: 1, device_model_id: 1, count: 28, speed_gbps: 25, label: '25G SFP28', created_at: '' },
    { id: 2, device_model_id: 1, count: 8, speed_gbps: 100, label: '100G QSFP28', created_at: '' },
  ],
  ...overrides,
});

describe('parseSpeedsParam', () => {
  it('returns empty set for null', () => {
    expect(parseSpeedsParam(null).size).toBe(0);
  });

  it('returns empty set for empty string', () => {
    expect(parseSpeedsParam('').size).toBe(0);
  });

  it('parses a single speed', () => {
    const result = parseSpeedsParam('100');
    expect(result.size).toBe(1);
    expect(result.has(100)).toBe(true);
  });

  it('parses multiple comma-separated speeds', () => {
    const result = parseSpeedsParam('25,100,400');
    expect(result.size).toBe(3);
    expect(result.has(25)).toBe(true);
    expect(result.has(100)).toBe(true);
    expect(result.has(400)).toBe(true);
  });

  it('ignores invalid (NaN) values', () => {
    const result = parseSpeedsParam('25,abc,100');
    expect(result.size).toBe(2);
    expect(result.has(25)).toBe(true);
    expect(result.has(100)).toBe(true);
  });

  it('ignores zero and negative values', () => {
    const result = parseSpeedsParam('0,-10,100');
    expect(result.size).toBe(1);
    expect(result.has(100)).toBe(true);
  });
});

describe('collectSpeedOptions', () => {
  it('returns empty array for no devices', () => {
    expect(collectSpeedOptions([])).toEqual([]);
  });

  it('collects distinct speeds sorted ascending', () => {
    const devices = [
      makeDevice({
        id: 1,
        port_groups: [
          { id: 1, device_model_id: 1, count: 32, speed_gbps: 400, label: '400G', created_at: '' },
          { id: 2, device_model_id: 1, count: 8, speed_gbps: 100, label: '100G', created_at: '' },
        ],
      }),
      makeDevice({
        id: 2,
        port_groups: [
          { id: 3, device_model_id: 2, count: 48, speed_gbps: 25, label: '25G', created_at: '' },
          { id: 4, device_model_id: 2, count: 8, speed_gbps: 100, label: '100G', created_at: '' },
        ],
      }),
    ];
    expect(collectSpeedOptions(devices)).toEqual([25, 100, 400]);
  });

  it('handles devices with no port_groups', () => {
    const devices = [
      makeDevice({ port_groups: undefined }),
      makeDevice({
        id: 2,
        port_groups: [
          { id: 1, device_model_id: 2, count: 48, speed_gbps: 25, label: '25G', created_at: '' },
        ],
      }),
    ];
    expect(collectSpeedOptions(devices)).toEqual([25]);
  });

  it('deduplicates speeds across devices', () => {
    const devices = [
      makeDevice({
        id: 1,
        port_groups: [{ id: 1, device_model_id: 1, count: 48, speed_gbps: 25, label: '25G', created_at: '' }],
      }),
      makeDevice({
        id: 2,
        port_groups: [{ id: 2, device_model_id: 2, count: 48, speed_gbps: 25, label: '25G', created_at: '' }],
      }),
    ];
    expect(collectSpeedOptions(devices)).toEqual([25]);
  });
});

describe('deviceMatchesSpeed', () => {
  it('matches any device when selectedSpeeds is empty', () => {
    const device = makeDevice();
    expect(deviceMatchesSpeed(device, new Set())).toBe(true);
  });

  it('matches device when one of its port groups has the selected speed', () => {
    const device = makeDevice(); // has 25G and 100G port groups
    expect(deviceMatchesSpeed(device, new Set([25]))).toBe(true);
  });

  it('matches device on the second port group speed', () => {
    const device = makeDevice(); // has 25G and 100G port groups
    expect(deviceMatchesSpeed(device, new Set([100]))).toBe(true);
  });

  it('matches device when multiple selected speeds are active (OR logic)', () => {
    const device = makeDevice(); // has 25G and 100G port groups
    expect(deviceMatchesSpeed(device, new Set([25, 400]))).toBe(true);
  });

  it('does not match device when no port group has the selected speed', () => {
    const device = makeDevice(); // has 25G and 100G port groups
    expect(deviceMatchesSpeed(device, new Set([400]))).toBe(false);
  });

  it('does not match device with no port_groups when speed filter is active', () => {
    const device = makeDevice({ port_groups: undefined });
    expect(deviceMatchesSpeed(device, new Set([25]))).toBe(false);
  });

  it('does not match device with empty port_groups when speed filter is active', () => {
    const device = makeDevice({ port_groups: [] });
    expect(deviceMatchesSpeed(device, new Set([25]))).toBe(false);
  });
});

describe('speed filter integration', () => {
  const cisco = makeDevice({
    id: 1,
    vendor: 'Cisco',
    model: 'Nexus 93180YC',
    port_groups: [
      { id: 1, device_model_id: 1, count: 48, speed_gbps: 25, label: '25G', created_at: '' },
      { id: 2, device_model_id: 1, count: 6, speed_gbps: 100, label: '100G', created_at: '' },
    ],
  });
  const arista = makeDevice({
    id: 2,
    vendor: 'Arista',
    model: '7050CX3-32S',
    port_groups: [
      { id: 3, device_model_id: 2, count: 32, speed_gbps: 100, label: '100G', created_at: '' },
    ],
  });
  const juniper = makeDevice({
    id: 3,
    vendor: 'Juniper',
    model: 'QFX5220',
    port_groups: [
      { id: 4, device_model_id: 3, count: 32, speed_gbps: 400, label: '400G', created_at: '' },
    ],
  });

  const devices = [cisco, arista, juniper];

  it('filters to only devices with 25G ports', () => {
    const speeds = new Set([25]);
    const result = devices.filter((d) => deviceMatchesSpeed(d, speeds));
    expect(result).toHaveLength(1);
    expect(result[0].vendor).toBe('Cisco');
  });

  it('filters to devices with 100G ports', () => {
    const speeds = new Set([100]);
    const result = devices.filter((d) => deviceMatchesSpeed(d, speeds));
    expect(result).toHaveLength(2);
    expect(result.map((d) => d.vendor)).toContain('Cisco');
    expect(result.map((d) => d.vendor)).toContain('Arista');
  });

  it('filters with multiple selected speeds uses OR logic', () => {
    const speeds = new Set([25, 400]);
    const result = devices.filter((d) => deviceMatchesSpeed(d, speeds));
    expect(result).toHaveLength(2);
    expect(result.map((d) => d.vendor)).toContain('Cisco');
    expect(result.map((d) => d.vendor)).toContain('Juniper');
  });

  it('combines speed filter with vendor filter (AND semantics)', () => {
    const speeds = new Set([100]);
    const vendorFilter: string = 'Arista';
    const result = devices.filter(
      (d) => (vendorFilter === 'all' || d.vendor === vendorFilter) && deviceMatchesSpeed(d, speeds)
    );
    expect(result).toHaveLength(1);
    expect(result[0].vendor).toBe('Arista');
  });

  it('URL roundtrip: parseSpeedsParam reconstructs the selected set', () => {
    const original = new Set([25, 100, 400]);
    const param = Array.from(original).sort((a, b) => a - b).join(',');
    const restored = parseSpeedsParam(param);
    expect(restored).toEqual(original);
  });
});
