/**
 * DeviceRole mirrors the Go models.DeviceRole enum.
 */
export type DeviceRole = 'spine' | 'leaf' | 'super_spine' | 'server' | 'other';

/**
 * Device mirrors the Go models.Device struct.
 * A Device is a physical network device installed in a Rack.
 */
export interface Device {
  id: number;
  rack_id: number;
  device_model_id: number;
  name: string;
  role: DeviceRole;
  position: number;
  description: string;
  created_at: string;
  updated_at: string;
}
