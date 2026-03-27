import { render, screen } from '@testing-library/angular';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { of } from 'rxjs';
import { DevicePickerComponent } from './device-picker.component';
import { DeviceCatalogService } from '../../features/catalog/device-catalog.service';
import { DeviceModel } from '../../models/device-model';

const mockModels: DeviceModel[] = [
  {
    id: 1,
    vendor: 'Cisco',
    model: 'Nexus 9364C-GX2A',
    port_count: 64,
    height_u: 2,
    power_watts: 2000,
    description: '',
    is_seed: true,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    vendor: 'Dell',
    model: 'PowerEdge R750',
    port_count: 0,
    height_u: 1,
    power_watts: 800,
    description: '',
    is_seed: false,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

const mockSvc = {
  list: () => of(mockModels),
};

describe('DevicePickerComponent', () => {
  async function setup() {
    return render(DevicePickerComponent, {
      providers: [
        provideAnimations(),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: DeviceCatalogService, useValue: mockSvc },
      ],
    });
  }

  it('should render the search input', async () => {
    await setup();
    expect(screen.getByRole('textbox', { name: /search device models/i })).toBeTruthy();
  });

  it('should render device models as list options', async () => {
    await setup();
    expect(await screen.findByText(/Nexus 9364C-GX2A/)).toBeTruthy();
    expect(screen.getByText(/PowerEdge R750/)).toBeTruthy();
  });

  it('should filter models by search term', async () => {
    const { fixture } = await setup();
    const comp = fixture.componentInstance;
    comp.onSearchChange('cisco');
    fixture.detectChanges();
    const filtered = comp.filteredModels();
    expect(filtered.length).toBe(1);
    expect(filtered[0].vendor).toBe('Cisco');
  });

  it('should emit deviceSelected when a model is selected', async () => {
    const { fixture } = await setup();
    const comp = fixture.componentInstance;
    const emitted: DeviceModel[] = [];
    comp.deviceSelected.subscribe((dm: DeviceModel) => emitted.push(dm));
    comp.select(mockModels[0]);
    expect(emitted.length).toBe(1);
    expect(emitted[0].id).toBe(1);
  });

  it('should emit deviceDragStart on drag', async () => {
    const { fixture } = await setup();
    const comp = fixture.componentInstance;
    const emitted: DeviceModel[] = [];
    comp.deviceDragStart.subscribe((dm: DeviceModel) => emitted.push(dm));

    const mockEvent = { dataTransfer: { setData: () => false, effectAllowed: '' } } as unknown as DragEvent;
    comp.onDragStart(mockEvent, mockModels[0]);
    expect(emitted.length).toBe(1);
  });

  it('should show empty state when no models match', async () => {
    const { fixture } = await setup();
    const comp = fixture.componentInstance;
    comp.onSearchChange('zzz-no-match-zzz');
    fixture.detectChanges();
    expect(screen.getByText(/no devices found/i)).toBeTruthy();
  });
});
