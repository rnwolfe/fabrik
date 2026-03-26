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
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

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

  readonly loading = signal(true);
  readonly notFound = signal(false);
  readonly error = signal<string | null>(null);
  readonly article = signal<KnowledgeArticle | null>(null);
  readonly renderedHtml = signal<SafeHtml>('');

  ngOnInit(): void {
    // Load KaTeX in parallel — don't block article fetching on it
    this.markdownRenderer.loadKatex().catch(() => {
      // KaTeX unavailable — math will render as raw LaTeX
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

    this.knowledgeService.getArticle(this.articlePath).subscribe({
      next: async (a) => {
        this.article.set(a);
        if (a.content) {
          try {
            const html = await this.markdownRenderer.render(a.content);
            this.renderedHtml.set(this.sanitizer.bypassSecurityTrustHtml(html));
            this.loading.set(false);
            // Post-process Mermaid after DOM update
            setTimeout(() => this.postProcessContent(), 0);
          } catch (err) {
            this.error.set('Failed to render article content.');
            this.loading.set(false);
            if (typeof ngDevMode !== 'undefined' && ngDevMode) {
              console.warn('ArticleViewComponent: render error', err);
            }
          }
        } else {
          this.loading.set(false);
        }
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
  }

  private postProcessContent(): void {
    const container = this.contentContainer?.nativeElement;
    if (!container) {
      return;
    }

    // Render Mermaid diagrams
    this.markdownRenderer.renderMermaid(container);

    // Wire up internal knowledge links
    const links = container.querySelectorAll<HTMLAnchorElement>('a[data-knowledge-link="true"]');
    for (const link of Array.from(links)) {
      link.addEventListener('click', (e) => {
        e.preventDefault();
        const href = link.getAttribute('href');
        if (href) {
          const path = href.replace(/\.md$/, '');
          this.internalLinkClicked.emit(path);
        } else {
          if (typeof ngDevMode !== 'undefined' && ngDevMode) {
            console.warn('ArticleViewComponent: internal link missing href', link);
          }
        }
      });
    }
  }

  formatCategory(cat: string): string {
    return cat.charAt(0).toUpperCase() + cat.slice(1).replace(/-/g, ' ');
  }
}
