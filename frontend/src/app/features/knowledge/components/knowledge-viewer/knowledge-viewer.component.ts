import {
  Component,
  OnInit,
  inject,
  signal,
  computed,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule, ActivatedRoute } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MatListModule } from '@angular/material/list';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatDividerModule } from '@angular/material/divider';

import { KnowledgeService } from '../../knowledge.service';
import { KnowledgeArticle } from '../../../../models';
import { ArticleViewComponent } from '../article-view/article-view.component';

/**
 * KnowledgeViewerComponent is the full-page knowledge base viewer at /knowledge.
 * It provides:
 *  - A hierarchical TOC with category grouping in a sidenav
 *  - Full-text search across all articles
 *  - Article rendering via ArticleViewComponent
 */
@Component({
  selector: 'app-knowledge-viewer',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    FormsModule,
    MatListModule,
    MatInputModule,
    MatFormFieldModule,
    MatIconModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatSidenavModule,
    MatDividerModule,
    ArticleViewComponent,
  ],
  templateUrl: './knowledge-viewer.component.html',
  styleUrl: './knowledge-viewer.component.scss',
})
export class KnowledgeViewerComponent implements OnInit {
  private readonly knowledgeService = inject(KnowledgeService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly allArticles = signal<KnowledgeArticle[]>([]);
  readonly searchQuery = signal('');
  readonly selectedPath = signal<string | null>(null);

  readonly filteredArticles = computed(() => {
    return this.knowledgeService.search(this.allArticles(), this.searchQuery());
  });

  readonly groupedArticles = computed(() => {
    return this.knowledgeService.groupByCategory(this.filteredArticles());
  });

  readonly sortedCategories = computed(() => {
    return Array.from(this.groupedArticles().keys()).sort();
  });

  ngOnInit(): void {
    // Load index
    this.knowledgeService.getIndex().subscribe({
      next: (idx) => {
        this.allArticles.set(idx.articles);
        this.loading.set(false);

        // Check for article param in route
        const articlePath = this.route.snapshot.queryParamMap.get('article');
        if (articlePath) {
          this.selectedPath.set(articlePath);
        } else if (idx.articles.length > 0) {
          this.selectedPath.set(idx.articles[0].path);
        }
      },
      error: (err) => {
        this.error.set('Failed to load knowledge base.');
        this.loading.set(false);
        if (typeof ngDevMode !== 'undefined' && ngDevMode) {
          console.warn('KnowledgeViewerComponent: failed to load index', err);
        }
      },
    });
  }

  selectArticle(path: string): void {
    this.selectedPath.set(path);
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { article: path },
      queryParamsHandling: 'merge',
    });
  }

  onSearchChange(value: string): void {
    this.searchQuery.set(value);
  }

  clearSearch(): void {
    this.searchQuery.set('');
  }

  formatCategory(cat: string): string {
    return cat.charAt(0).toUpperCase() + cat.slice(1).replace(/-/g, ' ');
  }

  articlesInCategory(cat: string): KnowledgeArticle[] {
    return this.groupedArticles().get(cat) ?? [];
  }

  isSelected(path: string): boolean {
    return this.selectedPath() === path;
  }
}
