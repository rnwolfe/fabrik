import { Injectable, signal } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class KnowledgePanelService {
  private _open = signal(false);
  private _articleSlug = signal<string | null>(null);

  readonly isOpen = this._open.asReadonly();
  readonly articleSlug = this._articleSlug.asReadonly();

  open(slug?: string): void {
    this._articleSlug.set(slug ?? null);
    this._open.set(true);
  }

  close(): void {
    this._open.set(false);
  }

  toggle(): void {
    this._open.update(v => !v);
  }
}
