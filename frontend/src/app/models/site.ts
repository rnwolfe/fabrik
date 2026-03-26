/**
 * Site mirrors the Go models.Site struct.
 * A Site is a physical datacenter location within a Design.
 */
export interface Site {
  id: number;
  design_id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}
