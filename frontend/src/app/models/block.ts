/**
 * Block mirrors the Go models.Block struct.
 * A Block is a logical grouping of Racks within a SuperBlock (e.g., a row or cluster).
 */
export interface Block {
  id: number;
  super_block_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

/**
 * NetworkPlane mirrors the Go models.NetworkPlane type.
 */
export type NetworkPlane = 'front_end' | 'management';

/**
 * BlockAggregation mirrors the Go models.BlockAggregation struct.
 * Represents a block-level aggregation switch assignment for a given network plane.
 */
export interface BlockAggregation {
  id: number;
  block_id: number;
  plane: NetworkPlane;
  device_id: number | null;
  max_ports: number;
  used_ports: number;
  description: string;
  created_at: string;
  updated_at: string;
}

/**
 * SetManagementAggRequest is the body for PUT /api/blocks/:block_id/management-agg.
 */
export interface SetManagementAggRequest {
  device_id?: number | null;
  max_ports: number;
  description?: string;
}
