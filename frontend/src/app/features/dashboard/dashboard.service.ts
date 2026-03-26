import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, catchError, of } from 'rxjs';
import { Design } from '../../models/design.model';

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private readonly _http = inject(HttpClient);

  getRecentDesigns(): Observable<Design[]> {
    return this._http.get<Design[]>('/api/designs').pipe(
      catchError(() => of([])),
    );
  }
}
