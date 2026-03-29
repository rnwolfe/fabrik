import { api } from './client';
import type { DeviceModel, PortGroupInput } from '@/models';

export interface DeviceModelRequest {
  vendor: string;
  model: string;
  device_model_type: string;
  port_count: number;
  port_groups?: PortGroupInput[];
  height_u: number;
  power_watts_idle: number;
  power_watts_typical: number;
  power_watts_max: number;
  cpu_sockets: number;
  cores_per_socket: number;
  ram_gb: number;
  storage_tb: number;
  gpu_count: number;
  description: string;
}

export const catalogApi = {
  list: () => api.get<DeviceModel[]>('/catalog/devices'),
  get: (id: number) => api.get<DeviceModel>(`/catalog/devices/${id}`),
  create: (data: DeviceModelRequest) => api.post<DeviceModel>('/catalog/devices', data),
  update: (id: number, data: DeviceModelRequest) => api.put<DeviceModel>(`/catalog/devices/${id}`, data),
  delete: (id: number) => api.delete(`/catalog/devices/${id}`),
  duplicate: (id: number) => api.post<DeviceModel>(`/catalog/devices/${id}/duplicate`, {}),
};
