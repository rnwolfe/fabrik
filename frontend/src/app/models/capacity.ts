/**
 * CapacityLevel enumerates the hierarchy level at which capacity is aggregated.
 */
export type CapacityLevel = 'rack' | 'block' | 'superblock' | 'site' | 'design';

/**
 * CapacitySummary mirrors the Go models.CapacitySummary struct.
 * It holds aggregated power and resource totals for a hierarchy level.
 */
export interface CapacitySummary {
  level: CapacityLevel;
  id: number;
  name: string;
  power_watts_idle: number;
  power_watts_typical: number;
  power_watts_max: number;
  total_vcpu: number;
  total_ram_gb: number;
  total_storage_tb: number;
  total_gpu_count: number;
  device_count: number;
}
