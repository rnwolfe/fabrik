import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import CreateDeviceSheet from './CreateDeviceSheet';
import { catalogApi } from '@/api/catalog';
import type { DeviceModel } from '@/models';

// Mock the catalog API
vi.mock('@/api/catalog', () => ({
  catalogApi: {
    create: vi.fn(),
    list: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
    duplicate: vi.fn(),
  },
}));

// Mock @base-ui/react/dialog used by Sheet — render children inline when open
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

const mockDevice: DeviceModel = {
  id: 42,
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
  description: 'Test device',
  is_seed: false,
  archived_at: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

function makeQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function renderSheet(props: Partial<Parameters<typeof CreateDeviceSheet>[0]> = {}) {
  const onOpenChange = vi.fn();
  const onCreated = vi.fn();
  const qc = makeQueryClient();

  const result = render(
    <QueryClientProvider client={qc}>
      <CreateDeviceSheet
        open={true}
        onOpenChange={onOpenChange}
        onCreated={onCreated}
        {...props}
      />
    </QueryClientProvider>
  );
  return { ...result, onOpenChange, onCreated, qc };
}

describe('CreateDeviceSheet', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the form when open', () => {
    renderSheet();
    expect(screen.getByText('Create New Device')).toBeInTheDocument();
    expect(screen.getByLabelText('Vendor')).toBeInTheDocument();
    expect(screen.getByLabelText('Model')).toBeInTheDocument();
  });

  it('does not render when closed', () => {
    renderSheet({ open: false });
    expect(screen.queryByText('Create New Device')).not.toBeInTheDocument();
  });

  it('calls onOpenChange(false) when Cancel is clicked', () => {
    const { onOpenChange } = renderSheet();
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('calls catalogApi.create and onCreated on successful submission', async () => {
    vi.mocked(catalogApi.create).mockResolvedValue(mockDevice);
    const { onCreated, onOpenChange } = renderSheet();

    fireEvent.change(screen.getByLabelText('Vendor'), { target: { value: 'Arista' } });
    fireEvent.change(screen.getByLabelText('Model'), { target: { value: 'DCS-7050CX3' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create Device' }));

    await waitFor(() => {
      expect(catalogApi.create).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(onCreated).toHaveBeenCalledWith(mockDevice);
      expect(onOpenChange).toHaveBeenCalledWith(false);
    });
  });

  it('shows validation error inline when vendor is missing', async () => {
    renderSheet();
    // Fill in model but not vendor
    fireEvent.change(screen.getByLabelText('Model'), { target: { value: 'SomeModel' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create Device' }));

    await waitFor(() => {
      expect(screen.getByText('Vendor required')).toBeInTheDocument();
    });
    // API should not have been called
    expect(catalogApi.create).not.toHaveBeenCalled();
  });

  it('shows server-side error inside the sheet and keeps it open', async () => {
    vi.mocked(catalogApi.create).mockRejectedValue(new Error('Duplicate device'));
    const { onOpenChange } = renderSheet();

    fireEvent.change(screen.getByLabelText('Vendor'), { target: { value: 'Arista' } });
    fireEvent.change(screen.getByLabelText('Model'), { target: { value: 'DCS-7050CX3' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create Device' }));

    await waitFor(() => {
      expect(screen.getByText('Duplicate device')).toBeInTheDocument();
    });
    // Sheet should remain open
    expect(onOpenChange).not.toHaveBeenCalledWith(false);
  });
});
