/**
 * DeviceModel mirrors the Go models.DeviceModel struct.
 * A DeviceModel is a hardware platform catalog entry (e.g., Arista 7050).
 */
export interface DeviceModel {
  id: number;
  vendor: string;
  model: string;
  port_count: number;
  height_u: number;
  power_watts: number;
  description: string;
  created_at: string;
  updated_at: string;
}
