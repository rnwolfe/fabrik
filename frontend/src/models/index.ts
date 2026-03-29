export type DeviceRole = 'spine' | 'leaf' | 'super_spine' | 'server' | 'other' | 'management_tor' | 'management_agg';
export type DeviceModelType = 'network' | 'server' | 'storage' | 'other';
export type PortType = 'ethernet' | 'fiber' | 'dac' | 'other';
export type FabricTier = 'frontend' | 'backend';
export type NetworkPlane = 'front_end' | 'management';
export type CapacityLevel = 'rack' | 'block' | 'superblock' | 'site' | 'design';

export interface Design {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

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
  port_groups?: PortGroup[];
}

export interface PortGroup {
  id: number;
  device_model_id: number;
  count: number;
  speed_gbps: number;
  label: string;
  created_at: string;
}

export interface PortGroupInput {
  count: number;
  speed_gbps: number;
  label: string;
}

export interface Fabric {
  id: number;
  design_id: number;
  name: string;
  tier: FabricTier;
  stages: number;
  radix: number;
  original_radix?: number;
  radix_correction_note?: string;
  oversubscription: number;
  leaf_model_id?: number;
  spine_model_id?: number;
  super_spine_model_id?: number;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface TopologyPlan {
  stages: number;
  radix: number;
  spine_radix?: number;
  original_radix?: number;
  radix_correction_note?: string;
  oversubscription: number;
  leaf_count: number;
  spine_count: number;
  super_spine_count?: number;
  agg1_count?: number;
  agg2_count?: number;
  leaf_downlinks: number;
  leaf_uplinks: number;
  total_switches: number;
  total_host_ports: number;
}

export interface FabricMetrics {
  leaf_spine_oversubscription: number;
  spine_super_spine_oversubscription?: number;
}

export interface FabricResponse extends Fabric {
  topology: TopologyPlan;
  warnings?: string[];
  metrics: FabricMetrics;
  leaf_model?: DeviceModel;
  spine_model?: DeviceModel;
  super_spine_model?: DeviceModel;
}

export interface Site {
  id: number;
  design_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface SuperBlock {
  id: number;
  site_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Block {
  id: number;
  super_block_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CreateBlockResult {
  block: Block;
  racks: RackSummary[];
  warning?: string;
}

export type AggregationScope = 'block' | 'super_block' | 'site';

export interface TierAggregationSummary {
  id: number;
  scope_type: AggregationScope;
  scope_id: number;
  plane: NetworkPlane;
  device_model_id: number;
  spine_count: number;
  total_ports: number;
  allocated_ports: number;
  available_ports: number;
  utilization: string;
  warning?: string;
  created_at: string;
  updated_at: string;
}

// Backward-compat alias — prefer TierAggregationSummary for new code.
export type BlockAggregationSummary = TierAggregationSummary;

export interface TierPortConnection {
  id: number;
  tier_aggregation_id: number;
  child_id: number;
  agg_port_index: number;
  child_device_name: string;
  created_at: string;
}

// Backward-compat alias — prefer TierPortConnection for new code.
export type PortConnection = TierPortConnection;

export interface AddRackToBlockResult {
  rack: RackSummary;
  connections: PortConnection[];
  warning?: string;
}

export interface KnowledgeArticle {
  path: string;
  title: string;
  category: string;
  tags: string[];
  content?: string;
}

export interface KnowledgeIndex {
  articles: KnowledgeArticle[];
}

export interface FabricMetricEntry {
  fabric_id: number;
  fabric_name: string;
  tier: FabricTier;
  stages: number;
  leaf_spine_oversubscription: number;
  spine_super_spine_oversubscription?: number;
  total_switches: number;
  total_host_ports: number;
}

export interface ChokePoint {
  fabric_id: number;
  fabric_name: string;
  tier: FabricTier;
  ratio: number;
}

export interface PowerMetrics {
  total_capacity_w: number;
  total_draw_w: number;
  utilization_pct: number;
}

export interface ResourceCapacity {
  total_vcpu: number;
  total_ram_gb: number;
  total_storage_tb: number;
  total_gpu_count: number;
}

export interface PortUtilizationEntry {
  fabric_id: number;
  fabric_name: string;
  tier_name: string;
  total_ports: number;
  allocated_ports: number;
  available_ports: number;
}

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

export interface RackType {
  id: number;
  name: string;
  height_u: number;
  power_capacity_w: number;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface DeviceSummary {
  id: number;
  rack_id: number;
  device_model_id: number;
  name: string;
  role: DeviceRole;
  position: number;
  description: string;
  model_vendor: string;
  model_name: string;
  model_type: DeviceModelType;
  height_u: number;
  power_watts_idle: number;
  power_watts_typical: number;
  power_watts_max: number;
  cpu_sockets: number;
  cores_per_socket: number;
  ram_gb: number;
  storage_tb: number;
  gpu_count: number;
  created_at: string;
  updated_at: string;
}

export interface RackSummary {
  id: number;
  block_id?: number;
  rack_type_id?: number;
  name: string;
  height_u: number;
  power_capacity_w: number;
  description: string;
  used_u: number;
  available_u: number;
  used_watts_idle: number;
  used_watts_typical: number;
  used_watts_max: number;
  devices: DeviceSummary[];
  warning?: string;
  created_at: string;
  updated_at: string;
}
