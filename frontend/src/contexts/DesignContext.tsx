import {
  createContext,
  useContext,
  useState,
  useMemo,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import { useQuery } from '@tanstack/react-query';
import { designsApi } from '@/api/designs';

const LS_ACTIVE_KEY = 'fabrik:activeDesignId';
const LS_RECENT_KEY = 'fabrik:recentDesignIds';
const RECENT_MAX = 5;

function readStoredId(): number | null {
  try {
    const raw = localStorage.getItem(LS_ACTIVE_KEY);
    if (raw === null) return null;
    const parsed = JSON.parse(raw) as unknown;
    if (typeof parsed === 'number' && Number.isFinite(parsed)) return parsed;
    // Corrupt value — clean it up
    localStorage.removeItem(LS_ACTIVE_KEY);
    return null;
  } catch {
    // Parse failed — clean up the corrupt entry
    localStorage.removeItem(LS_ACTIVE_KEY);
    return null;
  }
}

function readStoredRecent(): number[] {
  try {
    const raw = localStorage.getItem(LS_RECENT_KEY);
    if (raw === null) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (
      Array.isArray(parsed) &&
      parsed.every((x) => typeof x === 'number' && Number.isFinite(x))
    ) {
      const trimmed = (parsed as number[]).slice(0, RECENT_MAX);
      // Re-persist trimmed list if it was oversized
      if (trimmed.length < parsed.length) {
        try {
          localStorage.setItem(LS_RECENT_KEY, JSON.stringify(trimmed));
        } catch {
          // Ignore storage errors on re-persist
        }
      }
      return trimmed;
    }
    return [];
  } catch {
    return [];
  }
}

function pushRecent(id: number, existing: number[]): number[] {
  const without = existing.filter((x) => x !== id);
  return [id, ...without].slice(0, RECENT_MAX);
}

interface DesignContextType {
  activeDesignId: number | null;
  setActiveDesignId: (id: number | null) => void;
  recentDesignIds: number[];
}

const DesignContext = createContext<DesignContextType>({
  activeDesignId: null,
  setActiveDesignId: () => {},
  recentDesignIds: [],
});

export function DesignProvider({ children }: { children: ReactNode }) {
  // Raw stored value — what the user explicitly set or what was in localStorage
  const [storedId, setStoredId] = useState<number | null>(readStoredId);
  const [recentDesignIds, setRecentDesignIds] = useState<number[]>(readStoredRecent);

  // Fetch designs once to validate the stored ID
  const { data: designs, isSuccess } = useQuery({
    queryKey: ['designs'],
    queryFn: designsApi.list,
    staleTime: 60_000,
  });

  // Derive the validated active ID: show stored value optimistically until designs load.
  const activeDesignId = useMemo<number | null>(() => {
    if (!isSuccess) return storedId; // not yet validated — show stored value optimistically
    if (storedId === null) return null;
    const valid = designs.some((d) => d.id === storedId);
    return valid ? storedId : null;
  }, [storedId, isSuccess, designs]);

  // Clean up stale localStorage entry when designs have loaded and stored ID is invalid.
  // Only touches the external system (localStorage); does not call setState.
  useEffect(() => {
    if (!isSuccess || storedId === null) return;
    const valid = designs.some((d) => d.id === storedId);
    if (!valid) {
      try {
        localStorage.removeItem(LS_ACTIVE_KEY);
      } catch {
        // Ignore storage errors
      }
    }
  }, [isSuccess, designs, storedId]);

  const setActiveDesignId = useCallback((id: number | null) => {
    setStoredId(id);
    if (id === null) {
      try {
        localStorage.removeItem(LS_ACTIVE_KEY);
      } catch {
        // Ignore storage errors (quota exceeded, private mode, etc.)
      }
    } else {
      try {
        localStorage.setItem(LS_ACTIVE_KEY, JSON.stringify(id));
      } catch {
        // Ignore storage errors
      }
      setRecentDesignIds((prev) => {
        const next = pushRecent(id, prev);
        try {
          localStorage.setItem(LS_RECENT_KEY, JSON.stringify(next));
        } catch {
          // Ignore storage errors
        }
        return next;
      });
    }
  }, []);

  return (
    <DesignContext.Provider value={{ activeDesignId, setActiveDesignId, recentDesignIds }}>
      {children}
    </DesignContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export const useDesign = () => useContext(DesignContext);
