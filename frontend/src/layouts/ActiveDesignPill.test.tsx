/// <reference types="vitest/globals" />
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { DesignProvider } from '@/contexts/DesignContext';
import { ActiveDesignPill } from './ActiveDesignPill';
import type { Design } from '@/models';

const mockDesigns: Design[] = [
  { id: 1, name: 'Design Alpha', description: '', created_at: '', updated_at: '' },
  { id: 2, name: 'Design Beta', description: '', created_at: '', updated_at: '' },
  {
    id: 3,
    name: 'A Very Long Design Name That Should Be Truncated In The Pill',
    description: '',
    created_at: '',
    updated_at: '',
  },
];

const mockList = vi.fn(() => Promise.resolve(mockDesigns));

vi.mock('@/api/designs', () => ({
  designsApi: {
    list: () => mockList(),
  },
}));

function makeWrapper(initialPath = '/design') {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter initialEntries={[initialPath]}>
        <QueryClientProvider client={client}>
          <DesignProvider>{children}</DesignProvider>
        </QueryClientProvider>
      </MemoryRouter>
    );
  };
}

beforeEach(() => {
  localStorage.clear();
  vi.clearAllMocks();
  mockList.mockResolvedValue(mockDesigns);
});

describe('ActiveDesignPill', () => {
  it('renders a loading skeleton while designs are fetching', () => {
    // Never resolves during this test
    mockList.mockReturnValue(new Promise(() => {}));
    const { container } = render(<ActiveDesignPill />, { wrapper: makeWrapper() });
    // The skeleton should be present (a div with animate-pulse styles)
    expect(container.querySelector('.animate-pulse')).toBeTruthy();
  });

  it('shows "Select design…" when no design is active', async () => {
    render(<ActiveDesignPill />, { wrapper: makeWrapper() });
    await waitFor(() => {
      expect(screen.getByText('Select design…')).toBeInTheDocument();
    });
  });

  it('shows active design name when a design is active', async () => {
    localStorage.setItem('fabrik:activeDesignId', '1');
    render(<ActiveDesignPill />, { wrapper: makeWrapper() });
    await waitFor(() => {
      expect(screen.getByText('Design Alpha')).toBeInTheDocument();
    });
  });

  it('truncates long design names in the pill', async () => {
    localStorage.setItem('fabrik:activeDesignId', '3');
    render(<ActiveDesignPill />, { wrapper: makeWrapper() });
    await waitFor(() => {
      // Truncated text ends with ellipsis
      const truncated = screen.getByText(/A Very Long Design Name…/);
      expect(truncated.textContent!.length).toBeLessThan(mockDesigns[2].name.length);
    });
  });

  it('opens popover and shows recent designs + "View all designs" on click', async () => {
    // Set recent history
    localStorage.setItem('fabrik:recentDesignIds', JSON.stringify([2, 1]));
    localStorage.setItem('fabrik:activeDesignId', '1');

    render(<ActiveDesignPill />, { wrapper: makeWrapper() });

    // Wait for designs to load
    await waitFor(() => screen.getByText('Design Alpha'));

    // Open popover — trigger is a div[role=button]
    const trigger = screen.getByRole('button', { name: /Active design/i });
    fireEvent.click(trigger);

    await waitFor(() => {
      // Design Beta should appear in recent list (id 2 is recent but not active)
      expect(screen.getByText('Design Beta')).toBeInTheDocument();
      expect(screen.getByText('View all designs')).toBeInTheDocument();
    });
  });

  it('selecting a design from popover updates the context', async () => {
    localStorage.setItem('fabrik:recentDesignIds', JSON.stringify([2]));

    render(<ActiveDesignPill />, { wrapper: makeWrapper() });
    await waitFor(() => screen.getByText('Select design…'));

    // Open popover — trigger is a div[role=button]
    const trigger = screen.getByRole('button', { name: /Select design/i });
    fireEvent.click(trigger);

    await waitFor(() => screen.getByText('Design Beta'));

    fireEvent.click(screen.getByText('Design Beta'));

    await waitFor(() => {
      expect(localStorage.getItem('fabrik:activeDesignId')).toBe('2');
    });
  });
});
