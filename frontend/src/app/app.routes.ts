import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'knowledge',
    loadComponent: () =>
      import('./features/knowledge/components/knowledge-viewer/knowledge-viewer.component').then(
        m => m.KnowledgeViewerComponent
      ),
  },
  {
    path: '',
    redirectTo: 'knowledge',
    pathMatch: 'full',
  },
];
