import { api } from './client';
import type { DesignMetrics } from '@/models';

export const metricsApi = {
  getDesignMetrics: (designId: number) => api.get<DesignMetrics>(`/designs/${designId}/metrics`),
};
