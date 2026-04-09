import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import MetricsStrip from './MetricsStrip';

// ── Mocks ────────────────────────────────────────────────────────────────────

vi.mock('@/api/metrics', () => ({
  metricsApi: {
    getDesignMetrics: vi.fn(),
  },
}));

import { metricsApi } from '@/api/metrics';
const mockGetMetrics = vi.mocked(metricsApi.getDesignMetrics);

// ── Helpers ──────────────────────────────────────────────────────────────────

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        refetchOnWindowFocus: false,
        refetchInterval: false,
      },
    },
  });
}

function renderStrip(designId = 1) {
  const qc = makeQueryClient();
  return render(
    <QueryClientProvider client={qc}>
      <MetricsStrip designId={designId} />
    </QueryClientProvider>
  );
}

const emptyMetrics = {
  design_id: 1,
  total_hosts: 0,
  total_switches: 0,
  bisection_bandwidth_gbps: 0,
  fabrics: [],
  power: { total_capacity_w: 0, total_draw_w: 0, utilization_pct: 0 },
  capacity: { total_vcpu: 0, total_ram_gb: 0, total_storage_tb: 0, total_gpu_count: 0 },
  port_utilization: [],
  empty: true,
};

const populatedMetrics = {
  design_id: 1,
  total_hosts: 1024,
  total_switches: 32,
  bisection_bandwidth_gbps: 800.5,
  fabrics: [
    {
      fabric_id: 1,
      fabric_name: 'Fabric A',
      tier: 'frontend' as const,
      stages: 2,
      leaf_spine_oversubscription: 3.0,
      total_switches: 16,
      total_host_ports: 512,
    },
    {
      fabric_id: 2,
      fabric_name: 'Fabric B',
      tier: 'backend' as const,
      stages: 2,
      leaf_spine_oversubscription: 1.5,
      total_switches: 16,
      total_host_ports: 512,
    },
  ],
  choke_point: {
    fabric_id: 1,
    fabric_name: 'Fabric A',
    tier: 'frontend' as const,
    ratio: 3.0,
  },
  power: { total_capacity_w: 100000, total_draw_w: 60000, utilization_pct: 60.0 },
  capacity: { total_vcpu: 512, total_ram_gb: 2048, total_storage_tb: 10.0, total_gpu_count: 0 },
  port_utilization: [
    {
      fabric_id: 1,
      fabric_name: 'Fabric A',
      tier_name: 'leaf',
      total_ports: 200,
      allocated_ports: 120,
      available_ports: 80,
    },
  ],
  empty: false,
};

// ── Tests ────────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  // Default to expanded viewport
  Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1440 });
});

describe('MetricsStrip — loading state', () => {
  it('renders skeleton placeholders while loading', () => {
    // Never resolves — stays loading
    mockGetMetrics.mockReturnValue(new Promise(() => {}));
    renderStrip();
    // Strip container is present
    expect(screen.getByTestId('metrics-strip')).toBeInTheDocument();
    // Skeletons are rendered (animate-pulse elements)
    const skeletons = document.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});

describe('MetricsStrip — error state', () => {
  it('shows error message and retry button on fetch failure', async () => {
    mockGetMetrics.mockRejectedValue(new Error('Network error'));
    renderStrip();
    const retryBtn = await screen.findByRole('button', { name: /retry/i });
    expect(retryBtn).toBeInTheDocument();
    expect(screen.getByText(/failed to load metrics/i)).toBeInTheDocument();
  });

  it('retry button calls refetch', async () => {
    mockGetMetrics.mockRejectedValueOnce(new Error('fail'));
    renderStrip();
    const retryBtn = await screen.findByRole('button', { name: /retry/i });
    await userEvent.click(retryBtn);
    // After click, getDesignMetrics should be called again (at least twice: initial + retry)
    expect(mockGetMetrics).toHaveBeenCalledTimes(2);
  });
});

describe('MetricsStrip — empty state', () => {
  it('shows "Add a block to see metrics" when metrics are empty', async () => {
    mockGetMetrics.mockResolvedValue(emptyMetrics);
    renderStrip();
    expect(await screen.findByText(/add a block to see metrics/i)).toBeInTheDocument();
  });
});

describe('MetricsStrip — populated state', () => {
  it('displays worst oversubscription from choke_point', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    expect(await screen.findByText('3.00:1')).toBeInTheDocument();
  });

  it('displays total host ports', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    expect(await screen.findByText('1,024')).toBeInTheDocument();
  });

  it('displays bisection bandwidth', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    expect(await screen.findByText('800.5 Gbps')).toBeInTheDocument();
  });

  it('displays power utilization percentage', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    // Both power (60.0%) and port util (60.0%) appear — use findAllByText
    const pctElements = await screen.findAllByText('60.0%');
    expect(pctElements.length).toBeGreaterThanOrEqual(1);
  });

  it('displays port utilization percentage', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    // 120 / 200 = 60%
    const utilValues = await screen.findAllByText('60.0%');
    expect(utilValues.length).toBeGreaterThanOrEqual(2); // power + port util
  });

  it('applies amber color class for oversubscription 3.0:1', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    const el = await screen.findByText('3.00:1');
    expect(el.className).toContain('amber');
  });
});

describe('MetricsStrip — collapsible behavior', () => {
  it('renders expanded by default on wide viewport', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    // Should see metrics content (not just the toggle)
    expect(await screen.findByText('800.5 Gbps')).toBeInTheDocument();
  });

  it('collapses when toggle button is clicked', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    // Wait for metrics to load
    await screen.findByText('800.5 Gbps');

    const toggleBtn = screen.getByRole('button', { name: /collapse metrics strip/i });
    await userEvent.click(toggleBtn);

    // Metrics content should be gone
    expect(screen.queryByText('800.5 Gbps')).not.toBeInTheDocument();
  });

  it('persists collapsed state in localStorage', async () => {
    mockGetMetrics.mockResolvedValue(populatedMetrics);
    renderStrip();
    await screen.findByText('800.5 Gbps');

    const toggleBtn = screen.getByRole('button', { name: /collapse metrics strip/i });
    await userEvent.click(toggleBtn);

    expect(localStorage.getItem('fabrik:designMetricsStripCollapsed')).toBe('true');
  });

  it('reads collapsed state from localStorage on mount', () => {
    localStorage.setItem('fabrik:designMetricsStripCollapsed', 'true');
    mockGetMetrics.mockReturnValue(new Promise(() => {}));
    renderStrip();

    // Collapsed from the start — the expand button should be visible
    expect(screen.getByRole('button', { name: /expand metrics strip/i })).toBeInTheDocument();
    // No metrics content visible
    expect(screen.queryByText(/add a block/i)).not.toBeInTheDocument();
  });

  it('defaults to collapsed on narrow viewport', () => {
    Object.defineProperty(window, 'innerWidth', { value: 768, configurable: true });
    localStorage.clear();
    mockGetMetrics.mockReturnValue(new Promise(() => {}));
    renderStrip();
    // Should show "expand" button (collapsed state)
    expect(screen.getByRole('button', { name: /expand metrics strip/i })).toBeInTheDocument();
  });
});
