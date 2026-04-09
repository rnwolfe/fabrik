import { describe, it, expect } from 'vitest';
import { deviceFormSchema, deviceToFormValues, defaultDeviceFormValues } from './DeviceForm';
import type { DeviceModel } from '@/models';

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

describe('deviceFormSchema', () => {
  it('accepts valid network device data', () => {
    const result = deviceFormSchema.safeParse({
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
      description: '',
      port_groups: [],
    });
    expect(result.success).toBe(true);
  });

  it('rejects missing vendor', () => {
    const result = deviceFormSchema.safeParse({
      vendor: '',
      model: 'Nexus 9336C',
      device_model_type: 'network',
      port_count: 36,
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
      port_groups: [],
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const vendorError = result.error.issues.find((i) => i.path[0] === 'vendor');
      expect(vendorError).toBeDefined();
    }
  });

  it('rejects missing model', () => {
    const result = deviceFormSchema.safeParse({
      vendor: 'Cisco',
      model: '',
      device_model_type: 'network',
      port_count: 36,
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
      port_groups: [],
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const modelError = result.error.issues.find((i) => i.path[0] === 'model');
      expect(modelError).toBeDefined();
    }
  });

  it('coerces empty string port_count to 0', () => {
    const result = deviceFormSchema.safeParse({
      vendor: 'Cisco',
      model: 'Nexus 9336C',
      device_model_type: 'network',
      port_count: '',
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
      port_groups: [],
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.port_count).toBe(0);
    }
  });

  it('accepts server device type', () => {
    const result = deviceFormSchema.safeParse({
      vendor: 'Dell',
      model: 'PowerEdge R750',
      device_model_type: 'server',
      port_count: 4,
      height_u: 2,
      power_watts_idle: 100,
      power_watts_typical: 400,
      power_watts_max: 800,
      cpu_sockets: 2,
      cores_per_socket: 16,
      ram_gb: 256,
      storage_tb: 10,
      gpu_count: 0,
      description: '',
      port_groups: [],
    });
    expect(result.success).toBe(true);
  });

  it('rejects invalid device_model_type', () => {
    const result = deviceFormSchema.safeParse({
      vendor: 'Cisco',
      model: 'Nexus',
      device_model_type: 'router',
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
      port_groups: [],
    });
    expect(result.success).toBe(false);
  });
});

describe('deviceToFormValues', () => {
  it('maps DeviceModel to form values correctly', () => {
    const device = makeDevice();
    const values = deviceToFormValues(device);

    expect(values.vendor).toBe('Cisco');
    expect(values.model).toBe('Nexus 9336C');
    expect(values.device_model_type).toBe('network');
    expect(values.port_count).toBe(36);
    expect(values.height_u).toBe(1);
    expect(values.power_watts_idle).toBe(200);
    expect(values.power_watts_typical).toBe(350);
    expect(values.power_watts_max).toBe(500);
    expect(values.description).toBe('Test device');
  });

  it('maps port_groups correctly', () => {
    const device = makeDevice();
    const values = deviceToFormValues(device);

    expect(values.port_groups).toHaveLength(2);
    expect(values.port_groups[0]).toEqual({ count: 28, speed_gbps: 25, label: '25G SFP28' });
    expect(values.port_groups[1]).toEqual({ count: 8, speed_gbps: 100, label: '100G QSFP28' });
  });

  it('handles missing port_groups gracefully', () => {
    const device = makeDevice({ port_groups: undefined });
    const values = deviceToFormValues(device);
    expect(values.port_groups).toEqual([]);
  });
});

describe('defaultDeviceFormValues', () => {
  it('has sensible defaults', () => {
    expect(defaultDeviceFormValues.device_model_type).toBe('network');
    expect(defaultDeviceFormValues.height_u).toBe(1);
    expect(defaultDeviceFormValues.port_count).toBe(0);
    expect(defaultDeviceFormValues.port_groups).toEqual([]);
  });
});
