/**
 * RackTemplate mirrors the Go models.RackTemplate struct.
 * A RackTemplate is a named template defining rack hardware specifications.
 */
export interface RackTemplate {
  id: number;
  name: string;
  height_u: number;
  power_capacity_w: number;
  description: string;
  created_at: string;
  updated_at: string;
}

/**
 * Rack mirrors the Go models.Rack struct.
 * A Rack is a physical rack (optionally within a Block).
 */
export interface Rack {
  id: number;
  block_id: number | null;
  rack_type_id: number | null;
  name: string;
  height_u: number;
  power_capacity_w: number;
  description: string;
  created_at: string;
  updated_at: string;
}

/**
 * DeviceSummary mirrors the Go models.DeviceSummary struct.
 * A DeviceSummary is a device with model info included.
 */
export interface DeviceSummary {
  id: number;
  rack_id: number;
  device_model_id: number;
  name: string;
  role: string;
  position: number;
  description: string;
  created_at: string;
  updated_at: string;
  model_vendor: string;
  model_name: string;
  height_u: number;
  power_watts: number;
}

/**
 * RackSummary mirrors the Go models.RackSummary struct.
 * A RackSummary is a rack with computed usage metrics and device list.
 */
export interface RackSummary extends Rack {
  used_u: number;
  available_u: number;
  used_watts: number;
  devices: DeviceSummary[];
  warning?: string;
}

/**
 * PlaceDeviceResult mirrors the Go models.PlaceDeviceResult struct.
 */
export interface PlaceDeviceResult {
  device: {
    id: number;
    rack_id: number;
    device_model_id: number;
    name: string;
    role: string;
    position: number;
    description: string;
    created_at: string;
    updated_at: string;
  };
  warning?: string;
}
