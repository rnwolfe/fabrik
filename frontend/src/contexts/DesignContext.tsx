import { createContext, useContext, useState, type ReactNode } from 'react';

interface DesignContextType {
  activeDesignId: number | null;
  setActiveDesignId: (id: number | null) => void;
}

const DesignContext = createContext<DesignContextType>({
  activeDesignId: null,
  setActiveDesignId: () => {},
});

export function DesignProvider({ children }: { children: ReactNode }) {
  const [activeDesignId, setActiveDesignId] = useState<number | null>(null);
  return (
    <DesignContext.Provider value={{ activeDesignId, setActiveDesignId }}>
      {children}
    </DesignContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export const useDesign = () => useContext(DesignContext);
