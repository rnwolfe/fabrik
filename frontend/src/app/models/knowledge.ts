/** KnowledgeArticle mirrors the Go knowledge.Article type. */
export interface KnowledgeArticle {
  /** URL-safe identifier, e.g. "networking/clos-fabric-fundamentals". */
  path: string;
  title: string;
  category: string;
  tags: string[];
  /** Raw Markdown body — only present on full article loads (not index). */
  content?: string;
}

/** KnowledgeIndex mirrors the Go knowledge.Index type. */
export interface KnowledgeIndex {
  articles: KnowledgeArticle[];
}
