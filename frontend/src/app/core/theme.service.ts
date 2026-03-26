import { Injectable } from '@angular/core';

export type Theme = 'light' | 'dark';

const THEME_KEY = 'fabrik-theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private _current: Theme = 'light';

  constructor() {
    this._current = this._resolveInitialTheme();
    this._applyTheme(this._current);
  }

  get current(): Theme {
    return this._current;
  }

  toggle(): void {
    this.setTheme(this._current === 'light' ? 'dark' : 'light');
  }

  setTheme(theme: Theme): void {
    this._current = theme;
    this._applyTheme(theme);
    this._persist(theme);
  }

  private _resolveInitialTheme(): Theme {
    const saved = this._loadSaved();
    if (saved) return saved;
    const prefersDark =
      typeof window !== 'undefined' &&
      window.matchMedia?.('(prefers-color-scheme: dark)').matches;
    return prefersDark ? 'dark' : 'light';
  }

  private _applyTheme(theme: Theme): void {
    if (typeof document === 'undefined') return;
    document.body.setAttribute('data-theme', theme);
    document.body.style.colorScheme = theme;
  }

  private _persist(theme: Theme): void {
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch {
      // localStorage unavailable — ignore silently
    }
  }

  private _loadSaved(): Theme | null {
    try {
      const v = localStorage.getItem(THEME_KEY);
      if (v === 'light' || v === 'dark') return v;
    } catch {
      // localStorage unavailable — ignore silently
    }
    return null;
  }
}
