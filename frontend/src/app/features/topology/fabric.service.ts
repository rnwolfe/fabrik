import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import {
  FabricResponse,
  CreateFabricRequest,
  UpdateFabricRequest,
  TopologyPlan,
  PreviewRequest,
} from '../../models/fabric';
import { DeviceModel } from '../../models/device-model';

@Injectable({ providedIn: 'root' })
export class FabricService {
  private readonly _http = inject(HttpClient);

  createFabric(req: CreateFabricRequest): Observable<FabricResponse> {
    return this._http.post<FabricResponse>('/api/fabrics', req);
  }

  listFabrics(): Observable<FabricResponse[]> {
    return this._http.get<FabricResponse[]>('/api/fabrics');
  }

  getFabric(id: number): Observable<FabricResponse> {
    return this._http.get<FabricResponse>(`/api/fabrics/${id}`);
  }

  updateFabric(id: number, req: UpdateFabricRequest): Observable<FabricResponse> {
    return this._http.put<FabricResponse>(`/api/fabrics/${id}`, req);
  }

  deleteFabric(id: number): Observable<void> {
    return this._http.delete<void>(`/api/fabrics/${id}`);
  }

  previewTopology(req: PreviewRequest): Observable<TopologyPlan> {
    return this._http.post<TopologyPlan>('/api/fabrics/preview', req);
  }

  listDeviceModels(): Observable<DeviceModel[]> {
    return this._http.get<DeviceModel[]>('/api/device-models');
  }
}
