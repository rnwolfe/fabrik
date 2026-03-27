/**
 * RackTemplate mirrors the Go models.RackTemplate struct.
 * A RackTemplate is a named template defining rack hardware specifications.
 */
export interface RackTemplate {
  id: number;
  name: string;
  height_u: number;
  power_capacity_w: number;
  power_oversub_pct_warn: number;
  power_oversub_pct_max: number;
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
  power_oversub_pct_warn: number;
  power_oversub_pct_max: number;
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
  model_type: string;
  height_u: number;
  power_watts_idle: number;
  power_watts_typical: number;
  power_watts_max: number;
  cpu_sockets: number;
  cores_per_socket: number;
  ram_gb: number;
  storage_tb: number;
  gpu_count: number;
}

/**
 * RackSummary mirrors the Go models.RackSummary struct.
 * A RackSummary is a rack with computed usage metrics and device list.
 */
export interface RackSummary extends Rack {
  used_u: number;
  available_u: number;
  used_watts_idle: number;
  used_watts_typical: number;
  used_watts_max: number;
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
