import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { BlockAggregation, SetManagementAggRequest } from '../../models';

@Injectable({ providedIn: 'root' })
export class ManagementService {
  private readonly _http = inject(HttpClient);

  setManagementAgg(blockId: number, req: SetManagementAggRequest): Observable<BlockAggregation> {
    return this._http.put<BlockAggregation>(`/api/blocks/${blockId}/management-agg`, req);
  }

  getManagementAgg(blockId: number): Observable<BlockAggregation> {
    return this._http.get<BlockAggregation>(`/api/blocks/${blockId}/management-agg`);
  }

  removeManagementAgg(blockId: number): Observable<void> {
    return this._http.delete<void>(`/api/blocks/${blockId}/management-agg`);
  }

  listBlockAggregations(blockId: number): Observable<BlockAggregation[]> {
    return this._http.get<BlockAggregation[]>(`/api/blocks/${blockId}/aggregations`);
  }
}
