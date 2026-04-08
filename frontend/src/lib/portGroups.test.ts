import { describe, it, expect } from 'vitest';
import { portGroupSummary } from './portGroups';
import type { DeviceModel } from '@/models';

function makeDevice(overrides: Partial<DeviceModel> = {}): DeviceModel {
  return {
    id: 1,
    vendor: 'Acme',
    model: 'Switch-1',
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
  };
}

describe('portGroupSummary', () => {
  it('returns "—" when port_groups is undefined', () => {
    const dm = makeDevice({ port_count: 0, port_groups: undefined });
    expect(portGroupSummary(dm)).toBe('—');
  });

  it('returns "—" when port_groups is an empty array', () => {
    const dm = makeDevice({ port_count: 48, port_groups: [] });
    expect(portGroupSummary(dm)).toBe('—');
  });

  it('formats a single port group correctly', () => {
    const dm = makeDevice({
      port_groups: [
        { id: 1, device_model_id: 1, count: 48, speed_gbps: 25, label: '', created_at: '' },
      ],
    });
    expect(portGroupSummary(dm)).toBe('48×25G');
  });

  it('formats multiple port groups joined by " + "', () => {
    const dm = makeDevice({
      port_groups: [
        { id: 1, device_model_id: 1, count: 48, speed_gbps: 25, label: '', created_at: '' },
        { id: 2, device_model_id: 1, count: 6, speed_gbps: 100, label: '', created_at: '' },
      ],
    });
    expect(portGroupSummary(dm)).toBe('48×25G + 6×100G');
  });

  it('handles mixed speeds with large counts', () => {
    const dm = makeDevice({
      port_groups: [
        { id: 1, device_model_id: 1, count: 128, speed_gbps: 400, label: '', created_at: '' },
        { id: 2, device_model_id: 1, count: 2, speed_gbps: 1000, label: '', created_at: '' },
      ],
    });
    expect(portGroupSummary(dm)).toBe('128×400G + 2×1000G');
  });

  it('formats three or more port groups correctly', () => {
    const dm = makeDevice({
      port_groups: [
        { id: 1, device_model_id: 1, count: 24, speed_gbps: 10, label: '', created_at: '' },
        { id: 2, device_model_id: 1, count: 24, speed_gbps: 25, label: '', created_at: '' },
        { id: 3, device_model_id: 1, count: 4, speed_gbps: 100, label: '', created_at: '' },
      ],
    });
    expect(portGroupSummary(dm)).toBe('24×10G + 24×25G + 4×100G');
  });
});
