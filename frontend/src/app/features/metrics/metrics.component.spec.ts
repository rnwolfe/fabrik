import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { of, throwError } from 'rxjs';

import { MetricsComponent } from './metrics.component';
import { MetricsService } from './metrics.service';
import { DesignMetrics, Design } from '../../models';

const mockDesigns: Design[] = [
  { id: 1, name: 'design-alpha', description: '', created_at: '', updated_at: '' },
  { id: 2, name: 'design-beta', description: '', created_at: '', updated_at: '' },
];

const mockMetrics: DesignMetrics = {
  design_id: 1,
  total_hosts: 24,
  total_switches: 9,
  bisection_bandwidth_gbps: 0,
  fabrics: [
    {
      fabric_id: 1,
      fabric_name: 'front-end',
      tier: 'frontend',
      stages: 2,
      leaf_spine_oversubscription: 3.0,
      spine_super_spine_oversubscription: 0,
      total_switches: 9,
      total_host_ports: 24,
    },
  ],
  choke_point: {
    fabric_id: 1,
    fabric_name: 'front-end',
    tier: 'leaf→spine',
    ratio: 3.0,
  },
  power: {
    total_capacity_w: 10000,
    total_draw_w: 5000,
    utilization_pct: 50,
  },
  capacity: {
    total_vcpu: 128,
    total_ram_gb: 512,
    total_storage_tb: 10,
    total_gpu_count: 0,
  },
  port_utilization: [
    {
      fabric_id: 1,
      fabric_name: 'front-end',
      tier_name: 'leaf',
      total_ports: 32,
      allocated_ports: 32,
      available_ports: 0,
    },
  ],
  empty: false,
};

describe('MetricsComponent', () => {
  let fixture: ComponentFixture<MetricsComponent>;
  let httpMock: HttpTestingController;
  let mockGetDesignMetrics: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    mockGetDesignMetrics = vi.fn().mockReturnValue(of(mockMetrics));
    const mockMetricsSvc = { getDesignMetrics: mockGetDesignMetrics };

    await TestBed.configureTestingModule({
      imports: [MetricsComponent, NoopAnimationsModule],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: MetricsService, useValue: mockMetricsSvc },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(MetricsComponent);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  function init(): void {
    fixture.detectChanges(); // triggers ngOnInit
    const designsReq = httpMock.expectOne('/api/designs');
    designsReq.flush(mockDesigns);
    fixture.detectChanges();
    fixture.detectChanges();
  }

  it('should display metrics after loading', () => {
    init();
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('24'); // total_hosts
    expect(text).toContain('9');  // total_switches
    expect(text).toContain('50'); // power utilization
  });

  it('should display choke point warning when present', () => {
    init();
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Chokepoint Detected');
    expect(text).toContain('front-end');
  });

  it('should display oversubscription table', () => {
    init();
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Oversubscription by Fabric');
    expect(text).toContain('3.0:1');
  });

  it('should display power metrics', () => {
    init();
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Power');
    expect(text).toContain('10,000'); // capacity
    expect(text).toContain('5,000');  // draw
  });

  it('should display resource capacity', () => {
    init();
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('128'); // vCPU
    expect(text).toContain('512'); // RAM
  });

  it('should display port utilization', () => {
    init();
    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Port Utilization');
    expect(text).toContain('leaf');
  });

  it('should show empty state when no designs exist', () => {
    fixture.detectChanges();
    const designsReq = httpMock.expectOne('/api/designs');
    designsReq.flush([]);
    fixture.detectChanges();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('No designs yet');
  });

  it('should show empty state when design has no devices', () => {
    mockGetDesignMetrics.mockReturnValue(of({ ...mockMetrics, empty: true }));
    fixture.detectChanges();
    const designsReq = httpMock.expectOne('/api/designs');
    designsReq.flush(mockDesigns);
    fixture.detectChanges();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('No devices in this design');
  });

  it('should show error message on metrics API failure', () => {
    mockGetDesignMetrics.mockReturnValue(throwError(() => new Error('network error')));
    fixture.detectChanges();
    const designsReq = httpMock.expectOne('/api/designs');
    designsReq.flush(mockDesigns);
    fixture.detectChanges();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Failed to load metrics');
  });

  it('should reload metrics when design changes', () => {
    init();
    fixture.componentInstance.onDesignChange(2);
    fixture.detectChanges();
    fixture.detectChanges();

    expect(mockGetDesignMetrics).toHaveBeenCalledWith(2);
    expect(fixture.componentInstance.selectedDesignId()).toBe(2);
  });

  it('should format oversubscription ratio as N:1', () => {
    const comp = fixture.componentInstance;
    expect(comp.formatOversub(3.0)).toBe('3.0:1');
    expect(comp.formatOversub(1.5)).toBe('1.5:1');
  });

  it('should classify oversubscription correctly', () => {
    const comp = fixture.componentInstance;
    expect(comp.oversubClass(1.0)).toBe('metric-good');
    expect(comp.oversubClass(2.0)).toBe('metric-warn');
    expect(comp.oversubClass(4.0)).toBe('metric-bad');
  });

  it('should compute port utilization percentage', () => {
    const comp = fixture.componentInstance;
    expect(comp.portUtilizationPct(32, 24)).toBe(75);
    expect(comp.portUtilizationPct(0, 0)).toBe(0);
    expect(comp.portUtilizationPct(32, 32)).toBe(100);
  });
});
