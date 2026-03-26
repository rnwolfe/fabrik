/**
 * RackType mirrors the Go models.RackType enum.
 */
export type RackType = 'physical' | 'logical';

/**
 * Rack mirrors the Go models.Rack struct.
 * A Rack is a physical or logical rack within a Block.
 */
export interface Rack {
  id: number;
  block_id: number;
  name: string;
  type: RackType;
  height_u: number;
  description: string;
  created_at: string;
  updated_at: string;
}
