import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import AppLayout from './layouts/AppLayout';
import { DesignProvider } from './contexts/DesignContext';
import DashboardPage from './features/dashboard/DashboardPage';
import DesignPage from './features/design/DesignPage';
import CatalogPage from './features/catalog/CatalogPage';
import RacksPage from './features/racks/RacksPage';
import MetricsPage from './features/metrics/MetricsPage';
import KnowledgePage from './features/knowledge/KnowledgePage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <DesignProvider>
        <BrowserRouter>
          <Routes>
            <Route element={<AppLayout />}>
              <Route index element={<DashboardPage />} />
              <Route path="design" element={<DesignPage />} />
              <Route path="catalog" element={<CatalogPage />} />
              <Route path="racks" element={<RacksPage />} />
              <Route path="metrics" element={<MetricsPage />} />
              <Route path="knowledge/*" element={<KnowledgePage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </DesignProvider>
    </QueryClientProvider>
  );
}
