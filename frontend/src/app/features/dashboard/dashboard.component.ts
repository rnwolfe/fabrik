import { Component, OnInit, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { DatePipe } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { DashboardService } from './dashboard.service';
import { Design } from '../../models/design.model';

interface QuickAction {
  label: string;
  description: string;
  icon: string;
  route: string;
}

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    RouterLink,
    DatePipe,
    MatCardModule,
    MatButtonModule,
    MatIconModule,
    MatListModule,
    MatDividerModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.scss',
})
export class DashboardComponent implements OnInit {
  private readonly _service = inject(DashboardService);

  designs = signal<Design[]>([]);
  loading = signal(true);

  readonly quickActions: QuickAction[] = [
    {
      label: 'New Design',
      description: 'Start a new datacenter topology',
      icon: 'add_circle',
      route: '/design',
    },
    {
      label: 'Browse Catalog',
      description: 'Explore hardware platforms',
      icon: 'inventory_2',
      route: '/catalog',
    },
    {
      label: 'View Metrics',
      description: 'Analyze capacity and power',
      icon: 'bar_chart',
      route: '/metrics',
    },
    {
      label: 'Knowledge Base',
      description: 'Learn datacenter design concepts',
      icon: 'menu_book',
      route: '/knowledge',
    },
  ];

  ngOnInit(): void {
    this._service.getRecentDesigns().subscribe(designs => {
      this.designs.set(designs);
      this.loading.set(false);
    });
  }
}
