import { api } from './client';
import type { FabricResponse, TopologyPlan } from '@/models';

export interface CreateFabricInput {
  design_id: number;
  name: string;
  tier: string;
  stages: number;
  radix: number;
  oversubscription: number;
  /** 0 = full fabric (populate all spine ports); 1 = minimum (single leaf) */
  leaf_count?: number;
  leaf_model_id?: number;
  spine_model_id?: number;
  super_spine_model_id?: number;
  description?: string;
}

export const fabricsApi = {
  list: () => api.get<FabricResponse[]>('/fabrics'),
  get: (id: number) => api.get<FabricResponse>(`/fabrics/${id}`),
  preview: (data: Omit<CreateFabricInput, 'design_id' | 'name' | 'description'>) =>
    api.post<{ topology: TopologyPlan; warnings?: string[] }>('/fabrics/preview', data),
  create: (data: CreateFabricInput) => api.post<FabricResponse>('/fabrics', data),
  update: (id: number, data: Partial<CreateFabricInput>) =>
    api.put<FabricResponse>(`/fabrics/${id}`, data),
  delete: (id: number) => api.delete(`/fabrics/${id}`),
};
