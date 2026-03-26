import { Routes } from '@angular/router';
import { ShellComponent } from './core/shell/shell.component';

export const routes: Routes = [
  {
    path: '',
    component: ShellComponent,
    children: [
      {
        path: '',
        loadChildren: () =>
          import('./features/dashboard/dashboard.routes').then(
            m => m.DASHBOARD_ROUTES,
          ),
      },
      {
        path: 'design',
        loadChildren: () =>
          import('./features/topology/topology.routes').then(
            m => m.TOPOLOGY_ROUTES,
          ),
      },
      {
        path: 'catalog',
        loadChildren: () =>
          import('./features/catalog/catalog.routes').then(
            m => m.CATALOG_ROUTES,
          ),
      },
      {
        path: 'metrics',
        loadChildren: () =>
          import('./features/metrics/metrics.routes').then(
            m => m.METRICS_ROUTES,
          ),
      },
      {
        path: 'knowledge',
        loadComponent: () =>
          import('./features/knowledge/components/knowledge-viewer/knowledge-viewer.component').then(
            m => m.KnowledgeViewerComponent,
          ),
      },
    ],
  },
];
