import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideRouter } from '@angular/router';

import { TopologyGraphComponent, NodeClickEvent, EdgeClickEvent } from './topology-graph.component';
import { FabricResponse } from '../../models/fabric';
import { NodeData, EdgeData } from './topology-graph.service';

const mockFabric: FabricResponse = {
  id: 1,
  design_id: 0,
  name: 'test-fabric',
  tier: 'frontend',
  stages: 2,
  radix: 4,
  oversubscription: 1.0,
  description: '',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  topology: {
    stages: 2,
    radix: 4,
    oversubscription: 1.0,
    leaf_count: 2,
    spine_count: 2,
    leaf_uplinks: 2,
    leaf_downlinks: 2,
    total_switches: 4,
    total_host_ports: 4,
  },
  metrics: {
    total_switches: 4,
    total_host_ports: 4,
    oversubscription_ratio: 1.0,
  },
};

describe('TopologyGraphComponent', () => {
  let component: TopologyGraphComponent;
  let fixture: ComponentFixture<TopologyGraphComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TopologyGraphComponent],
      providers: [provideAnimations(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(TopologyGraphComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should show empty state when fabrics is empty', () => {
    component.fabrics = [];
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.graph-empty')).toBeTruthy();
  });

  it('should hide empty state when fabrics are provided', () => {
    fixture.componentRef.setInput('fabrics', [mockFabric]);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    // cy-container should exist (not hidden)
    const container = el.querySelector('.cy-container');
    expect(container).toBeTruthy();
    expect(container?.classList.contains('cy-hidden')).toBe(false);
  });

  it('should render a chip per fabric in toolbar', () => {
    fixture.componentRef.setInput('fabrics', [mockFabric]);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const chips = el.querySelectorAll('.fabric-chip');
    expect(chips.length).toBe(1);
  });

  it('should render legend items', () => {
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.graph-legend')).toBeTruthy();
    expect(el.querySelector('.frontend-dot')).toBeTruthy();
    expect(el.querySelector('.backend-dot')).toBeTruthy();
  });

  it('should show performance warning when nodeWarning is set', () => {
    component.nodeWarning.set('Warning: 600 nodes');
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.node-warning')).toBeTruthy();
    expect(el.textContent).toContain('Warning: 600 nodes');
  });

  it('should not show warning when nodeWarning is null', () => {
    component.nodeWarning.set(null);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.node-warning')).toBeNull();
  });

  it('toggleCollapse should update collapseState for a fabric', () => {
    fixture.componentRef.setInput('fabrics', [mockFabric]);
    fixture.detectChanges();

    // Initially collapsed (default true)
    expect(component.collapseState().get(1) ?? true).toBe(true);

    component.toggleCollapse(1);
    expect(component.collapseState().get(1)).toBe(false);

    component.toggleCollapse(1);
    expect(component.collapseState().get(1)).toBe(true);
  });

  it('should emit nodeClick when nodeClick called from child', () => {
    const emitted: NodeClickEvent[] = [];
    component.nodeClick.subscribe(e => emitted.push(e));

    const mockNodeData: NodeData = {
      id: 'f1-leaf-group',
      label: 'Leaf ×2',
      role: 'leaf',
      tier: 'frontend',
      isGroup: true,
      index: 0,
      count: 2,
      fabricId: 1,
      fabricName: 'test-fabric',
    };

    component.nodeClick.emit({ data: mockNodeData });
    expect(emitted.length).toBe(1);
    expect(emitted[0].data.role).toBe('leaf');
  });

  it('should emit edgeClick when edgeClick called', () => {
    const emitted: EdgeClickEvent[] = [];
    component.edgeClick.subscribe(e => emitted.push(e));

    const mockEdgeData: EdgeData = {
      id: 'e1',
      source: 'f1-leaf-group',
      target: 'f1-spine-group',
      fabricId: 1,
    };

    component.edgeClick.emit({ data: mockEdgeData });
    expect(emitted.length).toBe(1);
    expect(emitted[0].data.source).toBe('f1-leaf-group');
  });

  it('should apply frontend-chip class for frontend tier fabric', () => {
    fixture.componentRef.setInput('fabrics', [mockFabric]);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const chip = el.querySelector('.fabric-chip');
    expect(chip?.classList.contains('frontend-chip')).toBe(true);
  });

  it('should apply backend-chip class for backend tier fabric', () => {
    fixture.componentRef.setInput('fabrics', [{ ...mockFabric, tier: 'backend' }]);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const chip = el.querySelector('.fabric-chip');
    expect(chip?.classList.contains('backend-chip')).toBe(true);
  });

  it('fit button should be present in toolbar', () => {
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const fitBtn = el.querySelector('[aria-label="Fit graph to view"]');
    expect(fitBtn).toBeTruthy();
  });

  it('cy-container should have accessibility label', () => {
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const container = el.querySelector('.cy-container');
    expect(container?.getAttribute('aria-label')).toBeTruthy();
  });
});
