import { api } from './client';
import type { RackSummary, RackType } from '@/models';

export const racksApi = {
  listTypes: () => api.get<RackType[]>('/rack-types'),
  createType: (data: Partial<RackType>) => api.post<RackType>('/rack-types', data),
  updateType: (id: number, data: Partial<RackType>) => api.put<RackType>(`/rack-types/${id}`, data),
  deleteType: (id: number) => api.delete(`/rack-types/${id}`),
  list: () => api.get<RackSummary[]>('/racks'),
  get: (id: number) => api.get<RackSummary>(`/racks/${id}`),
  create: (data: Partial<RackSummary>) => api.post<RackSummary>('/racks', data),
  update: (id: number, data: Partial<RackSummary>) => api.put<RackSummary>(`/racks/${id}`, data),
  delete: (id: number) => api.delete(`/racks/${id}`),
};
