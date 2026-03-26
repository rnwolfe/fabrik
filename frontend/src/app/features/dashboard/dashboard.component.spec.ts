import { render, screen } from '@testing-library/angular';
import { provideRouter } from '@angular/router';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { DashboardComponent } from './dashboard.component';
import { Design } from '../../models/design.model';

describe('DashboardComponent', () => {
  const renderDashboard = () =>
    render(DashboardComponent, {
      providers: [
        provideRouter([]),
        provideAnimationsAsync(),
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });

  it('should render welcome heading', async () => {
    await renderDashboard();
    // Flush empty response
    const ctrl = TestBed.inject(HttpTestingController);
    ctrl.expectOne('/api/designs').flush([]);
    expect(await screen.findByText('Welcome to fabrik')).toBeTruthy();
  });

  it('should show all quick actions', async () => {
    await renderDashboard();
    const ctrl = TestBed.inject(HttpTestingController);
    ctrl.expectOne('/api/designs').flush([]);
    expect(await screen.findByText('New Design')).toBeTruthy();
    expect(screen.getByText('Browse Catalog')).toBeTruthy();
    expect(screen.getByText('View Metrics')).toBeTruthy();
    expect(screen.getByText('Knowledge Base')).toBeTruthy();
  });

  it('should show empty state when no designs', async () => {
    await renderDashboard();
    const ctrl = TestBed.inject(HttpTestingController);
    ctrl.expectOne('/api/designs').flush([]);
    expect(await screen.findByText('No designs yet.')).toBeTruthy();
  });

  it('should show designs when returned by API', async () => {
    await renderDashboard();
    const ctrl = TestBed.inject(HttpTestingController);
    const mockDesigns: Design[] = [
      {
        id: 1,
        name: 'Test Fabric',
        description: 'A test design',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      },
    ];
    ctrl.expectOne('/api/designs').flush(mockDesigns);
    expect(await screen.findByText('Test Fabric')).toBeTruthy();
  });

  it('should show empty state when API fails', async () => {
    await renderDashboard();
    const ctrl = TestBed.inject(HttpTestingController);
    ctrl.expectOne('/api/designs').error(new ProgressEvent('error'));
    expect(await screen.findByText('No designs yet.')).toBeTruthy();
  });
});
