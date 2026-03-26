import { TestBed } from '@angular/core/testing';
import { ThemeService } from './theme.service';

describe('ThemeService', () => {
  let service: ThemeService;

  beforeEach(() => {
    localStorage.clear();
    document.body.removeAttribute('data-theme');
    TestBed.configureTestingModule({});
    service = TestBed.inject(ThemeService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should default to light theme when no preference saved', () => {
    // matchMedia not available in jsdom — defaults to light
    expect(service.current).toBe('light');
  });

  it('should apply data-theme attribute to body', () => {
    expect(document.body.getAttribute('data-theme')).toBe('light');
  });

  it('should toggle from light to dark', () => {
    service.toggle();
    expect(service.current).toBe('dark');
    expect(document.body.getAttribute('data-theme')).toBe('dark');
  });

  it('should toggle from dark to light', () => {
    service.setTheme('dark');
    service.toggle();
    expect(service.current).toBe('light');
  });

  it('should persist theme to localStorage', () => {
    service.setTheme('dark');
    expect(localStorage.getItem('fabrik-theme')).toBe('dark');
  });

  it('should restore theme from localStorage on init', () => {
    localStorage.setItem('fabrik-theme', 'dark');
    // Re-create service to trigger constructor
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({});
    const restored = TestBed.inject(ThemeService);
    expect(restored.current).toBe('dark');
  });
});
