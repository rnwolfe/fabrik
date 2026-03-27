import {
  Component,
  OnInit,
  OnDestroy,
  inject,
  signal,
  computed,
} from '@angular/core';
import {
  FormBuilder,
  FormGroup,
  Validators,
  ReactiveFormsModule,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import {
  debounceTime,
  distinctUntilChanged,
  Subject,
  takeUntil,
  switchMap,
  of,
  catchError,
} from 'rxjs';

import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatChipsModule } from '@angular/material/chips';
import { MatListModule } from '@angular/material/list';
import { MatTooltipModule } from '@angular/material/tooltip';

import { FabricService } from './fabric.service';
import { ConfirmDialogComponent } from './confirm-dialog.component';
import {
  FabricResponse,
  TopologyPlan,
  CreateFabricRequest,
  UpdateFabricRequest,
  FabricTier,
} from '../../models/fabric';
import { DeviceModel } from '../../models/device-model';

function positiveNumberValidator(control: AbstractControl): ValidationErrors | null {
  const val = Number(control.value);
  if (!Number.isFinite(val) || val <= 0) {
    return { positiveNumber: true };
  }
  return null;
}

function minValueValidator(min: number) {
  return (control: AbstractControl): ValidationErrors | null => {
    const val = Number(control.value);
    if (!Number.isFinite(val) || val < min) {
      return { minValue: { min, actual: val } };
    }
    return null;
  };
}

@Component({
  selector: 'app-topology',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatButtonModule,
    MatIconModule,
    MatDividerModule,
    MatProgressSpinnerModule,
    MatDialogModule,
    MatSnackBarModule,
    MatChipsModule,
    MatListModule,
    MatTooltipModule,
  ],
  templateUrl: './topology.component.html',
  styleUrl: './topology.component.scss',
})
export class TopologyComponent implements OnInit, OnDestroy {
  private readonly _fb = inject(FormBuilder);
  private readonly _svc = inject(FabricService);
  private readonly _dialog = inject(MatDialog);
  private readonly _snack = inject(MatSnackBar);
  private readonly _destroy$ = new Subject<void>();

  fabrics = signal<FabricResponse[]>([]);
  deviceModels = signal<DeviceModel[]>([]);
  isLoading = signal(false);
  isSaving = signal(false);
  editingFabric = signal<FabricResponse | null>(null);
  showForm = signal(false);
  livePreview = signal<TopologyPlan | null>(null);
  previewLoading = signal(false);
  previewError = signal<string | null>(null);
  serverWarnings = signal<string[]>([]);

  readonly stageCounts = [2, 3, 5] as const;

  readonly form: FormGroup = this._fb.group({
    name: ['', [Validators.required, Validators.minLength(1)]],
    tier: ['frontend' as FabricTier, Validators.required],
    stages: [2, Validators.required],
    radix: [64, [Validators.required, positiveNumberValidator]],
    oversubscription: [1.0, [Validators.required, minValueValidator(1.0)]],
    description: [''],
    leaf_model_id: [null],
    spine_model_id: [null],
    super_spine_model_id: [null],
  });

  readonly hasModels = computed(() => this.deviceModels().length > 0);

  readonly stageLabel: Record<number, string> = {
    2: '2-Stage (Leaf-Spine)',
    3: '3-Stage (Leaf-Spine-SuperSpine)',
    5: '5-Stage (Extended Clos)',
  };

  ngOnInit(): void {
    this._loadFabrics();
    this._loadDeviceModels();

    this.form.valueChanges
      .pipe(
        debounceTime(300),
        distinctUntilChanged(
          (a, b) =>
            a.stages === b.stages &&
            a.radix === b.radix &&
            a.oversubscription === b.oversubscription,
        ),
        switchMap(() => {
          const { stages, radix, oversubscription } = this.form.getRawValue();
          const s = Number(stages);
          const r = Number(radix);
          const os = Number(oversubscription);

          if (![2, 3, 5].includes(s) || r <= 0 || os < 1.0) {
            return of(null);
          }

          this.previewLoading.set(true);
          this.previewError.set(null);

          return this._svc
            .previewTopology({ stages: s, radix: r, oversubscription: os })
            .pipe(
              catchError(err => {
                const msg = (err?.error?.error as string | undefined) ?? 'Preview failed';
                this.previewError.set(msg);
                return of(null);
              }),
            );
        }),
        takeUntil(this._destroy$),
      )
      .subscribe(plan => {
        this.previewLoading.set(false);
        if (plan) {
          this.livePreview.set(plan);
          if (plan.radix_correction_note) {
            this.form.patchValue({ radix: plan.radix }, { emitEvent: false });
          }
        }
      });
  }

  ngOnDestroy(): void {
    this._destroy$.next();
    this._destroy$.complete();
  }

  openCreateForm(): void {
    this.editingFabric.set(null);
    this.form.reset({
      name: '',
      tier: 'frontend',
      stages: 2,
      radix: 64,
      oversubscription: 1.0,
      description: '',
      leaf_model_id: null,
      spine_model_id: null,
      super_spine_model_id: null,
    });
    this.livePreview.set(null);
    this.serverWarnings.set([]);
    this.showForm.set(true);
  }

  openEditForm(fabric: FabricResponse): void {
    this.editingFabric.set(fabric);
    this.form.reset({
      name: fabric.name,
      tier: fabric.tier,
      stages: fabric.stages,
      radix: fabric.radix,
      oversubscription: fabric.oversubscription,
      description: fabric.description ?? '',
      leaf_model_id: fabric.leaf_model_id ?? null,
      spine_model_id: fabric.spine_model_id ?? null,
      super_spine_model_id: fabric.super_spine_model_id ?? null,
    });
    this.livePreview.set(fabric.topology);
    this.serverWarnings.set(fabric.warnings ?? []);
    this.showForm.set(true);
  }

  cancelForm(): void {
    this.showForm.set(false);
    this.editingFabric.set(null);
  }

  saveFabric(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const val = this.form.getRawValue();
    this.isSaving.set(true);
    this.serverWarnings.set([]);

    const editing = this.editingFabric();
    if (editing) {
      const req: UpdateFabricRequest = {
        name: val.name,
        tier: val.tier,
        stages: Number(val.stages),
        radix: Number(val.radix),
        oversubscription: Number(val.oversubscription),
        description: val.description ?? '',
        leaf_model_id: val.leaf_model_id ?? undefined,
        spine_model_id: val.spine_model_id ?? undefined,
        super_spine_model_id: val.super_spine_model_id ?? undefined,
      };
      this._svc
        .updateFabric(editing.id, req)
        .pipe(takeUntil(this._destroy$))
        .subscribe({
          next: resp => {
            this.isSaving.set(false);
            if (resp.warnings?.length) {
              this.serverWarnings.set(resp.warnings);
            }
            this._snack.open(`Fabric "${resp.name}" updated`, 'Dismiss', { duration: 3000 });
            this._loadFabrics();
            this.showForm.set(false);
          },
          error: err => {
            this.isSaving.set(false);
            const msg = (err?.error?.error as string | undefined) ?? 'Failed to update fabric';
            this._snack.open(msg, 'Dismiss', { duration: 5000 });
          },
        });
    } else {
      const req: CreateFabricRequest = {
        design_id: 0,
        name: val.name,
        tier: val.tier,
        stages: Number(val.stages),
        radix: Number(val.radix),
        oversubscription: Number(val.oversubscription),
        description: val.description ?? '',
        leaf_model_id: val.leaf_model_id ?? undefined,
        spine_model_id: val.spine_model_id ?? undefined,
        super_spine_model_id: val.super_spine_model_id ?? undefined,
      };
      this._svc
        .createFabric(req)
        .pipe(takeUntil(this._destroy$))
        .subscribe({
          next: resp => {
            this.isSaving.set(false);
            if (resp.warnings?.length) {
              this.serverWarnings.set(resp.warnings);
            }
            this._snack.open(`Fabric "${resp.name}" created`, 'Dismiss', { duration: 3000 });
            this._loadFabrics();
            this.showForm.set(false);
          },
          error: err => {
            this.isSaving.set(false);
            const msg = (err?.error?.error as string | undefined) ?? 'Failed to create fabric';
            this._snack.open(msg, 'Dismiss', { duration: 5000 });
          },
        });
    }
  }

  confirmDelete(fabric: FabricResponse): void {
    const ref = this._dialog.open(ConfirmDialogComponent, {
      data: {
        title: 'Delete fabric',
        message: `Delete "${fabric.name}"? This cannot be undone.`,
        confirmLabel: 'Delete',
        confirmColor: 'warn',
      },
    });

    ref
      .afterClosed()
      .pipe(takeUntil(this._destroy$))
      .subscribe(confirmed => {
        if (confirmed) {
          this._deleteFabric(fabric);
        }
      });
  }

  private _deleteFabric(fabric: FabricResponse): void {
    this._svc
      .deleteFabric(fabric.id)
      .pipe(takeUntil(this._destroy$))
      .subscribe({
        next: () => {
          this._snack.open(`Fabric "${fabric.name}" deleted`, 'Dismiss', { duration: 3000 });
          this._loadFabrics();
        },
        error: err => {
          const msg = (err?.error?.error as string | undefined) ?? 'Failed to delete fabric';
          this._snack.open(msg, 'Dismiss', { duration: 5000 });
        },
      });
  }

  private _loadFabrics(): void {
    this.isLoading.set(true);
    this._svc
      .listFabrics()
      .pipe(takeUntil(this._destroy$))
      .subscribe({
        next: fabrics => {
          this.isLoading.set(false);
          this.fabrics.set(fabrics ?? []);
        },
        error: () => {
          this.isLoading.set(false);
          this.fabrics.set([]);
        },
      });
  }

  private _loadDeviceModels(): void {
    this._svc
      .listDeviceModels()
      .pipe(
        catchError(() => of([])),
        takeUntil(this._destroy$),
      )
      .subscribe(models => {
        this.deviceModels.set(models ?? []);
      });
  }

  stagesForPreview(plan: TopologyPlan): { role: string; count: number }[] {
    const layers: { role: string; count: number }[] = [
      { role: 'Leaf', count: plan.leaf_count },
      { role: 'Spine', count: plan.spine_count },
    ];
    if (plan.super_spine_count) {
      layers.push({ role: 'Super-Spine', count: plan.super_spine_count });
    }
    if (plan.agg1_count) {
      layers.push({ role: 'Aggregation-1', count: plan.agg1_count });
    }
    if (plan.agg2_count) {
      layers.push({ role: 'Aggregation-2', count: plan.agg2_count });
    }
    return layers;
  }
}
