import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { DeviceCatalogService } from './device-catalog.service';
import { DeviceModelFormComponent } from './device-model-form/device-model-form.component';
import { DeviceModel, DeviceModelPayload } from '../../models/device-model';

@Component({
  selector: 'app-catalog',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatInputModule,
    MatFormFieldModule,
    MatSelectModule,
    MatDialogModule,
    MatSnackBarModule,
    MatTooltipModule,
    MatChipsModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './catalog.component.html',
  styleUrls: ['./catalog.component.scss'],
})
export class CatalogComponent implements OnInit {
  private readonly catalogSvc = inject(DeviceCatalogService);
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);

  readonly deviceModels = signal<DeviceModel[]>([]);
  readonly loading = signal(true);
  readonly searchTerm = signal('');
  readonly vendorFilter = signal('');
  readonly typeFilter = signal('');

  readonly displayedColumns = ['vendor', 'model', 'port_count', 'height_u', 'power_watts', 'seed', 'actions'];

  readonly vendors = computed(() => {
    const seen = new Set<string>();
    this.deviceModels().forEach(dm => seen.add(dm.vendor));
    return Array.from(seen).sort();
  });

  readonly filteredModels = computed(() => {
    const term = this.searchTerm().toLowerCase();
    const vendor = this.vendorFilter().toLowerCase();
    const type = this.typeFilter().toLowerCase();

    return this.deviceModels().filter(dm => {
      const matchesSearch =
        !term ||
        dm.vendor.toLowerCase().includes(term) ||
        dm.model.toLowerCase().includes(term);

      const matchesVendor =
        !vendor || dm.vendor.toLowerCase() === vendor;

      const matchesType =
        !type ||
        (type === 'switch' && dm.port_count > 0) ||
        (type === 'server' && dm.port_count === 0);

      return matchesSearch && matchesVendor && matchesType;
    });
  });

  ngOnInit(): void {
    this.loadModels();
  }

  loadModels(): void {
    this.loading.set(true);
    this.catalogSvc.list().subscribe({
      next: models => {
        this.deviceModels.set(models);
        this.loading.set(false);
      },
      error: () => {
        this.snackBar.open('Failed to load device catalog', 'Dismiss', { duration: 4000 });
        this.loading.set(false);
      },
    });
  }

  openAddDialog(): void {
    const ref = this.dialog.open(DeviceModelFormComponent, {
      width: '560px',
      data: { deviceModel: null },
    });
    ref.afterClosed().subscribe((payload: DeviceModelPayload | undefined) => {
      if (!payload) return;
      this.catalogSvc.create(payload).subscribe({
        next: () => {
          this.snackBar.open('Device model created', 'Dismiss', { duration: 3000 });
          this.loadModels();
        },
        error: () => {
          this.snackBar.open('Failed to create device model', 'Dismiss', { duration: 4000 });
        },
      });
    });
  }

  openEditDialog(dm: DeviceModel): void {
    const ref = this.dialog.open(DeviceModelFormComponent, {
      width: '560px',
      data: { deviceModel: dm },
    });
    ref.afterClosed().subscribe((payload: DeviceModelPayload | undefined) => {
      if (!payload) return;
      this.catalogSvc.update(dm.id, payload).subscribe({
        next: () => {
          this.snackBar.open('Device model updated', 'Dismiss', { duration: 3000 });
          this.loadModels();
        },
        error: () => {
          this.snackBar.open('Failed to update device model', 'Dismiss', { duration: 4000 });
        },
      });
    });
  }

  duplicate(dm: DeviceModel): void {
    this.catalogSvc.duplicate(dm.id).subscribe({
      next: () => {
        this.snackBar.open('Device model duplicated', 'Dismiss', { duration: 3000 });
        this.loadModels();
      },
      error: () => {
        this.snackBar.open('Failed to duplicate device model', 'Dismiss', { duration: 4000 });
      },
    });
  }

  archive(dm: DeviceModel): void {
    this.catalogSvc.archive(dm.id).subscribe({
      next: () => {
        this.snackBar.open('Device model archived', 'Dismiss', { duration: 3000 });
        this.loadModels();
      },
      error: () => {
        this.snackBar.open('Failed to archive device model', 'Dismiss', { duration: 4000 });
      },
    });
  }

  onSearchChange(value: string): void {
    this.searchTerm.set(value);
  }

  onVendorFilterChange(value: string): void {
    this.vendorFilter.set(value);
  }

  onTypeFilterChange(value: string): void {
    this.typeFilter.set(value);
  }
}
