/// <reference types="vitest/globals" />
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { DesignProvider, useDesign } from './DesignContext';
import type { Design } from '@/models';

const LS_ACTIVE_KEY = 'fabrik:activeDesignId';
const LS_RECENT_KEY = 'fabrik:recentDesignIds';

// Mock the designs API
const mockDesigns: Design[] = [
  { id: 1, name: 'Design A', description: '', created_at: '', updated_at: '' },
  { id: 2, name: 'Design B', description: '', created_at: '', updated_at: '' },
];

vi.mock('@/api/designs', () => ({
  designsApi: {
    list: vi.fn(() => Promise.resolve(mockDesigns)),
  },
}));

function makeWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>
        <DesignProvider>{children}</DesignProvider>
      </QueryClientProvider>
    );
  };
}

beforeEach(() => {
  localStorage.clear();
  vi.clearAllMocks();
});

describe('DesignContext localStorage hydration', () => {
  it('starts with null when no value in localStorage', () => {
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    expect(result.current.activeDesignId).toBeNull();
  });

  it('hydrates from localStorage when stored ID is valid', async () => {
    localStorage.setItem(LS_ACTIVE_KEY, JSON.stringify(1));
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    // Initial state from localStorage before validation
    expect(result.current.activeDesignId).toBe(1);
    // After designs load, valid ID stays
    await waitFor(() => {
      expect(result.current.activeDesignId).toBe(1);
    });
  });

  it('clears invalid stored ID after designs load', async () => {
    localStorage.setItem(LS_ACTIVE_KEY, JSON.stringify(999));
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    // Initially reads stored value
    expect(result.current.activeDesignId).toBe(999);
    // After designs are fetched, 999 is not in the list → cleared
    await waitFor(() => {
      expect(result.current.activeDesignId).toBeNull();
    });
    expect(localStorage.getItem(LS_ACTIVE_KEY)).toBeNull();
  });

  it('silently clears malformed JSON in localStorage', async () => {
    localStorage.setItem(LS_ACTIVE_KEY, 'not-json{{{');
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    expect(result.current.activeDesignId).toBeNull();
    await waitFor(() => {
      expect(result.current.activeDesignId).toBeNull();
    });
  });

  it('silently clears non-number JSON value', async () => {
    localStorage.setItem(LS_ACTIVE_KEY, JSON.stringify('one'));
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    expect(result.current.activeDesignId).toBeNull();
  });
});

describe('DesignContext setActiveDesignId', () => {
  it('writes to localStorage and updates state', () => {
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    act(() => {
      result.current.setActiveDesignId(2);
    });
    expect(result.current.activeDesignId).toBe(2);
    expect(localStorage.getItem(LS_ACTIVE_KEY)).toBe('2');
  });

  it('removes from localStorage when set to null', () => {
    localStorage.setItem(LS_ACTIVE_KEY, '1');
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    act(() => {
      result.current.setActiveDesignId(null);
    });
    expect(result.current.activeDesignId).toBeNull();
    expect(localStorage.getItem(LS_ACTIVE_KEY)).toBeNull();
  });

  it('updates recent list (MRU-first, capped at 5)', () => {
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    act(() => {
      for (const id of [1, 2, 3, 4, 5, 6]) {
        result.current.setActiveDesignId(id);
      }
    });
    expect(result.current.recentDesignIds).toEqual([6, 5, 4, 3, 2]);
    const stored = JSON.parse(localStorage.getItem(LS_RECENT_KEY) ?? '[]') as number[];
    expect(stored).toEqual([6, 5, 4, 3, 2]);
  });

  it('deduplicates recent list when same ID selected again', () => {
    const { result } = renderHook(() => useDesign(), { wrapper: makeWrapper() });
    act(() => {
      result.current.setActiveDesignId(1);
      result.current.setActiveDesignId(2);
      result.current.setActiveDesignId(1);
    });
    expect(result.current.recentDesignIds).toEqual([1, 2]);
  });
});
