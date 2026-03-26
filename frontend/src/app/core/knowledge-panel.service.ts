import { Injectable, signal } from '@angular/core';

/** State managed by KnowledgePanelService. */
export interface PanelState {
  /** Whether the slide-out panel is open. */
  isOpen: boolean;
  /** The article path currently displayed, or null if none. */
  articlePath: string | null;
}

/**
 * KnowledgePanelService is a singleton that manages the state of the
 * contextual help slide-out panel. Components open/close the panel by
 * calling the public methods; the panel component reads the state via signals.
 */
@Injectable({ providedIn: 'root' })
export class KnowledgePanelService {
  private readonly _state = signal<PanelState>({ isOpen: false, articlePath: null });

  /** Read-only signal of the current panel state. */
  readonly state = this._state.asReadonly();

  /**
   * Opens the slide-out panel and loads the given article.
   * If the panel is already open with the same article, it closes instead
   * (toggle behaviour).
   */
  open(articlePath: string): void {
    const current = this._state();
    if (current.isOpen && current.articlePath === articlePath) {
      this.close();
    } else {
      this._state.set({ isOpen: true, articlePath });
    }
  }

  /** Closes the slide-out panel. */
  close(): void {
    this._state.set({ isOpen: false, articlePath: null });
  }
}
