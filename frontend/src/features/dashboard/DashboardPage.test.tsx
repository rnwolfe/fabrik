/// <reference types="vitest/globals" />
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { DesignProvider } from '@/contexts/DesignContext';
import DashboardPage from './DashboardPage';

// ── Mock dependencies ────────────────────────────────────────────────────────

const mockNavigate = vi.fn();

vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>();
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockDesignsCreate = vi.fn();
const mockDesignsList = vi.fn(() => Promise.resolve([]));

vi.mock('@/api/designs', () => ({
  designsApi: {
    list: () => mockDesignsList(),
    create: (...args: unknown[]) => mockDesignsCreate(...args),
  },
}));

const mockBlankScaffold = vi.fn((_designId: number) => Promise.resolve());
const mockTwoStageScaffold = vi.fn((_designId: number) => Promise.resolve());

vi.mock('./designTemplates', () => ({
  DESIGN_TEMPLATES: [
    {
      id: 'blank',
      name: 'Blank',
      description: 'Start with an empty canvas.',
      scaffold: (designId: number) => mockBlankScaffold(designId),
    },
    {
      id: '2-stage-clos',
      name: '2-stage Clos',
      description: 'Single pod with one leaf–spine block.',
      scaffold: (designId: number) => mockTwoStageScaffold(designId),
    },
    {
      id: '3-stage-clos',
      name: '3-stage Clos',
      description: 'Single pod with super-spine tier.',
      scaffold: vi.fn(() => Promise.resolve()),
    },
    {
      id: 'pod-based',
      name: 'Pod-based fabric',
      description: 'Two independent pods.',
      scaffold: vi.fn(() => Promise.resolve()),
    },
  ],
}));

// ── Test helpers ─────────────────────────────────────────────────────────────

function makeWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <MemoryRouter>
        <QueryClientProvider client={client}>
          <DesignProvider>{children}</DesignProvider>
        </QueryClientProvider>
      </MemoryRouter>
    );
  };
}

async function openDialog() {
  const btn = screen.getByRole('button', { name: /new design/i });
  fireEvent.click(btn);
  // Wait for the dialog to appear
  await waitFor(() => screen.getByRole('dialog'));
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockDesignsList.mockResolvedValue([]);
});

// ── Template visibility ───────────────────────────────────────────────────────

describe('New Design dialog — template section visible on open', () => {
  it('shows "Start from template" label when dialog opens', async () => {
    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();
    expect(screen.getByText('Start from template')).toBeInTheDocument();
  });

  it('shows all 4 template options', async () => {
    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();
    expect(screen.getByText('Blank')).toBeInTheDocument();
    expect(screen.getByText('2-stage Clos')).toBeInTheDocument();
    expect(screen.getByText('3-stage Clos')).toBeInTheDocument();
    expect(screen.getByText('Pod-based fabric')).toBeInTheDocument();
  });

  it('Blank template is selected by default (aria-pressed=true)', async () => {
    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();
    const blankBtn = screen.getByRole('button', { name: /blank/i });
    expect(blankBtn).toHaveAttribute('aria-pressed', 'true');
  });

  it('non-blank templates are not selected by default', async () => {
    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();
    const twoStageBtn = screen.getByRole('button', { name: /2-stage clos/i });
    expect(twoStageBtn).toHaveAttribute('aria-pressed', 'false');
  });
});

// ── Blank template — no scaffold ─────────────────────────────────────────────

describe('Blank template', () => {
  it('does not call scaffold when Blank is selected and form submitted', async () => {
    const createdDesign = { id: 99, name: 'My Design', description: '', created_at: '', updated_at: '' };
    mockDesignsCreate.mockResolvedValue(createdDesign);

    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();

    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'My Design' } });
    fireEvent.click(screen.getByRole('button', { name: /create design/i }));

    await waitFor(() => expect(mockDesignsCreate).toHaveBeenCalledTimes(1));
    expect(mockBlankScaffold).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/design');
  });
});

// ── Non-blank template — scaffold is called ───────────────────────────────────

describe('Non-blank template', () => {
  it('calls scaffold for the selected template after design creation', async () => {
    const createdDesign = { id: 55, name: 'Test', description: '', created_at: '', updated_at: '' };
    mockDesignsCreate.mockResolvedValue(createdDesign);

    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();

    // Select 2-stage Clos
    const twoStageBtn = screen.getByRole('button', { name: /2-stage clos/i });
    fireEvent.click(twoStageBtn);
    expect(twoStageBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'Test' } });
    fireEvent.click(screen.getByRole('button', { name: /create design/i }));

    await waitFor(() => expect(mockTwoStageScaffold).toHaveBeenCalledWith(55));
    expect(mockNavigate).toHaveBeenCalledWith('/design');
  });

  it('navigates to /design even when scaffold throws (partial failure)', async () => {
    const createdDesign = { id: 66, name: 'Partial', description: '', created_at: '', updated_at: '' };
    mockDesignsCreate.mockResolvedValue(createdDesign);
    mockTwoStageScaffold.mockRejectedValue(new Error('API error'));

    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();

    fireEvent.click(screen.getByRole('button', { name: /2-stage clos/i }));
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'Partial' } });
    fireEvent.click(screen.getByRole('button', { name: /create design/i }));

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith('/design'));
  });

  it('shows a toast when scaffold fails', async () => {
    const createdDesign = { id: 77, name: 'Toast', description: '', created_at: '', updated_at: '' };
    mockDesignsCreate.mockResolvedValue(createdDesign);
    mockTwoStageScaffold.mockRejectedValue(new Error('some API error'));

    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();

    fireEvent.click(screen.getByRole('button', { name: /2-stage clos/i }));
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'Toast' } });
    fireEvent.click(screen.getByRole('button', { name: /create design/i }));

    await waitFor(() => {
      expect(screen.getByRole('status')).toBeInTheDocument();
    });
    expect(screen.getByRole('status').textContent).toContain('Template partially applied');
  });

  it('blank scaffold is never called even when another template is selected', async () => {
    const createdDesign = { id: 88, name: 'NoBlank', description: '', created_at: '', updated_at: '' };
    mockDesignsCreate.mockResolvedValue(createdDesign);

    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();

    fireEvent.click(screen.getByRole('button', { name: /2-stage clos/i }));
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: 'NoBlank' } });
    fireEvent.click(screen.getByRole('button', { name: /create design/i }));

    await waitFor(() => expect(mockTwoStageScaffold).toHaveBeenCalled());
    expect(mockBlankScaffold).not.toHaveBeenCalled();
  });
});

// ── Dialog reset ─────────────────────────────────────────────────────────────

describe('Dialog reset on close', () => {
  it('resets template selection to Blank when dialog is closed and re-opened', async () => {
    render(<DashboardPage />, { wrapper: makeWrapper() });
    await openDialog();

    // Select non-blank template
    fireEvent.click(screen.getByRole('button', { name: /2-stage clos/i }));
    expect(screen.getByRole('button', { name: /2-stage clos/i })).toHaveAttribute('aria-pressed', 'true');

    // Close dialog
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

    // Re-open
    await openDialog();
    expect(screen.getByRole('button', { name: /blank/i })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: /2-stage clos/i })).toHaveAttribute('aria-pressed', 'false');
  });
});
