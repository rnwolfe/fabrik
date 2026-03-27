import { render, screen, fireEvent } from '@testing-library/angular';
import { of, throwError } from 'rxjs';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { vi } from 'vitest';

import { RackListComponent } from './rack-list.component';
import { RackService } from '../../rack.service';
import { Rack } from '../../../../models';

const mockRacks: Rack[] = [
  {
    id: 1, name: 'rack-01', height_u: 42, power_capacity_w: 10000,
    block_id: null, rack_type_id: null, description: '', created_at: '', updated_at: '',
  },
  {
    id: 2, name: 'rack-02', height_u: 24, power_capacity_w: 5000,
    block_id: 3, rack_type_id: 1, description: 'Edge rack', created_at: '', updated_at: '',
  },
];

function makeService(racks: Rack[] = mockRacks) {
  return {
    listRacks: vi.fn(() => of(racks)),
    deleteRack: vi.fn(() => of(undefined)),
    createRack: vi.fn(() => of(racks[0])),
  } as unknown as RackService;
}

async function renderComponent(service: RackService) {
  return render(RackListComponent, {
    providers: [
      { provide: RackService, useValue: service },
      provideAnimationsAsync(),
    ],
  });
}

describe('RackListComponent', () => {
  it('should display rack names after load', async () => {
    const svc = makeService();
    await renderComponent(svc);
    expect(screen.getByText('rack-01')).toBeTruthy();
    expect(screen.getByText('rack-02')).toBeTruthy();
  });

  it('should display empty state when no racks', async () => {
    const svc = makeService([]);
    await renderComponent(svc);
    expect(screen.getByText(/No racks yet/i)).toBeTruthy();
  });

  it('should show error message on load failure', async () => {
    const svc = {
      listRacks: vi.fn(() => throwError(() => new Error('network error'))),
      deleteRack: vi.fn(() => of(undefined)),
      createRack: vi.fn(() => of(mockRacks[0])),
    } as unknown as RackService;
    await renderComponent(svc);
    expect(screen.getByRole('alert')).toBeTruthy();
  });

  it('should display rack height in U', async () => {
    const svc = makeService();
    await renderComponent(svc);
    expect(screen.getByText(/42U/)).toBeTruthy();
    expect(screen.getByText(/24U/)).toBeTruthy();
  });

  it('should display power capacity', async () => {
    const svc = makeService();
    await renderComponent(svc);
    expect(screen.getByText(/10000W/)).toBeTruthy();
  });

  it('should call deleteRack when delete confirmed', async () => {
    const svc = makeService();
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    await renderComponent(svc);

    const deleteButtons = screen.getAllByRole('button', { name: /delete rack/i });
    fireEvent.click(deleteButtons[0]);

    expect(svc.deleteRack).toHaveBeenCalledWith(1);
    confirmSpy.mockRestore();
  });

  it('should not call deleteRack when confirmation declined', async () => {
    const svc = makeService();
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);
    await renderComponent(svc);

    const deleteButtons = screen.getAllByRole('button', { name: /delete rack/i });
    fireEvent.click(deleteButtons[0]);

    expect(svc.deleteRack).not.toHaveBeenCalled();
    confirmSpy.mockRestore();
  });

  it('should render Add Rack button', async () => {
    const svc = makeService();
    await renderComponent(svc);
    expect(screen.getByRole('button', { name: /create rack/i })).toBeTruthy();
  });

  it('should reload racks when retry clicked', async () => {
    const svc = {
      listRacks: vi.fn(() => throwError(() => new Error('network error'))),
      deleteRack: vi.fn(() => of(undefined)),
      createRack: vi.fn(() => of(mockRacks[0])),
    } as unknown as RackService;
    await renderComponent(svc);

    (svc.listRacks as ReturnType<typeof vi.fn>).mockReturnValue(of(mockRacks));
    const retryBtn = screen.getByRole('button', { name: /retry/i });
    fireEvent.click(retryBtn);

    expect(svc.listRacks).toHaveBeenCalledTimes(2);
  });
});
