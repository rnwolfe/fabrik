import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Rack, RackSummary, RackTemplate, PlaceDeviceResult } from '../../models';

/** Request body for creating or updating a rack type. */
export interface RackTypePayload {
  name: string;
  description?: string;
  height_u: number;
  power_capacity_w: number;
}

/** Request body for creating a rack. */
export interface CreateRackPayload {
  name: string;
  description?: string;
  block_id?: number | null;
  rack_type_id?: number | null;
  height_u?: number;
  power_capacity_w?: number;
}

/** Request body for updating a rack. */
export interface UpdateRackPayload {
  name: string;
  description?: string;
  block_id?: number | null;
}

/** Request body for placing a device. */
export interface PlaceDevicePayload {
  device_model_id: number;
  name: string;
  description?: string;
  role?: string;
  position?: number;
}

/** Request body for moving a device within a rack. */
export interface MoveDevicePayload {
  position: number;
}

/** Request body for moving a device to a different rack. */
export interface MoveCrossRackPayload {
  dest_rack_id: number;
  position?: number;
}

/**
 * RackService communicates with the Go REST API for rack types, racks,
 * and device placement operations.
 */
@Injectable({ providedIn: 'root' })
export class RackService {
  private readonly http = inject(HttpClient);
  private readonly rackTypesBase = '/api/rack-types';
  private readonly racksBase = '/api/racks';

  // --- Rack Types ---

  /** Lists all rack type templates. */
  listRackTypes(): Observable<RackTemplate[]> {
    return this.http.get<RackTemplate[]>(this.rackTypesBase);
  }

  /** Gets a single rack type. */
  getRackType(id: number): Observable<RackTemplate> {
    return this.http.get<RackTemplate>(`${this.rackTypesBase}/${id}`);
  }

  /** Creates a new rack type. */
  createRackType(payload: RackTypePayload): Observable<RackTemplate> {
    return this.http.post<RackTemplate>(this.rackTypesBase, payload);
  }

  /** Updates an existing rack type. */
  updateRackType(id: number, payload: RackTypePayload): Observable<RackTemplate> {
    return this.http.put<RackTemplate>(`${this.rackTypesBase}/${id}`, payload);
  }

  /** Deletes a rack type. Returns 409 if racks reference it. */
  deleteRackType(id: number): Observable<void> {
    return this.http.delete<void>(`${this.rackTypesBase}/${id}`);
  }

  // --- Racks ---

  /** Lists all racks, optionally filtered by block. */
  listRacks(blockId?: number): Observable<Rack[]> {
    if (blockId != null) {
      return this.http.get<Rack[]>(this.racksBase, { params: { block_id: String(blockId) } });
    }
    return this.http.get<Rack[]>(this.racksBase);
  }

  /** Gets a rack with usage summary and device list. */
  getRack(id: number): Observable<RackSummary> {
    return this.http.get<RackSummary>(`${this.racksBase}/${id}`);
  }

  /** Creates a new rack. */
  createRack(payload: CreateRackPayload): Observable<Rack> {
    return this.http.post<Rack>(this.racksBase, payload);
  }

  /** Updates a rack's name, description, and block assignment. */
  updateRack(id: number, payload: UpdateRackPayload): Observable<Rack> {
    return this.http.put<Rack>(`${this.racksBase}/${id}`, payload);
  }

  /** Deletes a rack and all its devices. */
  deleteRack(id: number): Observable<void> {
    return this.http.delete<void>(`${this.racksBase}/${id}`);
  }

  // --- Device Placement ---

  /** Places a device in a rack at the given position. */
  placeDevice(rackId: number, payload: PlaceDevicePayload): Observable<PlaceDeviceResult> {
    return this.http.post<PlaceDeviceResult>(`${this.racksBase}/${rackId}/devices`, payload);
  }

  /** Moves a device to a new position within the same rack. */
  moveDeviceInRack(rackId: number, deviceId: number, payload: MoveDevicePayload): Observable<PlaceDeviceResult> {
    return this.http.put<PlaceDeviceResult>(`${this.racksBase}/${rackId}/devices/${deviceId}`, payload);
  }

  /** Moves a device to a different rack. */
  moveDeviceCrossRack(rackId: number, deviceId: number, payload: MoveCrossRackPayload): Observable<PlaceDeviceResult> {
    return this.http.put<PlaceDeviceResult>(`${this.racksBase}/${rackId}/devices/${deviceId}/move`, payload);
  }

  /**
   * Removes a device from a rack.
   * @param compact When true, devices above the gap shift down to fill it.
   */
  removeDevice(rackId: number, deviceId: number, compact = false): Observable<void> {
    const params = { compact: String(compact) };
    return this.http.delete<void>(`${this.racksBase}/${rackId}/devices/${deviceId}`, { params });
  }
}
