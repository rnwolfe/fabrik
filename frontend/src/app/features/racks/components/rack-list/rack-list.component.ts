import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatDialogModule } from '@angular/material/dialog';

import { RackService } from '../../rack.service';
import { Rack } from '../../../../models';

/**
 * RackListComponent displays all racks with their capacity summary
 * and allows creating, navigating to, and deleting racks.
 */
@Component({
  selector: 'app-rack-list',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatChipsModule,
    MatTooltipModule,
    MatSnackBarModule,
    MatDialogModule,
  ],
  template: `
    <div class="rack-list-page">
      <header class="page-header">
        <h1>Rack Management</h1>
        <p class="subtitle">Plan and manage physical rack placements.</p>
      </header>

      <div class="toolbar">
        <button mat-raised-button color="primary" (click)="createRack()" aria-label="Create rack">
          <mat-icon>add</mat-icon>
          Add Rack
        </button>
      </div>

      @if (loading()) {
        <div class="loading-center" role="status" aria-label="Loading racks">
          <mat-spinner diameter="48"></mat-spinner>
        </div>
      } @else if (error()) {
        <div class="error-banner" role="alert">
          <mat-icon>error</mat-icon>
          <span>{{ error() }}</span>
          <button mat-button (click)="loadRacks()">Retry</button>
        </div>
      } @else if (racks().length === 0) {
        <div class="empty-state" role="status">
          <mat-icon class="empty-icon">dns</mat-icon>
          <h2>No racks yet</h2>
          <p>Create your first rack to start planning device placement.</p>
          <button mat-raised-button color="primary" (click)="createRack()">Add Rack</button>
        </div>
      } @else {
        <div class="rack-grid">
          @for (rack of racks(); track rack.id) {
            <mat-card class="rack-card" appearance="outlined">
              <mat-card-header>
                <mat-card-title>{{ rack.name }}</mat-card-title>
                <mat-card-subtitle>
                  {{ rack.height_u }}U &bull; {{ rack.power_capacity_w > 0 ? (rack.power_capacity_w + 'W') : 'No power limit' }}
                </mat-card-subtitle>
              </mat-card-header>

              <mat-card-content>
                @if (rack.description) {
                  <p class="rack-description">{{ rack.description }}</p>
                }
                <div class="rack-meta">
                  @if (rack.block_id) {
                    <mat-chip>Block {{ rack.block_id }}</mat-chip>
                  } @else {
                    <mat-chip>Standalone</mat-chip>
                  }
                  @if (rack.rack_type_id) {
                    <mat-chip>Type {{ rack.rack_type_id }}</mat-chip>
                  }
                </div>
              </mat-card-content>

              <mat-card-actions align="end">
                <button
                  mat-icon-button
                  color="warn"
                  [matTooltip]="'Delete rack'"
                  aria-label="Delete rack"
                  (click)="deleteRack(rack)"
                >
                  <mat-icon>delete</mat-icon>
                </button>
              </mat-card-actions>
            </mat-card>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .rack-list-page {
      padding: 1.5rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .page-header {
      margin-bottom: 1.5rem;
    }

    .page-header h1 {
      margin: 0 0 0.25rem;
      font-size: 1.75rem;
      font-weight: 500;
    }

    .subtitle {
      margin: 0;
      color: var(--mat-sys-on-surface-variant);
    }

    .toolbar {
      margin-bottom: 1.5rem;
      display: flex;
      gap: 0.5rem;
    }

    .loading-center {
      display: flex;
      justify-content: center;
      padding: 3rem;
    }

    .error-banner {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 1rem;
      border-radius: 8px;
      background: var(--mat-sys-error-container);
      color: var(--mat-sys-on-error-container);
    }

    .empty-state {
      text-align: center;
      padding: 4rem 2rem;
      color: var(--mat-sys-on-surface-variant);
    }

    .empty-icon {
      font-size: 64px;
      width: 64px;
      height: 64px;
      margin-bottom: 1rem;
    }

    .empty-state h2 {
      margin: 0 0 0.5rem;
      font-size: 1.25rem;
    }

    .empty-state p {
      margin: 0 0 1.5rem;
    }

    .rack-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 1rem;
    }

    .rack-card {
      transition: box-shadow 0.2s ease;
    }

    .rack-description {
      font-size: 0.875rem;
      color: var(--mat-sys-on-surface-variant);
      margin-bottom: 0.75rem;
    }

    .rack-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }
  `],
})
export class RackListComponent implements OnInit {
  private readonly rackService = inject(RackService);
  private readonly snackBar = inject(MatSnackBar);

  readonly racks = signal<Rack[]>([]);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);

  ngOnInit(): void {
    this.loadRacks();
  }

  loadRacks(): void {
    this.loading.set(true);
    this.error.set(null);
    this.rackService.listRacks().subscribe({
      next: racks => {
        this.racks.set(racks);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Failed to load racks. Please try again.');
        this.loading.set(false);
      },
    });
  }

  createRack(): void {
    const name = window.prompt('Enter rack name:');
    if (!name?.trim()) return;

    const heightStr = window.prompt('Enter rack height in U (default: 42):', '42');
    const heightU = Math.max(1, parseInt(heightStr ?? '42', 10) || 42);

    const powerStr = window.prompt('Enter power capacity in Watts (0 for no limit):', '0');
    const powerCapacityW = Math.max(0, parseInt(powerStr ?? '0', 10) || 0);

    this.rackService.createRack({ name: name.trim(), height_u: heightU, power_capacity_w: powerCapacityW }).subscribe({
      next: rack => {
        this.racks.update(racks => [...racks, rack]);
        this.snackBar.open(`Rack "${rack.name}" created`, 'Dismiss', { duration: 3000 });
      },
      error: () => {
        this.snackBar.open('Failed to create rack', 'Dismiss', { duration: 3000 });
      },
    });
  }

  deleteRack(rack: Rack): void {
    if (!window.confirm(`Delete rack "${rack.name}"? This will remove all devices in it.`)) return;

    this.rackService.deleteRack(rack.id).subscribe({
      next: () => {
        this.racks.update(racks => racks.filter(r => r.id !== rack.id));
        this.snackBar.open(`Rack "${rack.name}" deleted`, 'Dismiss', { duration: 3000 });
      },
      error: () => {
        this.snackBar.open('Failed to delete rack', 'Dismiss', { duration: 3000 });
      },
    });
  }
}
