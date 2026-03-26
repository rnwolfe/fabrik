import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { KnowledgeArticle, KnowledgeIndex } from '../../models';

/**
 * KnowledgeService fetches knowledge base data from the Go REST API.
 * It also provides client-side full-text search across the loaded index.
 */
@Injectable({ providedIn: 'root' })
export class KnowledgeService {
  private readonly http = inject(HttpClient);
  private readonly base = '/api/knowledge';

  /** Fetches the knowledge index (all article metadata, no content). */
  getIndex(): Observable<KnowledgeIndex> {
    return this.http.get<KnowledgeIndex>(this.base);
  }

  /** Fetches a single article including Markdown content. */
  getArticle(path: string): Observable<KnowledgeArticle> {
    return this.http.get<KnowledgeArticle>(`${this.base}/${path}`);
  }

  /**
   * Performs client-side full-text search across article metadata.
   * Searches title, category, tags, and path.
   * Returns articles sorted by relevance (title match > category > tags > path).
   */
  search(articles: KnowledgeArticle[], query: string): KnowledgeArticle[] {
    const q = query.trim().toLowerCase();
    if (!q) {
      return articles;
    }

    const scored = articles.map(a => {
      let score = 0;
      const title = a.title.toLowerCase();
      const category = a.category.toLowerCase();
      const path = a.path.toLowerCase();
      const tags = a.tags.map(t => t.toLowerCase());

      if (title.includes(q)) score += 10;
      if (title.startsWith(q)) score += 5;
      if (category.includes(q)) score += 4;
      if (tags.some(t => t.includes(q))) score += 3;
      if (path.includes(q)) score += 2;

      return { article: a, score };
    });

    return scored
      .filter(s => s.score > 0)
      .sort((a, b) => b.score - a.score)
      .map(s => s.article);
  }

  /**
   * Groups articles by category.
   */
  groupByCategory(articles: KnowledgeArticle[]): Map<string, KnowledgeArticle[]> {
    const map = new Map<string, KnowledgeArticle[]>();
    for (const article of articles) {
      const cat = article.category;
      if (!map.has(cat)) {
        map.set(cat, []);
      }
      map.get(cat)!.push(article);
    }
    return map;
  }
}
