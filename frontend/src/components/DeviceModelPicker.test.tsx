import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import DeviceModelPicker from './DeviceModelPicker';
import { catalogApi } from '@/api/catalog';
import type { DeviceModel } from '@/models';

// Mock catalog API (used by CreateDeviceSheet inside picker)
vi.mock('@/api/catalog', () => ({
  catalogApi: {
    create: vi.fn(),
    list: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    duplicate: vi.fn(),
  },
}));

// Mock @base-ui/react/popover — render children always visible for tests
vi.mock('@base-ui/react/popover', () => {
  const Root = ({ children }: { children: React.ReactNode }) => <div>{children}</div>;
  const Trigger = ({ children, render: renderProp }: { children?: React.ReactNode; render?: React.ReactElement }) => {
    if (renderProp) {
      return <button data-testid="picker-trigger">{children}</button>;
    }
    return <button data-testid="picker-trigger">{children}</button>;
  };
  const Portal = ({ children }: { children: React.ReactNode }) => <>{children}</>;
  const Positioner = ({ children }: { children: React.ReactNode }) => <div>{children}</div>;
  const Popup = ({ children }: { children: React.ReactNode }) => <div data-testid="picker-popup">{children}</div>;
  return { Popover: { Root, Trigger, Portal, Positioner, Popup } };
});

// Mock @base-ui/react/dialog used by Sheet (CreateDeviceSheet)
vi.mock('@base-ui/react/dialog', () => {
  const Root = ({ children, open }: { children: React.ReactNode; open?: boolean; onOpenChange?: (v: boolean) => void }) =>
    open ? <div data-testid="sheet-root">{children}</div> : null;
  const Portal = ({ children }: { children: React.ReactNode }) => <>{children}</>;
  const Backdrop = () => null;
  const Popup = ({ children }: { children: React.ReactNode }) => (
    <div data-testid="sheet-popup">{children}</div>
  );
  const Title = ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>;
  const Description = ({ children }: { children: React.ReactNode }) => <p>{children}</p>;
  const Close = ({ children }: { children: React.ReactNode }) => <>{children}</>;
  const Trigger = ({ children }: { children: React.ReactNode }) => <>{children}</>;
  return { Dialog: { Root, Portal, Backdrop, Popup, Title, Description, Close, Trigger } };
});

const mockDevices: DeviceModel[] = [
  {
    id: 1,
    vendor: 'Arista',
    model: 'DCS-7050CX3',
    device_model_type: 'network',
    port_count: 32,
    height_u: 1,
    power_watts_idle: 100,
    power_watts_typical: 150,
    power_watts_max: 200,
    cpu_sockets: 0,
    cores_per_socket: 0,
    ram_gb: 0,
    storage_tb: 0,
    gpu_count: 0,
    description: 'Arista leaf switch',
    is_seed: false,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    vendor: 'Cisco',
    model: 'Nexus 9336C',
    device_model_type: 'network',
    port_count: 36,
    height_u: 1,
    power_watts_idle: 120,
    power_watts_typical: 180,
    power_watts_max: 250,
    cpu_sockets: 0,
    cores_per_socket: 0,
    ram_gb: 0,
    storage_tb: 0,
    gpu_count: 0,
    description: 'Cisco spine switch',
    is_seed: false,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
];

const newDevice: DeviceModel = {
  id: 99,
  vendor: 'Juniper',
  model: 'QFX5100',
  device_model_type: 'network',
  port_count: 48,
  height_u: 1,
  power_watts_idle: 80,
  power_watts_typical: 120,
  power_watts_max: 180,
  cpu_sockets: 0,
  cores_per_socket: 0,
  ram_gb: 0,
  storage_tb: 0,
  gpu_count: 0,
  description: 'Juniper leaf switch',
  is_seed: false,
  archived_at: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

function makeQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function renderPicker(onSelect = vi.fn()) {
  const qc = makeQueryClient();
  const result = render(
    <QueryClientProvider client={qc}>
      <DeviceModelPicker devices={mockDevices} onSelect={onSelect} />
    </QueryClientProvider>
  );
  return { ...result, onSelect, qc };
}

describe('DeviceModelPicker', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders device list', () => {
    renderPicker();
    expect(screen.getByText('Arista DCS-7050CX3')).toBeInTheDocument();
    expect(screen.getByText('Cisco Nexus 9336C')).toBeInTheDocument();
  });

  it('shows "Create new device…" action in the list', () => {
    renderPicker();
    expect(screen.getByText('Create new device…')).toBeInTheDocument();
  });

  it('calls onSelect when a device is clicked', () => {
    const onSelect = vi.fn();
    renderPicker(onSelect);
    fireEvent.click(screen.getByText('Arista DCS-7050CX3'));
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it('opens CreateDeviceSheet when "Create new device…" is clicked', () => {
    renderPicker();
    expect(screen.queryByText('Create New Device')).not.toBeInTheDocument();
    fireEvent.click(screen.getByText('Create new device…'));
    expect(screen.getByText('Create New Device')).toBeInTheDocument();
  });

  it('auto-selects the new device after successful creation and closes the sheet', async () => {
    vi.mocked(catalogApi.create).mockResolvedValue(newDevice);
    const onSelect = vi.fn();
    renderPicker(onSelect);

    // Open sheet
    fireEvent.click(screen.getByText('Create new device…'));
    expect(screen.getByText('Create New Device')).toBeInTheDocument();

    // Fill in and submit the form
    fireEvent.change(screen.getByLabelText('Vendor'), { target: { value: 'Juniper' } });
    fireEvent.change(screen.getByLabelText('Model'), { target: { value: 'QFX5100' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create Device' }));

    await waitFor(() => {
      expect(catalogApi.create).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(onSelect).toHaveBeenCalledWith(99);
    });
  });
});
