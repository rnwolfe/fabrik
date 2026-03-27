import {
  Component,
  OnInit,
  inject,
  signal,
  computed,
  output,
  input,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatListModule } from '@angular/material/list';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { DeviceCatalogService } from '../../features/catalog/device-catalog.service';
import { DeviceModel } from '../../models/device-model';

/**
 * DevicePickerComponent — shared component for selecting a DeviceModel from the
 * catalog. Emits the selected model via (deviceSelected). Supports drag-and-drop
 * by setting draggable=true on list items.
 */
@Component({
  selector: 'app-device-picker',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    MatListModule,
    MatInputModule,
    MatFormFieldModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './device-picker.component.html',
  styleUrls: ['./device-picker.component.scss'],
})
export class DevicePickerComponent implements OnInit {
  private readonly catalogSvc = inject(DeviceCatalogService);

  /** The currently selected device model id, if any. */
  readonly selectedId = input<number | null>(null);

  /** Emitted when the user selects a device model. */
  readonly deviceSelected = output<DeviceModel>();

  /** Emitted when the user starts dragging a device model item. */
  readonly deviceDragStart = output<DeviceModel>();

  readonly deviceModels = signal<DeviceModel[]>([]);
  readonly loading = signal(true);
  readonly searchTerm = signal('');

  readonly filteredModels = computed(() => {
    const term = this.searchTerm().toLowerCase();
    return this.deviceModels().filter(
      dm =>
        !term ||
        dm.vendor.toLowerCase().includes(term) ||
        dm.model.toLowerCase().includes(term),
    );
  });

  ngOnInit(): void {
    this.catalogSvc.list().subscribe({
      next: models => {
        this.deviceModels.set(models);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      },
    });
  }

  select(dm: DeviceModel): void {
    this.deviceSelected.emit(dm);
  }

  onDragStart(event: DragEvent, dm: DeviceModel): void {
    if (event.dataTransfer) {
      event.dataTransfer.setData('application/json', JSON.stringify(dm));
      event.dataTransfer.effectAllowed = 'copy';
    }
    this.deviceDragStart.emit(dm);
  }

  onSearchChange(value: string): void {
    this.searchTerm.set(value);
  }

  isSelected(dm: DeviceModel): boolean {
    return this.selectedId() === dm.id;
  }
}
