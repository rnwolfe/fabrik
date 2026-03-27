/**
 * DeviceModelType enumerates the category of a device model.
 */
export type DeviceModelType = 'network' | 'server' | 'storage' | 'other';

/**
 * DeviceModel mirrors the Go models.DeviceModel struct.
 * A DeviceModel is a hardware platform catalog entry (e.g., Arista 7050).
 */
export interface DeviceModel {
  id: number;
  vendor: string;
  model: string;
  device_model_type: DeviceModelType;
  port_count: number;
  height_u: number;
  power_watts_idle: number;
  power_watts_typical: number;
  power_watts_max: number;
  // Server resource fields (0 for non-server models)
  cpu_sockets: number;
  cores_per_socket: number;
  ram_gb: number;
  storage_tb: number;
  gpu_count: number;
  description: string;
  is_seed: boolean;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

/** Payload for creating or updating a device model. */
export interface DeviceModelPayload {
  vendor: string;
  model: string;
  device_model_type: DeviceModelType;
  port_count: number;
  height_u: number;
  power_watts_idle: number;
  power_watts_typical: number;
  power_watts_max: number;
  cpu_sockets: number;
  cores_per_socket: number;
  ram_gb: number;
  storage_tb: number;
  gpu_count: number;
  description: string;
}
