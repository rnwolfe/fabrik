import { render, screen } from '@testing-library/angular';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { of } from 'rxjs';
import { MatDialog } from '@angular/material/dialog';
import { CatalogComponent } from './catalog.component';
import { DeviceCatalogService } from './device-catalog.service';
import { DeviceModel } from '../../models/device-model';

const mockModels: DeviceModel[] = [
  {
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
    description: '64x 400GbE spine switch',
    is_seed: true,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    vendor: 'Dell',
    model: 'PowerEdge R750',
    device_model_type: 'server',
    port_count: 0,
    height_u: 1,
    power_watts_idle: 200,
    power_watts_typical: 800,
    power_watts_max: 1200,
    cpu_sockets: 2,
    cores_per_socket: 24,
    ram_gb: 512,
    storage_tb: 10,
    gpu_count: 0,
    description: '2-socket server',
    is_seed: false,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

const mockSvc = {
  list: () => of(mockModels),
  create: () => of(mockModels[0]),
  update: () => of(mockModels[0]),
  archive: () => of(undefined),
  duplicate: () => of({ ...mockModels[0], id: 99, model: 'Nexus (copy)', is_seed: false }),
};

describe('CatalogComponent', () => {
  async function setup(overrideSvc?: Partial<typeof mockSvc>) {
    return render(CatalogComponent, {
      providers: [
        provideAnimations(),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: DeviceCatalogService, useValue: { ...mockSvc, ...overrideSvc } },
        { provide: MatDialog, useValue: { open: () => ({ afterClosed: () => of(null) }) } },
      ],
    });
  }

  it('should render the page heading', async () => {
    await setup();
    expect(screen.getByRole('heading', { name: /device catalog/i })).toBeTruthy();
  });

  it('should show device models in the table', async () => {
    await setup();
    expect(await screen.findByText('Nexus 9364C-GX2A')).toBeTruthy();
    expect(screen.getByText('PowerEdge R750')).toBeTruthy();
  });

  it('should show Add Device button', async () => {
    await setup();
    expect(screen.getByRole('button', { name: /add device/i })).toBeTruthy();
  });

  it('should show empty state when no models match filter', async () => {
    const { fixture } = await setup();
    const comp = fixture.componentInstance;
    comp.onSearchChange('zzz-no-match-zzz');
    fixture.detectChanges();
    expect(screen.getByText(/no device models found/i)).toBeTruthy();
  });

  it('should show loading spinner while loading', async () => {
    const { fixture } = await setup();
    const comp = fixture.componentInstance;
    comp.loading.set(true);
    fixture.detectChanges();
    expect(document.querySelector('mat-spinner')).toBeTruthy();
  });
});
