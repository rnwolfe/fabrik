import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { DeviceCatalogService } from './device-catalog.service';
import { DeviceModel, DeviceModelPayload } from '../../models/device-model';

const mockModel: DeviceModel = {
  id: 1,
  vendor: 'Cisco',
  model: 'Nexus 9364C-GX2A',
  device_model_type: 'network',
  port_count: 64,
  height_u: 2,
  power_watts_idle: 1500,
  power_watts_typical: 2000,
  power_watts_max: 2500,
  cpu_sockets: 0,
  cores_per_socket: 0,
  ram_gb: 0,
  storage_tb: 0,
  gpu_count: 0,
  description: '64x 400GbE QSFP-DD spine switch',
  is_seed: true,
  archived_at: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('DeviceCatalogService', () => {
  let service: DeviceCatalogService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(DeviceCatalogService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should list device models', () => {
    let result: DeviceModel[] | undefined;
    service.list().subscribe(r => (result = r));

    const req = httpMock.expectOne('/api/catalog/devices');
    expect(req.request.method).toBe('GET');
    expect(req.request.params.has('include_archived')).toBe(false);
    req.flush([mockModel]);
    expect(result).toEqual([mockModel]);
  });

  it('should include include_archived param when requested', () => {
    service.list(true).subscribe();
    const req = httpMock.expectOne(r => r.url === '/api/catalog/devices' && r.params.get('include_archived') === 'true');
    req.flush([]);
  });

  it('should get a single device model', () => {
    let result: DeviceModel | undefined;
    service.get(1).subscribe(r => (result = r));

    const req = httpMock.expectOne('/api/catalog/devices/1');
    expect(req.request.method).toBe('GET');
    req.flush(mockModel);
    expect(result).toEqual(mockModel);
  });

  it('should create a device model', () => {
    const payload: DeviceModelPayload = {
      vendor: 'Dell',
      model: 'PowerEdge R750',
      device_model_type: 'server',
      port_count: 0,
      height_u: 1,
      power_watts_idle: 0,
      power_watts_typical: 800,
      power_watts_max: 0,
      cpu_sockets: 0,
      cores_per_socket: 0,
      ram_gb: 0,
      storage_tb: 0,
      gpu_count: 0,
      description: '',
    };
    let result: DeviceModel | undefined;
    service.create(payload).subscribe(r => (result = r));

    const req = httpMock.expectOne('/api/catalog/devices');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual(payload);
    req.flush({ ...mockModel, id: 2, ...payload });
    expect(result?.vendor).toBe('Dell');
  });

  it('should update a device model', () => {
    const payload: DeviceModelPayload = {
      vendor: 'Dell',
      model: 'Updated',
      device_model_type: 'server',
      port_count: 0,
      height_u: 1,
      power_watts_idle: 0,
      power_watts_typical: 900,
      power_watts_max: 0,
      cpu_sockets: 0,
      cores_per_socket: 0,
      ram_gb: 0,
      storage_tb: 0,
      gpu_count: 0,
      description: '',
    };
    service.update(2, payload).subscribe();
    const req = httpMock.expectOne('/api/catalog/devices/2');
    expect(req.request.method).toBe('PUT');
    req.flush({ ...mockModel, id: 2, ...payload });
  });

  it('should archive a device model', () => {
    service.archive(1).subscribe();
    const req = httpMock.expectOne('/api/catalog/devices/1');
    expect(req.request.method).toBe('DELETE');
    req.flush(null, { status: 204, statusText: 'No Content' });
  });

  it('should duplicate a device model', () => {
    let result: DeviceModel | undefined;
    service.duplicate(1).subscribe(r => (result = r));

    const req = httpMock.expectOne('/api/catalog/devices/1/duplicate');
    expect(req.request.method).toBe('POST');
    req.flush({ ...mockModel, id: 99, model: 'Nexus 9364C-GX2A (copy)', is_seed: false });
    expect(result?.id).toBe(99);
    expect(result?.is_seed).toBe(false);
  });
});
