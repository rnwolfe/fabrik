import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DesignMetrics } from '../../models';

/**
 * MetricsService communicates with the Go REST API for design metrics.
 */
@Injectable({ providedIn: 'root' })
export class MetricsService {
  private readonly http = inject(HttpClient);

  /**
   * Fetches all computed metrics for the given design.
   */
  getDesignMetrics(designId: number): Observable<DesignMetrics> {
    return this.http.get<DesignMetrics>(`/api/designs/${designId}/metrics`);
  }
}
