import {
  createContext,
  useContext,
  useState,
  useMemo,
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
    return null;
  } catch {
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
      return parsed as number[];
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

  // Derive the validated active ID: once designs have loaded, clear any stored ID
  // that doesn't correspond to a known design. No setState needed in useEffect.
  const activeDesignId = useMemo<number | null>(() => {
    if (!isSuccess) return storedId; // not yet validated — show stored value optimistically
    if (storedId === null) return null;
    const valid = designs.some((d) => d.id === storedId);
    if (!valid) {
      // Silently clean up stale localStorage entry (side effect in memo is intentional
      // here — it's a write to an external system, not setState)
      localStorage.removeItem(LS_ACTIVE_KEY);
      return null;
    }
    return storedId;
  }, [storedId, isSuccess, designs]);

  const setActiveDesignId = useCallback((id: number | null) => {
    setStoredId(id);
    if (id === null) {
      localStorage.removeItem(LS_ACTIVE_KEY);
    } else {
      localStorage.setItem(LS_ACTIVE_KEY, JSON.stringify(id));
      setRecentDesignIds((prev) => {
        const next = pushRecent(id, prev);
        localStorage.setItem(LS_RECENT_KEY, JSON.stringify(next));
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
