/**
 * SuperBlock mirrors the Go models.SuperBlock struct.
 * A SuperBlock groups Blocks within a Site (e.g., a data hall or pod).
 */
export interface SuperBlock {
  id: number;
  site_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}
