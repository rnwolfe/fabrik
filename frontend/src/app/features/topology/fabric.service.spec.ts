import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { FabricService } from './fabric.service';
import { FabricResponse, TopologyPlan } from '../../models/fabric';

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

describe('FabricService', () => {
  let service: FabricService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        FabricService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(FabricService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('listFabrics sends GET /api/fabrics', () => {
    service.listFabrics().subscribe(fabrics => {
      expect(fabrics.length).toBe(1);
      expect(fabrics[0].name).toBe('test-fabric');
    });

    const req = httpMock.expectOne('/api/fabrics');
    expect(req.request.method).toBe('GET');
    req.flush([mockFabric]);
  });

  it('getFabric sends GET /api/fabrics/:id', () => {
    service.getFabric(1).subscribe(fabric => {
      expect(fabric.id).toBe(1);
    });

    const req = httpMock.expectOne('/api/fabrics/1');
    expect(req.request.method).toBe('GET');
    req.flush(mockFabric);
  });

  it('createFabric sends POST /api/fabrics', () => {
    const createReq = {
      design_id: 0,
      name: 'new-fabric',
      tier: 'frontend' as const,
      stages: 2,
      radix: 64,
      oversubscription: 1.0,
    };

    service.createFabric(createReq).subscribe(fabric => {
      expect(fabric.name).toBe('test-fabric');
    });

    const req = httpMock.expectOne('/api/fabrics');
    expect(req.request.method).toBe('POST');
    expect(req.request.body.name).toBe('new-fabric');
    req.flush(mockFabric);
  });

  it('updateFabric sends PUT /api/fabrics/:id', () => {
    const updateReq = {
      name: 'updated-fabric',
      tier: 'backend' as const,
      stages: 3,
      radix: 48,
      oversubscription: 2.0,
    };

    service.updateFabric(1, updateReq).subscribe(fabric => {
      expect(fabric.id).toBe(1);
    });

    const req = httpMock.expectOne('/api/fabrics/1');
    expect(req.request.method).toBe('PUT');
    req.flush(mockFabric);
  });

  it('deleteFabric sends DELETE /api/fabrics/:id', () => {
    service.deleteFabric(1).subscribe();

    const req = httpMock.expectOne('/api/fabrics/1');
    expect(req.request.method).toBe('DELETE');
    req.flush(null);
  });

  it('previewTopology sends POST /api/fabrics/preview', () => {
    service.previewTopology({ stages: 2, radix: 64, oversubscription: 1.0 }).subscribe(plan => {
      expect(plan.stages).toBe(2);
      expect(plan.total_switches).toBe(33);
    });

    const req = httpMock.expectOne('/api/fabrics/preview');
    expect(req.request.method).toBe('POST');
    expect(req.request.body.stages).toBe(2);
    req.flush(mockPlan);
  });
});
