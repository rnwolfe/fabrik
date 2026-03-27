import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CapacitySummary, CapacityLevel } from '../../models';

/**
 * CapacityService communicates with the Go REST API for power and resource
 * capacity aggregation at any hierarchy level.
 */
@Injectable({ providedIn: 'root' })
export class CapacityService {
  private readonly http = inject(HttpClient);

  /**
   * Fetches capacity aggregation for the given design at design level.
   */
  getDesignCapacity(designId: number): Observable<CapacitySummary> {
    return this.http.get<CapacitySummary>(`/api/designs/${designId}/capacity`);
  }

  /**
   * Fetches capacity aggregation at the requested hierarchy level.
   * For levels other than "design", entityId must identify the specific entity.
   */
  getCapacity(designId: number, level: CapacityLevel, entityId?: number): Observable<CapacitySummary> {
    let params = new HttpParams().set('level', level);
    if (level !== 'design' && entityId !== undefined) {
      params = params.set('entity_id', entityId.toString());
    }
    return this.http.get<CapacitySummary>(`/api/designs/${designId}/capacity`, { params });
  }
}
