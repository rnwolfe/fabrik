import { api } from './client';
import type { Site, DesignHierarchy } from '@/models';

export const sitesApi = {
  list: (designId: number) =>
    api.get<Site[]>(`/designs/${designId}/sites`),
  create: (designId: number, data: { name: string; description?: string }) =>
    api.post<Site>(`/designs/${designId}/sites`, data),
  get: (id: number) =>
    api.get<Site>(`/sites/${id}`),
  update: (id: number, data: { name: string; description?: string }) =>
    api.put<Site>(`/sites/${id}`, data),
  delete: (id: number) =>
    api.delete(`/sites/${id}`),

  // Full hierarchy tree: site → superblock → block
  getHierarchy: (designId: number) =>
    api.get<DesignHierarchy>(`/designs/${designId}/hierarchy`),
};
