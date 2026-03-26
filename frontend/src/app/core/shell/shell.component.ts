import {
  Component,
  OnInit,
  OnDestroy,
  inject,
  signal,
  computed,
  HostListener,
} from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { NavigationStart, NavigationEnd, NavigationCancel, NavigationError, Router } from '@angular/router';
import { filter, Subscription } from 'rxjs';

import { ThemeService } from '../theme.service';

const SIDEBAR_KEY = 'fabrik-sidebar-open';
const MOBILE_BREAKPOINT = 768;

export interface NavItem {
  label: string;
  icon: string;
  route: string;
}

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [
    RouterLink,
    RouterLinkActive,
    RouterOutlet,
    MatToolbarModule,
    MatSidenavModule,
    MatListModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    MatProgressBarModule,
  ],
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.scss',
})
export class ShellComponent implements OnInit, OnDestroy {
  private readonly _theme = inject(ThemeService);
  private readonly _router = inject(Router);
  private _sub?: Subscription;

  readonly navItems: NavItem[] = [
    { label: 'Home', icon: 'home', route: '/' },
    { label: 'Design', icon: 'device_hub', route: '/design' },
    { label: 'Catalog', icon: 'inventory_2', route: '/catalog' },
    { label: 'Metrics', icon: 'bar_chart', route: '/metrics' },
    { label: 'Knowledge Base', icon: 'menu_book', route: '/knowledge' },
  ];

  isMobile = signal(window.innerWidth < MOBILE_BREAKPOINT);
  sidenavOpen = signal(this._loadSidebarPref());
  isLoading = signal(false);

  readonly isDark = computed(() => this._theme.current === 'dark');

  get sidenavMode(): 'over' | 'side' {
    return this.isMobile() ? 'over' : 'side';
  }

  @HostListener('window:resize')
  onResize(): void {
    const mobile = window.innerWidth < MOBILE_BREAKPOINT;
    this.isMobile.set(mobile);
    if (mobile) {
      this.sidenavOpen.set(false);
    }
  }

  ngOnInit(): void {
    this._sub = this._router.events
      .pipe(
        filter(
          e =>
            e instanceof NavigationStart ||
            e instanceof NavigationEnd ||
            e instanceof NavigationCancel ||
            e instanceof NavigationError,
        ),
      )
      .subscribe(e => {
        this.isLoading.set(e instanceof NavigationStart);
        if (this.isMobile() && e instanceof NavigationEnd) {
          this.sidenavOpen.set(false);
        }
      });
  }

  ngOnDestroy(): void {
    this._sub?.unsubscribe();
  }

  toggleSidenav(): void {
    const next = !this.sidenavOpen();
    this.sidenavOpen.set(next);
    this._persistSidebarPref(next);
  }

  toggleTheme(): void {
    this._theme.toggle();
  }

  private _loadSidebarPref(): boolean {
    try {
      const v = localStorage.getItem(SIDEBAR_KEY);
      if (v !== null) return v === 'true';
    } catch {
      // localStorage unavailable
    }
    return true; // default open
  }

  private _persistSidebarPref(open: boolean): void {
    try {
      localStorage.setItem(SIDEBAR_KEY, String(open));
    } catch {
      // localStorage unavailable
    }
  }
}
