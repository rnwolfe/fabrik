import {
  Component,
  OnInit,
  OnChanges,
  SimpleChanges,
  Input,
  Output,
  EventEmitter,
  ElementRef,
  ViewChild,
  inject,
  signal,
  DestroyRef,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { Subject, from, switchMap, map, takeUntil } from 'rxjs';
import DOMPurify from 'dompurify';

import { KnowledgeService } from '../../knowledge.service';
import { MarkdownRendererService } from '../../markdown-renderer.service';
import { KnowledgeArticle } from '../../../../models';

/**
 * ArticleViewComponent renders a single knowledge base article.
 * It handles:
 *  - Fetching the article from the API
 *  - Rendering Markdown with syntax highlighting, Mermaid, and KaTeX
 *  - Internal cross-link resolution
 *  - "Not found" state
 */
@Component({
  selector: 'app-article-view',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatChipsModule,
  ],
  templateUrl: './article-view.component.html',
  styleUrl: './article-view.component.scss',
})
export class ArticleViewComponent implements OnInit, OnChanges {
  @Input({ required: true }) articlePath!: string;
  /** Compact mode: used when embedded in the slide-out panel */
  @Input() compact = false;
  /** Emitted when the user clicks an internal knowledge link. */
  @Output() internalLinkClicked = new EventEmitter<string>();

  @ViewChild('contentContainer') contentContainer?: ElementRef<HTMLElement>;

  private readonly knowledgeService = inject(KnowledgeService);
  private readonly markdownRenderer = inject(MarkdownRendererService);
  private readonly sanitizer = inject(DomSanitizer);
  private readonly destroyRef = inject(DestroyRef);

  readonly loading = signal(true);
  readonly notFound = signal(false);
  readonly error = signal<string | null>(null);
  readonly article = signal<KnowledgeArticle | null>(null);
  readonly renderedHtml = signal<SafeHtml>('');

  /** Subject used to cancel in-flight requests when articlePath changes. */
  private readonly articlePath$ = new Subject<string>();
  /** Emits when the component is destroyed, completing all subscriptions. */
  private readonly destroy$ = new Subject<void>();

  ngOnInit(): void {
    // Load KaTeX in parallel — don't block article fetching on it
    this.markdownRenderer.loadKatex().catch(() => {
      // KaTeX unavailable — math will render as raw LaTeX
    });

    // Register cleanup on destroy via DestroyRef
    this.destroyRef.onDestroy(() => {
      this.destroy$.next();
      this.destroy$.complete();
    });

    // switchMap cancels in-flight requests AND in-flight renders when a new
    // path arrives. Piping the async render through from() inside the inner
    // switchMap ensures a stale render cannot overwrite a newer article.
    this.articlePath$
      .pipe(
        switchMap((path) =>
          this.knowledgeService.getArticle(path).pipe(
            switchMap((article) => {
              if (!article.content) {
                return from(Promise.resolve({ article, rawHtml: null }));
              }
              return from(this.markdownRenderer.render(article.content)).pipe(
                map((rawHtml) => ({ article, rawHtml })),
              );
            }),
          ),
        ),
        takeUntil(this.destroy$),
      )
      .subscribe({
        next: ({ article, rawHtml }) => {
          this.article.set(article);
          if (rawHtml !== null) {
            // Sanitize with DOMPurify before bypassing Angular's sanitizer.
            // ADD_ATTR preserves target (external links) and data-knowledge-link
            // (internal link wiring) which DOMPurify strips by default.
            const safeHtml = DOMPurify.sanitize(rawHtml, {
              USE_PROFILES: { html: true },
              ADD_ATTR: ['target', 'data-knowledge-link'],
            });
            this.renderedHtml.set(this.sanitizer.bypassSecurityTrustHtml(safeHtml));
          }
          this.loading.set(false);
          // Post-process Mermaid after DOM update
          setTimeout(() => this.postProcessContent(), 0);
        },
        error: (err) => {
          if (err.status === 404) {
            this.notFound.set(true);
          } else {
            this.error.set('Failed to load article.');
            if (typeof ngDevMode !== 'undefined' && ngDevMode) {
              console.warn('ArticleViewComponent: load error', err);
            }
          }
          this.loading.set(false);
        },
      });

    this.loadArticle();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['articlePath'] && !changes['articlePath'].firstChange) {
      this.loadArticle();
    }
  }

  private loadArticle(): void {
    this.loading.set(true);
    this.notFound.set(false);
    this.error.set(null);
    this.articlePath$.next(this.articlePath);
  }

  private postProcessContent(): void {
    const container = this.contentContainer?.nativeElement;
    if (!container) return;

    this.markdownRenderer.renderMermaid(container);

    // Use event delegation — single listener on container, not per-link
    container.addEventListener('click', (e: MouseEvent) => {
      const link = (e.target as HTMLElement).closest('a[data-knowledge-link="true"]') as HTMLAnchorElement | null;
      if (!link) return;
      e.preventDefault();
      const href = link.getAttribute('href');
      if (href) {
        this.internalLinkClicked.emit(href.replace(/\.md$/, ''));
      } else if (typeof ngDevMode !== 'undefined' && ngDevMode) {
        console.warn('ArticleViewComponent: internal link missing href', link);
      }
    });
  }

  formatCategory(cat: string): string {
    return cat.charAt(0).toUpperCase() + cat.slice(1).replace(/-/g, ' ');
  }
}
