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

