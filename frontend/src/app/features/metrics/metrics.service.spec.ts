import { TestBed } from '@angular/core/testing';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { MetricsService } from './metrics.service';
import { DesignMetrics } from '../../models';

const mockMetrics: DesignMetrics = {
  design_id: 1,
  total_hosts: 24,
  total_switches: 9,
  bisection_bandwidth_gbps: 0,
  fabrics: [],
  power: { total_capacity_w: 0, total_draw_w: 0, utilization_pct: 0 },
  capacity: { total_vcpu: 0, total_ram_gb: 0, total_storage_tb: 0, total_gpu_count: 0 },
  port_utilization: [],
  empty: false,
};

describe('MetricsService', () => {
  let svc: MetricsService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    svc = TestBed.inject(MetricsService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should GET /api/designs/1/metrics', () => {
    let result: DesignMetrics | undefined;
    svc.getDesignMetrics(1).subscribe(m => (result = m));

    const req = httpMock.expectOne('/api/designs/1/metrics');
    expect(req.request.method).toBe('GET');
    req.flush(mockMetrics);

    expect(result).toEqual(mockMetrics);
  });
});
