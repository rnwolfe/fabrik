/**
 * NetworkPlane mirrors the Go models.NetworkPlane type.
 * Enumerates the network planes for block aggregation assignments.
 */
export type NetworkPlane = 'frontend' | 'management';

/**
 * BlockAggregation mirrors the Go models.BlockAggregation struct.
 * Represents an aggregation switch model assigned to a block for a given network plane.
 */
export interface BlockAggregation {
  id: number;
  block_id: number;
  plane: NetworkPlane;
  device_model_id: number;
  created_at: string;
  updated_at: string;
}

/**
 * BlockAggregationSummary mirrors the Go models.BlockAggregationSummary struct.
 * Extends BlockAggregation with capacity utilization.
 */
export interface BlockAggregationSummary extends BlockAggregation {
  total_ports: number;
  allocated_ports: number;
  available_ports: number;
  utilization: string;
  warning?: string;
}

/**
 * PortConnection mirrors the Go models.PortConnection struct.
 * Represents a single port allocation between a rack leaf and the block's agg switch.
 */
export interface PortConnection {
  id: number;
  block_aggregation_id: number;
  rack_id: number;
  agg_port_index: number;
  leaf_device_name: string;
  created_at: string;
}

/**
 * AddRackToBlockResult mirrors the Go models.AddRackToBlockResult struct.
 */
export interface AddRackToBlockResult {
  rack: {
    id: number;
    block_id: number | null;
    name: string;
    height_u: number;
    power_capacity_w: number;
    description: string;
    created_at: string;
    updated_at: string;
  };
  connections: PortConnection[];
  warning?: string;
}
