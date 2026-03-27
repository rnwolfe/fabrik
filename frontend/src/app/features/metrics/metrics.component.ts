import {
  ChangeDetectionStrategy,
  Component,
  OnInit,
  OnDestroy,
  inject,
  signal,
  computed,
} from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { Subject, takeUntil, of, catchError, interval, startWith } from 'rxjs';
import { HttpClient } from '@angular/common/http';

import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';
import { FormsModule } from '@angular/forms';

import { MetricsService } from './metrics.service';
import { Design, DesignMetrics } from '../../models';

/**
 * MetricsComponent displays the at-a-glance metrics dashboard for a selected design.
 * Shows oversubscription, power, resource capacity, and port utilization.
 */
@Component({
  selector: 'app-metrics',
  standalone: true,
  imports: [
    DecimalPipe,
    FormsModule,
    MatCardModule,
    MatProgressBarModule,
    MatChipsModule,
    MatIconModule,
    MatSelectModule,
    MatDividerModule,
    MatProgressSpinnerModule,
    MatTooltipModule,
  ],
  templateUrl: './metrics.component.html',
  styleUrl: './metrics.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class MetricsComponent implements OnInit, OnDestroy {
  private readonly http = inject(HttpClient);
  private readonly metricsSvc = inject(MetricsService);
  private readonly destroy$ = new Subject<void>();
  private readonly refresh$ = new Subject<void>();

  designs = signal<Design[]>([]);
  selectedDesignId = signal<number | null>(null);
  metrics = signal<DesignMetrics | null>(null);
  loading = signal(false);
  error = signal<string | null>(null);

  /** Power utilization percentage clamped to [0, 100]. */
  powerUtilizationPct = computed(() => {
    const m = this.metrics();
    if (!m) return 0;
    return Math.min(m.power.utilization_pct, 100);
  });

  ngOnInit(): void {
    // Load all designs for the selector.
    this.http.get<Design[]>('/api/designs').pipe(
      catchError(() => of([] as Design[])),
      takeUntil(this.destroy$),
    ).subscribe(designs => {
      this.designs.set(designs);
      if (designs.length > 0 && this.selectedDesignId() === null) {
        this.selectedDesignId.set(designs[0].id);
        this.loadMetrics();
      }
    });

    // Auto-refresh every 30 seconds when a design is selected.
    interval(30_000).pipe(
      startWith(0),
      takeUntil(this.destroy$),
    ).subscribe(() => {
      if (this.selectedDesignId() !== null) {
        this.loadMetrics();
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onDesignChange(id: number): void {
    this.selectedDesignId.set(id);
    this.loadMetrics();
  }

  private loadMetrics(): void {
    const id = this.selectedDesignId();
    if (id === null) return;

    this.loading.set(true);
    this.error.set(null);

    this.metricsSvc.getDesignMetrics(id).pipe(
      catchError(() => {
        this.error.set('Failed to load metrics. Please try again.');
        this.loading.set(false);
        return of(null);
      }),
      takeUntil(this.destroy$),
    ).subscribe(m => {
      if (m !== null) {
        this.metrics.set(m);
      }
      this.loading.set(false);
    });
  }

  /** Returns a CSS color class for an oversubscription ratio. */
  oversubClass(ratio: number): string {
    if (ratio <= 1.5) return 'metric-good';
    if (ratio <= 3.0) return 'metric-warn';
    return 'metric-bad';
  }

  /** Formats an oversubscription ratio as "N:1". */
  formatOversub(ratio: number): string {
    return `${ratio.toFixed(1)}:1`;
  }

  /** Returns port utilization percentage for display. */
  portUtilizationPct(total: number, allocated: number): number {
    if (total === 0) return 0;
    return Math.round((allocated / total) * 100);
  }
}
