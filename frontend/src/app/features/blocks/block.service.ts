import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
  Block,
  BlockAggregationSummary,
  PortConnection,
  AddRackToBlockResult,
  NetworkPlane,
} from '../../models';

/** Request body for creating a block. */
export interface CreateBlockPayload {
  super_block_id: number;
  name: string;
  description?: string;
}

/** Request body for assigning an aggregation switch to a block plane. */
export interface AssignAggregationPayload {
  device_model_id: number;
}

/** Request body for adding a rack to a block. */
export interface AddRackToBlockPayload {
  rack_id: number;
  block_id?: number | null;
  super_block_id?: number;
}

/**
 * BlockService communicates with the Go REST API for blocks and
 * block-level aggregation operations.
 */
@Injectable({ providedIn: 'root' })
export class BlockService {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/blocks';

  // --- Blocks ---

  /** Creates a new block under a super-block. */
  createBlock(payload: CreateBlockPayload): Observable<Block> {
    return this.http.post<Block>(this.base, payload);
  }

  /** Gets a single block by ID. */
  getBlock(id: number): Observable<Block> {
    return this.http.get<Block>(`${this.base}/${id}`);
  }

  /** Lists all blocks for a given super-block. */
  listBlocks(superBlockId: number): Observable<Block[]> {
    return this.http.get<Block[]>(this.base, {
      params: { super_block_id: String(superBlockId) },
    });
  }

  // --- Aggregation ---

  /**
   * Assigns (or replaces) the aggregation switch model for a block's network plane.
   * Returns the updated aggregation summary with capacity utilization.
   */
  assignAggregation(
    blockId: number,
    plane: NetworkPlane,
    payload: AssignAggregationPayload,
  ): Observable<BlockAggregationSummary> {
    return this.http.put<BlockAggregationSummary>(
      `${this.base}/${blockId}/aggregations/${plane}`,
      payload,
    );
  }

  /** Gets the aggregation summary for a block plane. */
  getAggregation(blockId: number, plane: NetworkPlane): Observable<BlockAggregationSummary> {
    return this.http.get<BlockAggregationSummary>(
      `${this.base}/${blockId}/aggregations/${plane}`,
    );
  }

  /** Lists all aggregation summaries for a block. */
  listAggregations(blockId: number): Observable<BlockAggregationSummary[]> {
    return this.http.get<BlockAggregationSummary[]>(
      `${this.base}/${blockId}/aggregations`,
    );
  }

  /**
   * Deletes the aggregation assignment for a block plane.
   * All port connections are also removed.
   */
  deleteAggregation(blockId: number, plane: NetworkPlane): Observable<void> {
    return this.http.delete<void>(`${this.base}/${blockId}/aggregations/${plane}`);
  }

  // --- Port Connections ---

  /** Lists all port connections for a block's aggregation plane. */
  listPortConnections(blockId: number, plane: NetworkPlane): Observable<PortConnection[]> {
    return this.http.get<PortConnection[]>(
      `${this.base}/${blockId}/aggregations/${plane}/connections`,
    );
  }

  // --- Rack Placement ---

  /**
   * Adds a rack to a block and auto-allocates agg ports for leaf devices.
   * If block_id is omitted, a default block is auto-created under super_block_id.
   */
  addRackToBlock(payload: AddRackToBlockPayload): Observable<AddRackToBlockResult> {
    return this.http.post<AddRackToBlockResult>(`${this.base}/add-rack`, payload);
  }

  /**
   * Removes a rack from its block and deallocates all agg port connections.
   */
  removeRackFromBlock(rackId: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/racks/${rackId}`);
  }
}
