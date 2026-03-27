import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideRouter } from '@angular/router';

import { TopologyDetailPanelComponent } from './topology-detail-panel.component';
import { NodeData, EdgeData } from './topology-graph.service';

const mockNodeData: NodeData = {
  id: 'f1-leaf-group',
  label: 'Leaf ×2',
  role: 'leaf',
  tier: 'frontend',
  isGroup: true,
  index: 0,
  count: 2,
  utilization: 0.62,
  utilLevel: 'healthy',
  fabricId: 1,
  fabricName: 'prod-fabric',
};

const mockEdgeData: EdgeData = {
  id: 'e1',
  source: 'f1-leaf-group',
  target: 'f1-spine-group',
  fabricId: 1,
};

describe('TopologyDetailPanelComponent', () => {
  let component: TopologyDetailPanelComponent;
  let fixture: ComponentFixture<TopologyDetailPanelComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TopologyDetailPanelComponent],
      providers: [provideAnimations(), provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(TopologyDetailPanelComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should show empty state when item is null', () => {
    component.item = null;
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.detail-empty')).toBeTruthy();
  });

  it('should show node details when item kind is node', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('prod-fabric');
    expect(el.textContent).toContain('Leaf');
  });

  it('should show fabric name for node', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('prod-fabric');
  });

  it('should show collapsed count when node isGroup', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('2');
    expect(el.textContent).toContain('collapsed');
  });

  it('should show utilization percentage', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('62%');
  });

  it('should show edge source and target', () => {
    fixture.componentRef.setInput('item', { kind: 'edge', data: mockEdgeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('f1-leaf-group');
    expect(el.textContent).toContain('f1-spine-group');
  });

  it('should emit closed event when close is called', () => {
    let closed = false;
    component.closed.subscribe(() => (closed = true));
    component.close();
    expect(closed).toBe(true);
  });

  it('should have close button in header', () => {
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const closeBtn = el.querySelector('[aria-label="Close detail panel"]');
    expect(closeBtn).toBeTruthy();
  });

  it('should have knowledge base help link', () => {
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const helpLink = el.querySelector('[aria-label*="knowledge"]');
    expect(helpLink).toBeTruthy();
  });

  it('utilPercent should return formatted string', () => {
    expect(component.utilPercent(0.75)).toBe('75%');
    expect(component.utilPercent(0)).toBe('0%');
    expect(component.utilPercent(undefined)).toBe('N/A');
  });

  it('utilClass should return correct class names', () => {
    expect(component.utilClass(0.5)).toBe('util-healthy');
    expect(component.utilClass(0.75)).toBe('util-warning');
    expect(component.utilClass(0.95)).toBe('util-critical');
    expect(component.utilClass(undefined)).toBe('');
  });

  it('nodeData getter returns data for node item', () => {
    component.item = { kind: 'node', data: mockNodeData };
    expect(component.nodeData).toEqual(mockNodeData);
  });

  it('nodeData getter returns null for edge item', () => {
    component.item = { kind: 'edge', data: mockEdgeData };
    expect(component.nodeData).toBeNull();
  });

  it('edgeData getter returns data for edge item', () => {
    component.item = { kind: 'edge', data: mockEdgeData };
    expect(component.edgeData).toEqual(mockEdgeData);
  });

  it('edgeData getter returns null for node item', () => {
    component.item = { kind: 'node', data: mockNodeData };
    expect(component.edgeData).toBeNull();
  });

  it('header title shows Device Details for node', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('Device Details');
  });

  it('header title shows Link Details for edge', () => {
    fixture.componentRef.setInput('item', { kind: 'edge', data: mockEdgeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('Link Details');
  });

  it('should show utilization bar for node with utilization', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.querySelector('.util-bar-container')).toBeTruthy();
    expect(el.querySelector('.util-bar')).toBeTruthy();
  });

  it('front-end tier label should display correctly', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: mockNodeData });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('Front-End');
  });

  it('back-end tier label should display correctly', () => {
    fixture.componentRef.setInput('item', { kind: 'node', data: { ...mockNodeData, tier: 'backend' } });
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('Back-End');
  });
});
