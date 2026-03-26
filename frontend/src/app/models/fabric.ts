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
