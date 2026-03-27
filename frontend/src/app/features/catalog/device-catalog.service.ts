import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DeviceModel, DeviceModelPayload } from '../../models/device-model';

/** Provides CRUD access to the /api/catalog/devices REST endpoint. */
@Injectable({ providedIn: 'root' })
export class DeviceCatalogService {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/catalog/devices';

  /** List device models. Pass includeArchived=true to include archived records. */
  list(includeArchived = false): Observable<DeviceModel[]> {
    const params = includeArchived
      ? new HttpParams().set('include_archived', 'true')
      : new HttpParams();
    return this.http.get<DeviceModel[]>(this.base, { params });
  }

  /** Get a single device model by id. */
  get(id: number): Observable<DeviceModel> {
    return this.http.get<DeviceModel>(`${this.base}/${id}`);
  }

  /** Create a new device model. */
  create(payload: DeviceModelPayload): Observable<DeviceModel> {
    return this.http.post<DeviceModel>(this.base, payload);
  }

  /** Update an existing device model (forbidden for seed models). */
  update(id: number, payload: DeviceModelPayload): Observable<DeviceModel> {
    return this.http.put<DeviceModel>(`${this.base}/${id}`, payload);
  }

  /** Archive (soft-delete) a device model (forbidden for seed models). */
  archive(id: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/${id}`);
  }

  /** Clone a device model as a new non-seed row. */
  duplicate(id: number): Observable<DeviceModel> {
    return this.http.post<DeviceModel>(`${this.base}/${id}/duplicate`, {});
  }
}
