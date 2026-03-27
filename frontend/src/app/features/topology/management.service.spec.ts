import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { ManagementService } from './management.service';
import { BlockAggregation } from '../../models';

const mockAgg: BlockAggregation = {
  id: 1,
  block_id: 10,
  plane: 'management',
  device_model_id: 5,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('ManagementService', () => {
  let service: ManagementService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), ManagementService],
    });
    service = TestBed.inject(ManagementService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('setManagementAgg should PUT to the correct endpoint', () => {
    service.setManagementAgg(10, { device_model_id: 5 }).subscribe(agg => {
      expect(agg).toEqual(mockAgg);
    });
    const req = httpMock.expectOne('/api/blocks/10/management-agg');
    expect(req.request.method).toBe('PUT');
    req.flush(mockAgg);
  });

  it('getManagementAgg should GET from the correct endpoint', () => {
    service.getManagementAgg(10).subscribe(agg => {
      expect(agg).toEqual(mockAgg);
    });
    const req = httpMock.expectOne('/api/blocks/10/management-agg');
    expect(req.request.method).toBe('GET');
    req.flush(mockAgg);
  });

  it('removeManagementAgg should DELETE to the correct endpoint', () => {
    service.removeManagementAgg(10).subscribe();
    const req = httpMock.expectOne('/api/blocks/10/management-agg');
    expect(req.request.method).toBe('DELETE');
    req.flush(null, { status: 204, statusText: 'No Content' });
  });

  it('listBlockAggregations should GET from the correct endpoint', () => {
    service.listBlockAggregations(10).subscribe(aggs => {
      expect(aggs).toEqual([mockAgg]);
    });
    const req = httpMock.expectOne('/api/blocks/10/aggregations');
    expect(req.request.method).toBe('GET');
    req.flush([mockAgg]);
  });
});
