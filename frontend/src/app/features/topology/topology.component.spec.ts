import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { Observable, of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { Component, Input, Output, EventEmitter } from '@angular/core';

import { TopologyComponent } from './topology.component';
import { FabricService } from './fabric.service';
import { FabricResponse, TopologyPlan } from '../../models/fabric';
import { TopologyGraphComponent, NodeClickEvent, EdgeClickEvent } from './topology-graph.component';
import { TopologyDetailPanelComponent, DetailItem } from './topology-detail-panel.component';

/** Stub for TopologyGraphComponent to avoid Cytoscape canvas initialization in tests */
@Component({
  selector: 'app-topology-graph',
  standalone: true,
  template: '<div class="cy-stub"></div>',
})
class TopologyGraphStubComponent {
  @Input() fabrics: FabricResponse[] = [];
  @Input() expandAll = false;
  @Output() nodeClick = new EventEmitter<NodeClickEvent>();
  @Output() edgeClick = new EventEmitter<EdgeClickEvent>();
}

/** Stub for TopologyDetailPanelComponent */
@Component({
  selector: 'app-topology-detail-panel',
  standalone: true,
  template: '<div class="detail-stub"></div>',
})
class TopologyDetailPanelStubComponent {
  @Input() item: DetailItem | null = null;
  @Output() closed = new EventEmitter<void>();
}

const mockPlan: TopologyPlan = {
  stages: 2,
  radix: 64,
  oversubscription: 1.0,
  leaf_count: 1,
  spine_count: 32,
  leaf_uplinks: 32,
  leaf_downlinks: 32,
  total_switches: 33,
  total_host_ports: 32,
};

const mockFabric: FabricResponse = {
  id: 1,
  design_id: 0,
  name: 'test-fabric',
  tier: 'frontend',
  stages: 2,
  radix: 64,
  oversubscription: 1.0,
  description: '',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  topology: mockPlan,
  metrics: {
    total_switches: 33,
    total_host_ports: 32,
    oversubscription_ratio: 1.0,
  },
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyFn = (...args: any[]) => any;

interface FabricServiceMock {
  listFabrics: ReturnType<typeof vi.fn<AnyFn>>;
  createFabric: ReturnType<typeof vi.fn<AnyFn>>;
  updateFabric: ReturnType<typeof vi.fn<AnyFn>>;
  deleteFabric: ReturnType<typeof vi.fn<AnyFn>>;
  previewTopology: ReturnType<typeof vi.fn<AnyFn>>;
  listDeviceModels: ReturnType<typeof vi.fn<AnyFn>>;
}

function createFabricServiceMock(): FabricServiceMock {
  return {
    listFabrics: vi.fn(() => of([] as FabricResponse[])),
    createFabric: vi.fn(() => of(mockFabric)),
    updateFabric: vi.fn(() => of(mockFabric)),
    deleteFabric: vi.fn(() => of(undefined as void)),
    previewTopology: vi.fn(() => of(mockPlan)),
    listDeviceModels: vi.fn(() => of([])),
  };
}

describe('TopologyComponent', () => {
  let component: TopologyComponent;
  let fixture: ComponentFixture<TopologyComponent>;
  let fabricSvc: FabricServiceMock;

  beforeEach(async () => {
    fabricSvc = createFabricServiceMock();

    await TestBed.configureTestingModule({
      imports: [TopologyComponent],
      providers: [
        provideAnimations(),
        provideRouter([]),
        provideHttpClient(),
        { provide: FabricService, useValue: fabricSvc },
      ],
    })
      .overrideComponent(TopologyComponent, {
        remove: { imports: [TopologyGraphComponent, TopologyDetailPanelComponent] },
        add: { imports: [TopologyGraphStubComponent, TopologyDetailPanelStubComponent] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(TopologyComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load fabrics on init', () => {
    expect(fabricSvc.listFabrics).toHaveBeenCalled();
  });

  it('should load device models on init', () => {
    expect(fabricSvc.listDeviceModels).toHaveBeenCalled();
  });

  it('should show empty state when no fabrics', () => {
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.empty-state')).toBeTruthy();
  });

  it('should open create form when openCreateForm called', () => {
    component.openCreateForm();
    fixture.detectChanges();
    expect(component.showForm()).toBe(true);
    expect(component.editingFabric()).toBeNull();
  });

  it('should show designer card after opening create form', () => {
    component.openCreateForm();
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.designer-card')).toBeTruthy();
  });

  it('should cancel form and hide it', () => {
    component.openCreateForm();
    fixture.detectChanges();
    component.cancelForm();
    fixture.detectChanges();
    expect(component.showForm()).toBe(false);
  });

  it('should populate form when editing a fabric', () => {
    component.openEditForm(mockFabric);
    expect(component.editingFabric()).toEqual(mockFabric);
    expect(component.form.get('name')?.value).toBe('test-fabric');
    expect(component.form.get('stages')?.value).toBe(2);
    expect(component.form.get('radix')?.value).toBe(64);
  });

  it('should mark form invalid and not call create when name is empty', () => {
    component.openCreateForm();
    component.form.get('name')?.setValue('');
    component.saveFabric();
    expect(fabricSvc.createFabric).not.toHaveBeenCalled();
  });

  it('should call createFabric service when form is valid', () => {
    component.openCreateForm();
    component.form.patchValue({
      name: 'new-fabric',
      tier: 'frontend',
      stages: 2,
      radix: 64,
      oversubscription: 1.0,
    });
    component.saveFabric();
    expect(fabricSvc.createFabric).toHaveBeenCalled();
  });

  it('should call updateFabric service when editing and saving', () => {
    component.openEditForm(mockFabric);
    component.form.patchValue({ name: 'updated-name' });
    component.saveFabric();
    expect(fabricSvc.updateFabric).toHaveBeenCalledWith(
      1,
      expect.objectContaining({ name: 'updated-name' }),
    );
  });

  it('should call deleteFabric when _deleteFabric invoked', () => {
    component['_deleteFabric'](mockFabric);
    expect(fabricSvc.deleteFabric).toHaveBeenCalledWith(1);
  });

  it('stagesForPreview returns 2 layers for 2-stage', () => {
    const layers = component.stagesForPreview(mockPlan);
    expect(layers.length).toBe(2);
    expect(layers[0].role).toBe('Leaf');
    expect(layers[1].role).toBe('Spine');
  });

  it('stagesForPreview includes SuperSpine for 3-stage plan', () => {
    const plan3: TopologyPlan = {
      ...mockPlan,
      stages: 3,
      super_spine_count: 2,
    };
    const layers = component.stagesForPreview(plan3);
    expect(layers.some(l => l.role === 'Super-Spine')).toBe(true);
  });

  it('stagesForPreview includes agg layers for 5-stage plan', () => {
    const plan5: TopologyPlan = {
      ...mockPlan,
      stages: 5,
      super_spine_count: 2,
      agg1_count: 32,
      agg2_count: 2,
    };
    const layers = component.stagesForPreview(plan5);
    expect(layers.some(l => l.role === 'Aggregation-1')).toBe(true);
    expect(layers.some(l => l.role === 'Aggregation-2')).toBe(true);
  });

  it('should update fabrics signal when _loadFabrics called', () => {
    fabricSvc.listFabrics.mockReturnValue(of([mockFabric]) as Observable<FabricResponse[]>);
    component['_loadFabrics']();
    fixture.detectChanges();
    expect(component.fabrics().length).toBe(1);
  });

  it('should display fabric name in the list', () => {
    fabricSvc.listFabrics.mockReturnValue(of([mockFabric]) as Observable<FabricResponse[]>);
    component['_loadFabrics']();
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('test-fabric');
  });

  it('should show no-models hint when device models list is empty', () => {
    component.openCreateForm();
    fixture.detectChanges();
    expect(component.hasModels()).toBe(false);
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.no-models-hint')).toBeTruthy();
  });

  it('should gracefully handle listFabrics error', () => {
    fabricSvc.listFabrics.mockReturnValue(throwError(() => new Error('network error')));
    component['_loadFabrics']();
    expect(component.fabrics()).toEqual([]);
    expect(component.isLoading()).toBe(false);
  });

  it('should show all stage labels defined', () => {
    expect(component.stageLabel[2]).toContain('2-Stage');
    expect(component.stageLabel[3]).toContain('3-Stage');
    expect(component.stageLabel[5]).toContain('5-Stage');
  });

  it('should set livePreview when fabric is opened for editing', () => {
    component.openEditForm(mockFabric);
    expect(component.livePreview()).toEqual(mockFabric.topology);
  });

  it('should clear serverWarnings when creating new fabric', () => {
    component.serverWarnings.set(['old warning']);
    component.openCreateForm();
    expect(component.serverWarnings()).toEqual([]);
  });

  // Management plane toggle tests
  it('showManagementPlane should default to false', () => {
    expect(component.showManagementPlane()).toBe(false);
  });

  it('toggleManagementPlane should toggle the signal from false to true', () => {
    expect(component.showManagementPlane()).toBe(false);
    component.toggleManagementPlane();
    expect(component.showManagementPlane()).toBe(true);
  });

  it('toggleManagementPlane should toggle back from true to false', () => {
    component.showManagementPlane.set(true);
    component.toggleManagementPlane();
    expect(component.showManagementPlane()).toBe(false);
  });

  it('should show management plane panel when showManagementPlane is true', () => {
    component.showManagementPlane.set(true);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.management-plane-panel')).toBeTruthy();
  });

  it('should hide management plane panel when showManagementPlane is false', () => {
    component.showManagementPlane.set(false);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.management-plane-panel')).toBeNull();
  });

  it('management plane panel should contain legend with distinct node colors', () => {
    component.showManagementPlane.set(true);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const torDot = el.querySelector('.mgmt-node-tor');
    const aggDot = el.querySelector('.mgmt-node-agg');
    const linkLine = el.querySelector('.mgmt-link-line');
    expect(torDot).toBeTruthy();
    expect(aggDot).toBeTruthy();
    expect(linkLine).toBeTruthy();
  });

  it('management plane panel should have accessible role and label', () => {
    component.showManagementPlane.set(true);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const panel = el.querySelector('.management-plane-panel');
    expect(panel?.getAttribute('role')).toBe('region');
    const label = panel?.getAttribute('aria-label') ?? '';
    expect(label.toLowerCase()).toContain('management');
  });

  it('toggle button should be present in page header', () => {
    const el: HTMLElement = fixture.nativeElement;
    const toggle = el.querySelector('.mgmt-plane-toggle');
    expect(toggle).toBeTruthy();
  });
});
