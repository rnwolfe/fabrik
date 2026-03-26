import { TestBed } from '@angular/core/testing';
import { render, screen } from '@testing-library/angular';
import userEvent from '@testing-library/user-event';
import { provideRouter } from '@angular/router';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { ShellComponent } from './shell.component';
import { ThemeService } from '../theme.service';

describe('ShellComponent', () => {
  const renderShell = () =>
    render(ShellComponent, {
      providers: [provideRouter([]), provideAnimationsAsync()],
    });

  beforeEach(() => {
    localStorage.clear();
    document.body.removeAttribute('data-theme');
  });

  it('should render the app title', async () => {
    await renderShell();
    expect(screen.getByText('fabrik')).toBeTruthy();
  });

  it('should render all nav items', async () => {
    await renderShell();
    expect(screen.getAllByText('Home').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Design').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Catalog').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Metrics').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('Knowledge Base').length).toBeGreaterThanOrEqual(1);
  });

  it('should have a menu toggle button', async () => {
    await renderShell();
    const btn = screen.getByRole('button', { name: /toggle navigation menu/i });
    expect(btn).toBeTruthy();
  });

  it('should have a theme toggle button', async () => {
    await renderShell();
    const btn = screen.getByRole('button', { name: /toggle dark mode/i });
    expect(btn).toBeTruthy();
  });

  it('should toggle theme when theme button is clicked', async () => {
    const user = userEvent.setup();
    await renderShell();
    const themeService = TestBed.inject(ThemeService);
    expect(themeService.current()).toBe('light');
    const btn = screen.getByRole('button', { name: /toggle dark mode/i });
    await user.click(btn);
    expect(themeService.current()).toBe('dark');
  });
});
