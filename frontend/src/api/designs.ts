import { api } from './client';
import type { Design, DerivedFabric } from '@/models';

export const designsApi = {
  list: () => api.get<Design[]>('/designs'),
  get: (id: number) => api.get<Design>(`/designs/${id}`),
  create: (data: { name: string; description?: string }) => api.post<Design>('/designs', data),
  delete: (id: number) => api.delete(`/designs/${id}`),
  getDerivedFabric: (id: number) => api.get<DerivedFabric>(`/designs/${id}/fabric`),
};
