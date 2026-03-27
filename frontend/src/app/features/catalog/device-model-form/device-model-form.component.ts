import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { DeviceModel, DeviceModelPayload } from '../../../models/device-model';

export interface DeviceModelFormData {
  deviceModel: DeviceModel | null;
}

@Component({
  selector: 'app-device-model-form',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
  ],
  templateUrl: './device-model-form.component.html',
})
export class DeviceModelFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly dialogRef = inject(MatDialogRef<DeviceModelFormComponent>);
  readonly data: DeviceModelFormData = inject(MAT_DIALOG_DATA);

  form!: FormGroup;

  get isEdit(): boolean {
    return !!this.data?.deviceModel;
  }

  get title(): string {
    return this.isEdit ? 'Edit Device Model' : 'Add Device Model';
  }

  ngOnInit(): void {
    const dm = this.data?.deviceModel;
    this.form = this.fb.group({
      vendor: [dm?.vendor ?? '', [Validators.required]],
      model: [dm?.model ?? '', [Validators.required]],
      device_model_type: [dm?.device_model_type ?? 'network'],
      port_count: [dm?.port_count ?? 0, [Validators.min(0)]],
      height_u: [dm?.height_u ?? 1, [Validators.required, Validators.min(1), Validators.max(50)]],
      power_watts_idle: [dm?.power_watts_idle ?? 0, [Validators.min(0)]],
      power_watts_typical: [dm?.power_watts_typical ?? 0, [Validators.min(0)]],
      power_watts_max: [dm?.power_watts_max ?? 0, [Validators.min(0)]],
      cpu_sockets: [dm?.cpu_sockets ?? 0, [Validators.min(0)]],
      cores_per_socket: [dm?.cores_per_socket ?? 0, [Validators.min(0)]],
      ram_gb: [dm?.ram_gb ?? 0, [Validators.min(0)]],
      storage_tb: [dm?.storage_tb ?? 0, [Validators.min(0)]],
      gpu_count: [dm?.gpu_count ?? 0, [Validators.min(0)]],
      description: [dm?.description ?? ''],
    });
  }

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const payload: DeviceModelPayload = this.form.value;
    this.dialogRef.close(payload);
  }

  cancel(): void {
    this.dialogRef.close();
  }
}
