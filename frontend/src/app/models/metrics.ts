/**
 * FabricMetricEntry holds per-fabric oversubscription and topology metrics.
 * Mirrors models.FabricMetricEntry in the Go backend.
 */
export interface FabricMetricEntry {
  fabric_id: number;
  fabric_name: string;
  tier: string;
  stages: number;
  leaf_spine_oversubscription: number;
  spine_super_spine_oversubscription: number;
  total_switches: number;
  total_host_ports: number;
}

/**
 * ChokePoint identifies the worst-case oversubscription tier.
 * Mirrors models.ChokePoint in the Go backend.
 */
export interface ChokePoint {
  fabric_id: number;
  fabric_name: string;
  tier: string;
  ratio: number;
}

/**
 * PowerMetrics holds power consumption totals.
 * Mirrors models.PowerMetrics in the Go backend.
 */
export interface PowerMetrics {
  total_capacity_w: number;
  total_draw_w: number;
  utilization_pct: number;
}

/**
 * ResourceCapacity holds compute resource totals.
 * Mirrors models.ResourceCapacity in the Go backend.
 */
export interface ResourceCapacity {
  total_vcpu: number;
  total_ram_gb: number;
  total_storage_tb: number;
  total_gpu_count: number;
}

/**
 * PortUtilizationEntry holds port utilization for a single fabric tier.
 * Mirrors models.PortUtilizationEntry in the Go backend.
 */
export interface PortUtilizationEntry {
  fabric_id: number;
  fabric_name: string;
  tier_name: string;
  total_ports: number;
  allocated_ports: number;
  available_ports: number;
}

/**
 * DesignMetrics is the top-level response from GET /api/designs/:id/metrics.
 * Mirrors models.DesignMetrics in the Go backend.
 */
export interface DesignMetrics {
  design_id: number;
  total_hosts: number;
  total_switches: number;
  bisection_bandwidth_gbps: number;
  fabrics: FabricMetricEntry[];
  choke_point?: ChokePoint;
  power: PowerMetrics;
  capacity: ResourceCapacity;
  port_utilization: PortUtilizationEntry[];
  empty: boolean;
}
