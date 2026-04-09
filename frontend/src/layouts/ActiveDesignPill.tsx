import { Link, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ChevronDown, Network } from 'lucide-react';
import { useDesign } from '@/contexts/DesignContext';
import { designsApi } from '@/api/designs';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Skeleton } from '@/components/ui/skeleton';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import type { Design } from '@/models';

const MAX_NAME_LEN = 24;

function truncate(name: string): string {
  return name.length > MAX_NAME_LEN ? name.slice(0, MAX_NAME_LEN - 1) + '…' : name;
}

const pillClass = cn(
  'flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors',
  'hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring cursor-pointer'
);

export function ActiveDesignPill() {
  const { activeDesignId, recentDesignIds, setActiveDesignId } = useDesign();
  const navigate = useNavigate();

  const { data: designs, isLoading } = useQuery({
    queryKey: ['designs'],
    queryFn: designsApi.list,
    staleTime: 60_000,
  });

  if (isLoading) {
    return <Skeleton className="h-7 w-36 rounded-full" />;
  }

  const activeDesign = designs?.find((d) => d.id === activeDesignId) ?? null;

  // Build the recent list: known designs, MRU, excluding the currently active one, cap 5
  const recentDesigns: Design[] = recentDesignIds
    .map((id) => designs?.find((d) => d.id === id))
    .filter((d): d is Design => d !== undefined)
    .filter((d) => d.id !== activeDesignId)
    .slice(0, 5);

  const handleSelect = (id: number) => {
    setActiveDesignId(id);
    // stay on current route
  };

  const triggerColorClass = activeDesign
    ? 'border-primary/30 bg-primary/10 text-primary'
    : 'border-border bg-background text-muted-foreground';

  return (
    <Popover>
      {/* Use render prop to avoid button-in-button: PopoverTrigger renders a <button> by
          default; we replace it entirely with a <div role="button"> so the inner
          Tooltip trigger (also a <button>) is valid HTML. */}
      <PopoverTrigger
        nativeButton={false}
        render={
          <div
            role="button"
            tabIndex={0}
            aria-label={activeDesign ? `Active design: ${activeDesign.name}` : 'Select design'}
            className={cn(pillClass, triggerColorClass)}
          />
        }
      >
        <Network className="size-3 shrink-0" />
        {activeDesign ? (
          <>
            {activeDesign.name.length > MAX_NAME_LEN ? (
              <Tooltip>
                <TooltipTrigger
                  render={<span className="cursor-pointer" />}
                >
                  {truncate(activeDesign.name)}
                </TooltipTrigger>
                <TooltipContent>{activeDesign.name}</TooltipContent>
              </Tooltip>
            ) : (
              <span>{activeDesign.name}</span>
            )}
          </>
        ) : (
          <Link
            to="/"
            onClick={(e) => e.stopPropagation()}
            className="hover:underline"
            aria-label="Select a design"
          >
            Select design…
          </Link>
        )}
        <ChevronDown className="size-3 shrink-0 opacity-60" />
      </PopoverTrigger>

      <PopoverContent align="end" className="w-56 p-1">
        {recentDesigns.length > 0 && (
          <>
            <p className="px-2 py-1 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground">
              Recent
            </p>
            {recentDesigns.map((d) => (
              <button
                key={d.id}
                type="button"
                onClick={() => handleSelect(d.id)}
                className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm hover:bg-muted"
              >
                <Network className="size-3.5 shrink-0 text-muted-foreground" />
                <span className="line-clamp-1">{d.name}</span>
              </button>
            ))}
            <div className="my-1 h-px bg-border" />
          </>
        )}

        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start text-xs"
          onClick={() => navigate('/')}
        >
          View all designs
        </Button>
      </PopoverContent>
    </Popover>
  );
}
