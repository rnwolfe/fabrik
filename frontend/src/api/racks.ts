import { api } from './client';
import type { RackSummary, RackType, RackRole } from '@/models';

export interface RackInput {
  name: string;
  description?: string;
  rack_type_id?: number;
  block_id?: number;
  height_u?: number;
  power_capacity_w?: number;
  role?: RackRole;
}

export interface RackTypeInput {
  name: string;
  height_u: number;
  power_capacity_w: number;
  description?: string;
}

export const racksApi = {
  listTypes: () => api.get<RackType[]>('/rack-types'),
  createType: (data: RackTypeInput) => api.post<RackType>('/rack-types', data),
  updateType: (id: number, data: Partial<RackTypeInput>) => api.put<RackType>(`/rack-types/${id}`, data),
  deleteType: (id: number) => api.delete(`/rack-types/${id}`),
  list: () => api.get<RackSummary[]>('/racks'),
  get: (id: number) => api.get<RackSummary>(`/racks/${id}`),
  create: (data: RackInput) => api.post<RackSummary>('/racks', data),
  update: (id: number, data: Partial<RackInput>) => api.put<RackSummary>(`/racks/${id}`, data),
  delete: (id: number) => api.delete(`/racks/${id}`),
  placeServerDevices: (rackId: number, deviceModelId: number, count: number) =>
    api.post<void>(`/racks/${rackId}/server-devices`, { device_model_id: deviceModelId, count }),
};
