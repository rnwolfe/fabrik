import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';

import { RackService } from './rack.service';
import { Rack, RackSummary, RackTemplate, PlaceDeviceResult } from '../../models';

describe('RackService', () => {
  let service: RackService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [RackService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(RackService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => httpMock.verify());

  describe('listRackTypes', () => {
    it('should GET /api/rack-types', () => {
      const mock: RackTemplate[] = [
        { id: 1, name: '42U', height_u: 42, power_capacity_w: 10000, power_oversub_pct_warn: 100, power_oversub_pct_max: 110, description: '', created_at: '', updated_at: '' },
      ];
      let result: RackTemplate[] | undefined;
      service.listRackTypes().subscribe(rts => { result = rts; });
      const req = httpMock.expectOne('/api/rack-types');
      expect(req.request.method).toBe('GET');
      req.flush(mock);
      expect(result?.length).toBe(1);
      expect(result?.[0].name).toBe('42U');
    });
  });

  describe('createRackType', () => {
    it('should POST /api/rack-types', () => {
      const mock: RackTemplate = { id: 2, name: '24U', height_u: 24, power_capacity_w: 5000, power_oversub_pct_warn: 100, power_oversub_pct_max: 110, description: '', created_at: '', updated_at: '' };
      let result: RackTemplate | undefined;
      service.createRackType({ name: '24U', height_u: 24, power_capacity_w: 5000 }).subscribe(rt => { result = rt; });
      const req = httpMock.expectOne('/api/rack-types');
      expect(req.request.method).toBe('POST');
      req.flush(mock);
      expect(result?.id).toBe(2);
    });
  });

  describe('updateRackType', () => {
    it('should PUT /api/rack-types/:id', () => {
      const mock: RackTemplate = { id: 1, name: 'updated', height_u: 42, power_capacity_w: 10000, power_oversub_pct_warn: 100, power_oversub_pct_max: 110, description: '', created_at: '', updated_at: '' };
      let result: RackTemplate | undefined;
      service.updateRackType(1, { name: 'updated', height_u: 42, power_capacity_w: 10000 }).subscribe(rt => { result = rt; });
      const req = httpMock.expectOne('/api/rack-types/1');
      expect(req.request.method).toBe('PUT');
      req.flush(mock);
      expect(result?.name).toBe('updated');
    });
  });

  describe('deleteRackType', () => {
    it('should DELETE /api/rack-types/:id', () => {
      let called = false;
      service.deleteRackType(1).subscribe(() => { called = true; });
      const req = httpMock.expectOne('/api/rack-types/1');
      expect(req.request.method).toBe('DELETE');
      req.flush(null, { status: 204, statusText: 'No Content' });
      expect(called).toBe(true);
    });
  });

  describe('listRacks', () => {
    it('should GET /api/racks', () => {
      const mock: Rack[] = [
        { id: 1, name: 'rack-01', height_u: 42, power_capacity_w: 5000, power_oversub_pct_warn: 100, power_oversub_pct_max: 110, block_id: null, rack_type_id: null, description: '', created_at: '', updated_at: '' },
      ];
      let result: Rack[] | undefined;
      service.listRacks().subscribe(racks => { result = racks; });
      const req = httpMock.expectOne('/api/racks');
      expect(req.request.method).toBe('GET');
      req.flush(mock);
      expect(result?.length).toBe(1);
    });

    it('should GET /api/racks?block_id=5 when blockId provided', () => {
      service.listRacks(5).subscribe();
      const req = httpMock.expectOne(r => r.url === '/api/racks' && r.params.get('block_id') === '5');
      expect(req.request.method).toBe('GET');
      req.flush([]);
    });
  });

  describe('getRack', () => {
    it('should GET /api/racks/:id with summary', () => {
      const mock: RackSummary = {
        id: 1, name: 'rack-01', height_u: 42, power_capacity_w: 5000, power_oversub_pct_warn: 100, power_oversub_pct_max: 110,
        block_id: null, rack_type_id: null, description: '', created_at: '', updated_at: '',
        used_u: 2, available_u: 40, used_watts_idle: 300, used_watts_typical: 500, used_watts_max: 700, devices: [],
      };
      let result: RackSummary | undefined;
      service.getRack(1).subscribe(summary => { result = summary; });
      const req = httpMock.expectOne('/api/racks/1');
      expect(req.request.method).toBe('GET');
      req.flush(mock);
      expect(result?.used_u).toBe(2);
      expect(result?.available_u).toBe(40);
    });
  });

  describe('createRack', () => {
    it('should POST /api/racks', () => {
      const mock: Rack = { id: 3, name: 'new-rack', height_u: 42, power_capacity_w: 0, power_oversub_pct_warn: 100, power_oversub_pct_max: 110, block_id: null, rack_type_id: null, description: '', created_at: '', updated_at: '' };
      let result: Rack | undefined;
      service.createRack({ name: 'new-rack', height_u: 42 }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/racks');
      expect(req.request.method).toBe('POST');
      req.flush(mock);
      expect(result?.id).toBe(3);
    });
  });

  describe('updateRack', () => {
    it('should PUT /api/racks/:id', () => {
      const mock: Rack = { id: 1, name: 'updated-rack', height_u: 42, power_capacity_w: 5000, power_oversub_pct_warn: 100, power_oversub_pct_max: 110, block_id: null, rack_type_id: null, description: '', created_at: '', updated_at: '' };
      let result: Rack | undefined;
      service.updateRack(1, { name: 'updated-rack' }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/racks/1');
      expect(req.request.method).toBe('PUT');
      req.flush(mock);
      expect(result?.name).toBe('updated-rack');
    });
  });

  describe('deleteRack', () => {
    it('should DELETE /api/racks/:id', () => {
      let called = false;
      service.deleteRack(1).subscribe(() => { called = true; });
      const req = httpMock.expectOne('/api/racks/1');
      expect(req.request.method).toBe('DELETE');
      req.flush(null, { status: 204, statusText: 'No Content' });
      expect(called).toBe(true);
    });
  });

  describe('placeDevice', () => {
    it('should POST /api/racks/:id/devices', () => {
      const mock: PlaceDeviceResult = {
        device: { id: 1, rack_id: 1, device_model_id: 2, name: 'dev', role: 'leaf', position: 1, description: '', created_at: '', updated_at: '' },
      };
      let result: PlaceDeviceResult | undefined;
      service.placeDevice(1, { device_model_id: 2, name: 'dev', position: 1 }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/racks/1/devices');
      expect(req.request.method).toBe('POST');
      req.flush(mock);
      expect(result?.device.id).toBe(1);
    });

    it('should include warning when power threshold exceeded', () => {
      const mock: PlaceDeviceResult = {
        device: { id: 1, rack_id: 1, device_model_id: 2, name: 'dev', role: 'leaf', position: 1, description: '', created_at: '', updated_at: '' },
        warning: 'power utilization at 90% (900W / 1000W)',
      };
      let result: PlaceDeviceResult | undefined;
      service.placeDevice(1, { device_model_id: 2, name: 'dev' }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/racks/1/devices');
      req.flush(mock);
      expect(result?.warning).toBeDefined();
      expect(result?.warning).toContain('power');
    });
  });

  describe('removeDevice', () => {
    it('should DELETE with compact=false by default', () => {
      let called = false;
      service.removeDevice(1, 5).subscribe(() => { called = true; });
      const req = httpMock.expectOne(r => r.url === '/api/racks/1/devices/5' && r.params.get('compact') === 'false');
      expect(req.request.method).toBe('DELETE');
      req.flush(null, { status: 204, statusText: 'No Content' });
      expect(called).toBe(true);
    });

    it('should pass compact=true when specified', () => {
      service.removeDevice(1, 5, true).subscribe();
      const req = httpMock.expectOne(r => r.url === '/api/racks/1/devices/5' && r.params.get('compact') === 'true');
      req.flush(null, { status: 204, statusText: 'No Content' });
    });
  });

  describe('moveDeviceInRack', () => {
    it('should PUT /api/racks/:rackId/devices/:deviceId', () => {
      const mock: PlaceDeviceResult = {
        device: { id: 5, rack_id: 1, device_model_id: 1, name: 'dev', role: 'leaf', position: 5, description: '', created_at: '', updated_at: '' },
      };
      let result: PlaceDeviceResult | undefined;
      service.moveDeviceInRack(1, 5, { position: 5 }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/racks/1/devices/5');
      expect(req.request.method).toBe('PUT');
      req.flush(mock);
      expect(result?.device.position).toBe(5);
    });
  });

  describe('moveDeviceCrossRack', () => {
    it('should PUT /api/racks/:rackId/devices/:deviceId/move', () => {
      const mock: PlaceDeviceResult = {
        device: { id: 5, rack_id: 2, device_model_id: 1, name: 'moved', role: 'spine', position: 3, description: '', created_at: '', updated_at: '' },
      };
      let result: PlaceDeviceResult | undefined;
      service.moveDeviceCrossRack(1, 5, { dest_rack_id: 2, position: 3 }).subscribe(r => { result = r; });
      const req = httpMock.expectOne('/api/racks/1/devices/5/move');
      expect(req.request.method).toBe('PUT');
      req.flush(mock);
      expect(result?.device.rack_id).toBe(2);
    });
  });
});
