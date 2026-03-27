import {
  Component,
  ElementRef,
  Input,
  OnDestroy,
  ViewChild,
  AfterViewInit,
  HostListener,
  signal,
  inject,
  PLATFORM_ID,
} from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';

const SESSION_KEY_PREFIX = 'fabrik-split-pane-';

/**
 * A two-pane horizontal split layout with a draggable divider.
 * Pane sizes are persisted in sessionStorage.
 *
 * Usage:
 *   <app-split-pane storageKey="my-split" [initialLeftPct]="65">
 *     <ng-container left>...</ng-container>
 *     <ng-container right>...</ng-container>
 *   </app-split-pane>
 */
@Component({
  selector: 'app-split-pane',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './split-pane.component.html',
  styleUrl: './split-pane.component.scss',
})
export class SplitPaneComponent implements AfterViewInit, OnDestroy {
  @ViewChild('container') containerRef!: ElementRef<HTMLDivElement>;
  @ViewChild('divider') dividerRef!: ElementRef<HTMLDivElement>;

  @Input() storageKey = 'default';
  @Input() initialLeftPct = 55;
  @Input() minLeftPct = 20;
  @Input() maxLeftPct = 80;

  private readonly _platformId = inject(PLATFORM_ID);
  private _isBrowser = isPlatformBrowser(this._platformId);

  leftPct = signal(this.initialLeftPct);

  private _dragging = false;
  private _startX = 0;
  private _startPct = 0;
  private _containerWidth = 0;

  private _onMouseMoveBound = this._onMouseMove.bind(this);
  private _onMouseUpBound = this._onMouseUp.bind(this);

  ngAfterViewInit(): void {
    const stored = this._loadFromSession();
    if (stored !== null) {
      this.leftPct.set(stored);
    } else {
      this.leftPct.set(this.initialLeftPct);
    }
  }

  ngOnDestroy(): void {
    this._removeListeners();
  }

  onDividerMousedown(evt: MouseEvent): void {
    if (!this.containerRef) return;
    evt.preventDefault();
    this._dragging = true;
    this._startX = evt.clientX;
    this._startPct = this.leftPct();
    this._containerWidth = this.containerRef.nativeElement.getBoundingClientRect().width;

    document.addEventListener('mousemove', this._onMouseMoveBound);
    document.addEventListener('mouseup', this._onMouseUpBound);
  }

  onDividerKeydown(evt: KeyboardEvent): void {
    const step = 2;
    if (evt.key === 'ArrowLeft') {
      evt.preventDefault();
      this._updatePct(this.leftPct() - step);
    } else if (evt.key === 'ArrowRight') {
      evt.preventDefault();
      this._updatePct(this.leftPct() + step);
    }
  }

  @HostListener('window:resize')
  onWindowResize(): void {
    // Clamp on resize in case container width changes
    const clamped = Math.min(Math.max(this.leftPct(), this.minLeftPct), this.maxLeftPct);
    this.leftPct.set(clamped);
  }

  private _onMouseMove(evt: MouseEvent): void {
    if (!this._dragging || !this._containerWidth) return;
    const delta = evt.clientX - this._startX;
    const deltaPct = (delta / this._containerWidth) * 100;
    this._updatePct(this._startPct + deltaPct);
  }

  private _onMouseUp(): void {
    this._dragging = false;
    this._removeListeners();
    this._saveToSession(this.leftPct());
  }

  private _updatePct(pct: number): void {
    const clamped = Math.min(Math.max(pct, this.minLeftPct), this.maxLeftPct);
    this.leftPct.set(clamped);
  }

  private _removeListeners(): void {
    document.removeEventListener('mousemove', this._onMouseMoveBound);
    document.removeEventListener('mouseup', this._onMouseUpBound);
  }

  private _sessionKey(): string {
    return `${SESSION_KEY_PREFIX}${this.storageKey}`;
  }

  private _saveToSession(pct: number): void {
    if (!this._isBrowser) return;
    try {
      sessionStorage.setItem(this._sessionKey(), String(pct));
    } catch {
      // sessionStorage may not be available in all environments
    }
  }

  private _loadFromSession(): number | null {
    if (!this._isBrowser) return null;
    try {
      const raw = sessionStorage.getItem(this._sessionKey());
      if (raw === null) return null;
      const n = parseFloat(raw);
      if (!Number.isFinite(n)) return null;
      return Math.min(Math.max(n, this.minLeftPct), this.maxLeftPct);
    } catch {
      return null;
    }
  }
}
