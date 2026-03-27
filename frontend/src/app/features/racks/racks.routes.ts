import { Routes } from '@angular/router';
import { RacksComponent } from './racks.component';

export const RACKS_ROUTES: Routes = [
  {
    path: '',
    component: RacksComponent,
    children: [
      {
        path: '',
        loadComponent: () =>
          import('./components/rack-list/rack-list.component').then(
            m => m.RackListComponent,
          ),
      },
    ],
  },
];
