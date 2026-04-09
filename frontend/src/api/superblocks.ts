import { api } from './client';
import type { SuperBlock } from '@/models';

export const superblocksApi = {
  list: (siteId: number) =>
    api.get<SuperBlock[]>(`/sites/${siteId}/superblocks`),
  create: (siteId: number, data: { name: string; description?: string }) =>
    api.post<SuperBlock>(`/sites/${siteId}/superblocks`, data),
  get: (id: number) =>
    api.get<SuperBlock>(`/superblocks/${id}`),
  update: (id: number, data: { name: string; description?: string }) =>
    api.put<SuperBlock>(`/superblocks/${id}`, data),
  patch: (id: number, data: { name?: string; description?: string }) =>
    api.patch<SuperBlock>(`/superblocks/${id}`, data),
  delete: (id: number) =>
    api.delete(`/superblocks/${id}`),
};
