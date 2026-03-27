import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { BlockService } from './block.service';
import { Block, BlockAggregationSummary, PortConnection, AddRackToBlockResult } from '../../models';

const mockBlock: Block = {
  id: 1,
  super_block_id: 10,
  name: 'row-A',
  description: '',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

const mockSummary: BlockAggregationSummary = {
  id: 1,
  block_id: 1,
  plane: 'front_end',
  device_model_id: 5,
  total_ports: 32,
  allocated_ports: 2,
  available_ports: 30,
  utilization: '2/32 ports allocated on front_end agg',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('BlockService', () => {
  let service: BlockService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [BlockService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(BlockService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  describe('createBlock', () => {
    it('should POST /api/blocks', () => {
      let result: Block | undefined;
      service.createBlock({ super_block_id: 10, name: 'row-A' }).subscribe(b => { result = b; });
      const req = httpMock.expectOne('/api/blocks');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ super_block_id: 10, name: 'row-A' });
      req.flush(mockBlock);
      expect(result?.name).toBe('row-A');
    });
  });

  describe('getBlock', () => {
    it('should GET /api/blocks/:id', () => {
      let result: Block | undefined;
      service.getBlock(1).subscribe(b => { result = b; });
      const req = httpMock.expectOne('/api/blocks/1');
      expect(req.request.method).toBe('GET');
      req.flush(mockBlock);
      expect(result?.id).toBe(1);
    });
  });

  describe('listBlocks', () => {
    it('should GET /api/blocks?super_block_id=10', () => {
      let result: Block[] | undefined;
      service.listBlocks(10).subscribe(bs => { result = bs; });
      const req = httpMock.expectOne(r => r.url === '/api/blocks' && r.params.get('super_block_id') === '10');
      expect(req.request.method).toBe('GET');
      req.flush([mockBlock]);
      expect(result?.length).toBe(1);
    });
  });

  describe('assignAggregation', () => {
    it('should PUT /api/blocks/:id/aggregations/:plane', () => {
      let result: BlockAggregationSummary | undefined;
      service.assignAggregation(1, 'front_end', { device_model_id: 5 }).subscribe(s => { result = s; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations/front_end');
      expect(req.request.method).toBe('PUT');
      expect(req.request.body).toEqual({ device_model_id: 5 });
      req.flush(mockSummary);
      expect(result?.total_ports).toBe(32);
      expect(result?.allocated_ports).toBe(2);
    });
  });

  describe('getAggregation', () => {
    it('should GET /api/blocks/:id/aggregations/:plane', () => {
      let result: BlockAggregationSummary | undefined;
      service.getAggregation(1, 'front_end').subscribe(s => { result = s; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations/front_end');
      expect(req.request.method).toBe('GET');
      req.flush(mockSummary);
      expect(result?.utilization).toBe('2/32 ports allocated on front_end agg');
    });
  });

  describe('listAggregations', () => {
    it('should GET /api/blocks/:id/aggregations', () => {
      let result: BlockAggregationSummary[] | undefined;
      service.listAggregations(1).subscribe(ss => { result = ss; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations');
      expect(req.request.method).toBe('GET');
      req.flush([mockSummary]);
      expect(result?.length).toBe(1);
    });
  });

  describe('deleteAggregation', () => {
    it('should DELETE /api/blocks/:id/aggregations/:plane', () => {
      let called = false;
      service.deleteAggregation(1, 'front_end').subscribe(() => { called = true; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations/front_end');
      expect(req.request.method).toBe('DELETE');
      req.flush(null, { status: 204, statusText: 'No Content' });
      expect(called).toBe(true);
    });
  });

  describe('listPortConnections', () => {
    it('should GET /api/blocks/:id/aggregations/:plane/connections', () => {
      const mockConns: PortConnection[] = [
        { id: 1, block_aggregation_id: 1, rack_id: 2, agg_port_index: 0, leaf_device_name: 'leaf-1', created_at: '' },
      ];
      let result: PortConnection[] | undefined;
      service.listPortConnections(1, 'front_end').subscribe(cs => { result = cs; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations/front_end/connections');
      expect(req.request.method).toBe('GET');
      req.flush(mockConns);
      expect(result?.length).toBe(1);
      expect(result?.[0].leaf_device_name).toBe('leaf-1');
    });
  });

  describe('addRackToBlock', () => {
    it('should POST /api/blocks/add-rack with block_id', () => {
      const mockResult: AddRackToBlockResult = {
        rack: { id: 3, block_id: 1, name: 'r1', height_u: 42, power_capacity_w: 0, description: '', created_at: '', updated_at: '' },
        connections: [],
      };
      let result: AddRackToBlockResult | undefined;
      service.addRackToBlock({ rack_id: 3, block_id: 1 }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/blocks/add-rack');
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual({ rack_id: 3, block_id: 1 });
      req.flush(mockResult);
      expect(result?.rack.id).toBe(3);
      expect(result?.connections.length).toBe(0);
    });

    it('should POST /api/blocks/add-rack with super_block_id for default block', () => {
      const mockResult: AddRackToBlockResult = {
        rack: { id: 4, block_id: 99, name: 'r2', height_u: 42, power_capacity_w: 0, description: '', created_at: '', updated_at: '' },
        connections: [],
        warning: 'no aggregation switch assigned to this block; rack placed without connectivity',
      };
      let result: AddRackToBlockResult | undefined;
      service.addRackToBlock({ rack_id: 4, super_block_id: 10 }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/blocks/add-rack');
      expect(req.request.method).toBe('POST');
      req.flush(mockResult);
      expect(result?.warning).toBeTruthy();
    });
  });

  describe('removeRackFromBlock', () => {
    it('should DELETE /api/blocks/racks/:rack_id', () => {
      let called = false;
      service.removeRackFromBlock(5).subscribe(() => { called = true; });
      const req = httpMock.expectOne('/api/blocks/racks/5');
      expect(req.request.method).toBe('DELETE');
      req.flush(null, { status: 204, statusText: 'No Content' });
      expect(called).toBe(true);
    });
  });

  describe('capacity display states', () => {
    it('should reflect full capacity in warning field', () => {
      const fullSummary: BlockAggregationSummary = {
        ...mockSummary,
        allocated_ports: 32,
        available_ports: 0,
        utilization: '32/32 ports allocated on front_end agg',
        warning: '32/32 ports allocated on front_end agg; no capacity for additional racks',
      };
      let result: BlockAggregationSummary | undefined;
      service.getAggregation(1, 'front_end').subscribe(s => { result = s; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations/front_end');
      req.flush(fullSummary);
      expect(result?.available_ports).toBe(0);
      expect(result?.warning).toBeTruthy();
    });

    it('should reflect empty capacity when no ports allocated', () => {
      const emptySummary: BlockAggregationSummary = {
        ...mockSummary,
        allocated_ports: 0,
        available_ports: 32,
        utilization: '0/32 ports allocated on front_end agg',
        warning: undefined,
      };
      let result: BlockAggregationSummary | undefined;
      service.getAggregation(1, 'front_end').subscribe(s => { result = s; });
      const req = httpMock.expectOne('/api/blocks/1/aggregations/front_end');
      req.flush(emptySummary);
      expect(result?.available_ports).toBe(32);
      expect(result?.warning).toBeFalsy();
    });
  });
});
