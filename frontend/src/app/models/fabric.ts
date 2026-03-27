import { DeviceModel } from './device-model';

/**
 * FabricTier mirrors the Go models.FabricTier enum.
 */
export type FabricTier = 'frontend' | 'backend';

/**
 * Fabric mirrors the Go models.Fabric struct.
 * A Fabric is a Clos fabric tier (front-end or back-end) within a Design.
 */
export interface Fabric {
  id: number;
  design_id: number;
  name: string;
  tier: FabricTier;
  description: string;
  created_at: string;
  updated_at: string;
}

/**
 * FabricRecord extends Fabric with topology parameters stored in the database.
 */
export interface FabricRecord extends Fabric {
  stages: number;
  radix: number;
  oversubscription: number;
  leaf_model_id?: number;
  spine_model_id?: number;
  super_spine_model_id?: number;
}

/**
 * TopologyPlan holds the calculated switch counts and port distribution.
 * Mirrors service.TopologyPlan in the Go backend.
 */
export interface TopologyPlan {
  stages: number;
  radix: number;
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

/**
 * FabricMetrics holds derived metrics for a fabric.
 */
export interface FabricMetrics {
  total_switches: number;
  total_host_ports: number;
  oversubscription_ratio: number;
  bisection_bandwidth_gbps?: number;
}

/**
 * FabricResponse is the enriched API response including topology and metrics.
 */
export interface FabricResponse extends FabricRecord {
  topology: TopologyPlan;
  warnings?: string[];
  metrics: FabricMetrics;
  leaf_model?: DeviceModel;
  spine_model?: DeviceModel;
  super_spine_model?: DeviceModel;
}

/**
 * CreateFabricRequest is the body for POST /api/fabrics.
 */
export interface CreateFabricRequest {
  design_id: number;
  name: string;
  tier: FabricTier;
  stages: number;
  radix: number;
  oversubscription: number;
  description?: string;
  leaf_model_id?: number;
  spine_model_id?: number;
  super_spine_model_id?: number;
}

/**
 * UpdateFabricRequest is the body for PUT /api/fabrics/:id.
 */
export interface UpdateFabricRequest {
  name: string;
  tier: FabricTier;
  stages: number;
  radix: number;
  oversubscription: number;
  description?: string;
  leaf_model_id?: number;
  spine_model_id?: number;
  super_spine_model_id?: number;
  force?: boolean;
}

/**
 * PreviewRequest is the body for POST /api/fabrics/preview.
 */
export interface PreviewRequest {
  stages: number;
  radix: number;
  oversubscription: number;
}
