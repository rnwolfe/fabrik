import { useState, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { knowledgeApi } from '@/api/knowledge';
import { EmptyState } from '@/components/EmptyState';
import { Input } from '@/components/ui/input';
import { BookOpen, Search, ChevronRight, FileText } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { KnowledgeArticle } from '@/models';

export default function KnowledgePage() {
  const { '*': articlePath } = useParams();
  const navigate = useNavigate();
  const [search, setSearch] = useState('');

  const { data: index, isLoading: indexLoading } = useQuery({
    queryKey: ['knowledge-index'],
    queryFn: knowledgeApi.index,
  });

  const { data: article, isLoading: articleLoading } = useQuery({
    queryKey: ['knowledge-article', articlePath],
    queryFn: () => knowledgeApi.article(articlePath!),
    enabled: !!articlePath,
  });

  const articles = index?.articles ?? [];

  const filtered = useMemo(() => {
    if (!search) return articles;
    const q = search.toLowerCase();
    return articles.filter(
      (a) =>
        a.title.toLowerCase().includes(q) ||
        a.category.toLowerCase().includes(q) ||
        a.tags.some((t) => t.toLowerCase().includes(q))
    );
  }, [articles, search]);

  const grouped = useMemo(() => {
    const map = new Map<string, KnowledgeArticle[]>();
    for (const a of filtered) {
      const list = map.get(a.category) ?? [];
      list.push(a);
      map.set(a.category, list);
    }
    return map;
  }, [filtered]);

  const handleSelectArticle = (path: string) => {
    navigate(`/knowledge/${path}`);
  };

  return (
    <div className="flex h-full gap-0 -m-6 overflow-hidden" style={{ height: 'calc(100vh - 48px)' }}>
      {/* Left sidebar */}
      <div className="flex w-[260px] shrink-0 flex-col border-r border-border bg-muted/20">
        <div className="border-b border-border p-3">
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search articles…"
              className="pl-8"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-2">
          {indexLoading ? (
            <div className="space-y-2 p-2">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-8 animate-pulse rounded-lg bg-muted" />
              ))}
            </div>
          ) : grouped.size === 0 ? (
            <p className="p-4 text-center text-xs text-muted-foreground">
              {search ? 'No matching articles' : 'No articles available'}
            </p>
          ) : (
            Array.from(grouped.entries()).map(([category, items]) => (
              <div key={category} className="mb-3">
                <p className="mb-1 px-2 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
                  {category}
                </p>
                <div className="space-y-0.5">
                  {items.map((item) => (
                    <button
                      key={item.path}
                      onClick={() => handleSelectArticle(item.path)}
                      className={cn(
                        'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors',
                        articlePath === item.path
                          ? 'bg-primary text-primary-foreground'
                          : 'text-foreground/70 hover:bg-muted hover:text-foreground'
                      )}
                    >
                      <FileText className="size-3.5 shrink-0 opacity-60" />
                      <span className="line-clamp-1 flex-1">{item.title}</span>
                      {articlePath === item.path && (
                        <ChevronRight className="size-3.5 shrink-0" />
                      )}
                    </button>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-1 flex-col overflow-y-auto">
        {!articlePath ? (
          <div className="flex flex-1 items-center justify-center p-8">
            <EmptyState
              icon={BookOpen}
              title="Knowledge Base"
              description="Select an article from the left sidebar to start reading about datacenter design concepts."
            />
          </div>
        ) : articleLoading ? (
          <div className="p-8 space-y-3">
            <div className="h-8 w-1/2 animate-pulse rounded bg-muted" />
            <div className="h-4 w-full animate-pulse rounded bg-muted" />
            <div className="h-4 w-3/4 animate-pulse rounded bg-muted" />
            <div className="h-4 w-full animate-pulse rounded bg-muted" />
          </div>
        ) : !article ? (
          <div className="flex flex-1 items-center justify-center p-8">
            <EmptyState
              icon={FileText}
              title="Article not found"
              description="The requested article could not be loaded."
            />
          </div>
        ) : (
          <div className="mx-auto w-full max-w-3xl p-8">
            <div className="mb-4 flex flex-wrap gap-2">
              <span className="rounded-full bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground">
                {article.category}
              </span>
              {article.tags.map((tag) => (
                <span
                  key={tag}
                  className="rounded-full bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary"
                >
                  {tag}
                </span>
              ))}
            </div>
            <article className="prose prose-slate dark:prose-invert max-w-none prose-headings:font-semibold prose-headings:tracking-tight prose-a:text-primary prose-a:no-underline hover:prose-a:underline prose-code:before:content-none prose-code:after:content-none prose-code:rounded prose-code:bg-muted prose-code:px-1 prose-code:py-0.5 prose-code:text-sm prose-code:font-mono prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-pre:text-foreground prose-table:text-sm prose-th:text-muted-foreground prose-img:rounded-lg">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {article.content ?? ''}
              </ReactMarkdown>
            </article>
          </div>
        )}
      </div>
    </div>
  );
}
