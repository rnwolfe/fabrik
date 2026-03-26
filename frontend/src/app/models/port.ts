/**
 * PortType mirrors the Go models.PortType enum.
 */
export type PortType = 'ethernet' | 'fiber' | 'dac' | 'other';

/**
 * Port mirrors the Go models.Port struct.
 * A Port is a physical network port on a Device.
 */
export interface Port {
  id: number;
  device_id: number;
  name: string;
  type: PortType;
  speed_gbps: number;
  description: string;
  created_at: string;
  updated_at: string;
}
