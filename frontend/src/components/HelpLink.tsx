import { CircleHelp } from 'lucide-react';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils';

interface HelpLinkProps {
  /** Slug of the knowledge article, e.g. "oversubscription" */
  article: string;
  /** Optional in-page anchor (heading slug), e.g. "leaf-spine-ratio" */
  anchor?: string;
  className?: string;
}

/**
 * A small inline icon button that navigates to a knowledge base article.
 * Renders an accessible "?" circle icon next to labels or headings.
 */
export function HelpLink({ article, anchor, className }: HelpLinkProps) {
  const href = anchor ? `/knowledge/${article}#${anchor}` : `/knowledge/${article}`;
  const label = `Learn more about ${article.replace(/-/g, ' ')}`;

  return (
    <Link
      to={href}
      aria-label={label}
      title={label}
      className={cn(
        'inline-flex items-center justify-center rounded-full text-muted-foreground',
        'hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        'transition-colors',
        className
      )}
    >
      <CircleHelp className="size-3.5" />
    </Link>
  );
}
