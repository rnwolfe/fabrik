import { api } from './client';
import type { Block, BlockAggregationSummary, AddRackToBlockResult, PortConnection } from '@/models';

export const blocksApi = {
  list: (superBlockId: number) =>
    api.get<Block[]>(`/blocks?super_block_id=${superBlockId}`),
  create: (data: { super_block_id: number; name: string; description?: string }) =>
    api.post<Block>('/blocks', data),
  get: (id: number) => api.get<Block>(`/blocks/${id}`),

  // Aggregation (spine/leaf model assignment per plane)
  assignAggregation: (blockId: number, plane: string, deviceModelId: number) =>
    api.put<BlockAggregationSummary>(`/blocks/${blockId}/aggregations/${plane}`, {
      device_model_id: deviceModelId,
    }),
  getAggregation: (blockId: number, plane: string) =>
    api.get<BlockAggregationSummary>(`/blocks/${blockId}/aggregations/${plane}`),
  listAggregations: (blockId: number) =>
    api.get<BlockAggregationSummary[]>(`/blocks/${blockId}/aggregations`),
  deleteAggregation: (blockId: number, plane: string) =>
    api.delete(`/blocks/${blockId}/aggregations/${plane}`),

  // Rack placement
  addRack: (data: { rack_id: number; block_id?: number; super_block_id?: number }) =>
    api.post<AddRackToBlockResult>('/blocks/add-rack', data),
  removeRack: (rackId: number) => api.delete(`/block-racks/${rackId}`),

  // Port connections
  listConnections: (blockId: number, plane: string) =>
    api.get<PortConnection[]>(`/blocks/${blockId}/aggregations/${plane}/connections`),
};

export const scaffoldApi = {
  get: (designId: number) =>
    api.get<{ site_id: number; super_block_id: number }>(`/designs/${designId}/scaffold`),
};
