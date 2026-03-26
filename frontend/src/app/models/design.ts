/**
 * Design mirrors the Go models.Design struct.
 * A Design is the top-level container for a datacenter network planning project.
 */
export interface Design {
  id: number;
  name: string;
  description: string;
  created_at: string; // ISO 8601
  updated_at: string; // ISO 8601
}
