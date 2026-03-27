import { Component, Input, OnChanges, SimpleChanges, inject, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatChipsModule } from '@angular/material/chips';
import { CapacitySummary, CapacityLevel } from '../../../../models';
import { CapacityService } from '../../capacity.service';

/**
 * CapacitySummaryComponent displays power and resource capacity totals
 * for a selected hierarchy level (Rack → Block → SuperBlock → Site → Design).
 *
 * Usage:
 *   <app-capacity-summary [designId]="1" [level]="'design'" />
 *   <app-capacity-summary [designId]="1" [level]="'rack'" [entityId]="7" />
 */
@Component({
  selector: 'app-capacity-summary',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatDividerModule, MatProgressBarModule, MatChipsModule],
  template: `
    <mat-card *ngIf="summary">
      <mat-card-header>
        <mat-card-title>{{ summary.name }}</mat-card-title>
        <mat-card-subtitle>{{ summary.level | titlecase }} · {{ summary.device_count }} device(s)</mat-card-subtitle>
      </mat-card-header>

      <mat-card-content>
        <section aria-label="Power consumption">
          <h3>Power Consumption</h3>
          <dl>
            <div>
              <dt>Idle</dt>
              <dd>{{ summary.power_watts_idle | number }} W</dd>
            </div>
            <div>
              <dt>Typical</dt>
              <dd>{{ summary.power_watts_typical | number }} W</dd>
            </div>
            <div>
              <dt>Max</dt>
              <dd>{{ summary.power_watts_max | number }} W</dd>
            </div>
          </dl>
        </section>

        <mat-divider />

        <section aria-label="Compute resources">
          <h3>Compute Resources</h3>
          <dl>
            <div>
              <dt>vCPU</dt>
              <dd>{{ summary.total_vcpu | number }}</dd>
            </div>
            <div>
              <dt>RAM</dt>
              <dd>{{ summary.total_ram_gb | number }} GB</dd>
            </div>
            <div>
              <dt>Storage</dt>
              <dd>{{ summary.total_storage_tb | number:'1.1-2' }} TB</dd>
            </div>
            <div *ngIf="summary.total_gpu_count > 0">
              <dt>GPU</dt>
              <dd>{{ summary.total_gpu_count | number }}</dd>
            </div>
          </dl>
        </section>
      </mat-card-content>
    </mat-card>

    <p *ngIf="error" role="alert">{{ error }}</p>
  `,
})
export class CapacitySummaryComponent implements OnChanges {
  @Input() designId!: number;
  @Input() level: CapacityLevel = 'design';
  @Input() entityId?: number;

  summary: CapacitySummary | null = null;
  error: string | null = null;

  private readonly capacitySvc = inject(CapacityService);
  private readonly cdr = inject(ChangeDetectorRef);

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['designId'] || changes['level'] || changes['entityId']) {
      this.load();
    }
  }

  private load(): void {
    this.error = null;
    this.capacitySvc.getCapacity(this.designId, this.level, this.entityId).subscribe({
      next: (c) => { this.summary = c; this.cdr.markForCheck(); },
      error: () => { this.error = 'Failed to load capacity data.'; this.cdr.markForCheck(); },
    });
  }
}
