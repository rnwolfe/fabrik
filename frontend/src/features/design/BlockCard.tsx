import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Plus,
  Trash2,
  Server,
  Cpu,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import type { Block, BlockAggregationSummary, RackSummary } from '@/models';

interface BlockCardProps {
  block: Block;
  aggs: BlockAggregationSummary[];
  racks: RackSummary[];
  isSelected: boolean;
  selectedRackId: number | null;
  onSelect: () => void;
  onSelectRack: (rackId: number) => void;
  onAddRack: () => void;
  onRemoveRack: (rackId: number) => void;
  onDelete: () => void;
}

export default function BlockCard({
  block,
  aggs,
  racks,
  isSelected,
  selectedRackId,
  onSelect,
  onSelectRack,
  onAddRack,
  onRemoveRack,
  onDelete,
}: BlockCardProps) {
  const [expanded, setExpanded] = useState(true);

  const frontendAgg = aggs.find((a) => a.plane === 'front_end');
  const rackCount = racks.length;
  const maxRacks = frontendAgg ? Math.floor(frontendAgg.total_ports / 2) : 0;
  const utilPct = maxRacks > 0 ? Math.round((rackCount / maxRacks) * 100) : 0;

  return (
    <div
      className={cn(
        'rounded-lg border transition-colors',
        isSelected
          ? 'border-primary bg-primary/5 shadow-sm'
          : 'border-border bg-card hover:border-primary/40'
      )}
    >
      {/* Block header */}
      <div
        className="flex items-center gap-2 px-3 py-2 cursor-pointer"
        onClick={onSelect}
      >
        <button
          onClick={(e) => {
            e.stopPropagation();
            setExpanded(!expanded);
          }}
          className="p-0.5 rounded hover:bg-muted"
        >
          {expanded ? (
            <ChevronDown className="size-3.5 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-3.5 text-muted-foreground" />
          )}
        </button>

        <Cpu className="size-4 text-blue-500 shrink-0" />
        <span className="text-sm font-medium truncate">{block.name}</span>

        <div className="ml-auto flex items-center gap-1.5">
          {frontendAgg && (
            <Badge variant="secondary" className="text-[10px] h-5 px-1.5">
              {rackCount}/{maxRacks} racks
            </Badge>
          )}
          {!frontendAgg && (
            <Badge variant="outline" className="text-[10px] h-5 px-1.5 text-amber-500 border-amber-500/30">
              No spine
            </Badge>
          )}
        </div>
      </div>

      {/* Utilization bar */}
      {frontendAgg && (
        <div className="px-3 pb-1">
          <div className="h-1 rounded-full bg-muted overflow-hidden">
            <div
              className={cn(
                'h-full rounded-full transition-all',
                utilPct >= 90
                  ? 'bg-red-500'
                  : utilPct >= 70
                  ? 'bg-amber-500'
                  : 'bg-emerald-500'
              )}
              style={{ width: `${utilPct}%` }}
            />
          </div>
        </div>
      )}

      {/* Expanded rack list */}
      {expanded && (
        <div className="border-t border-border/50">
          {racks.length === 0 ? (
            <div className="px-3 py-3 text-center">
              <p className="text-xs text-muted-foreground">No racks yet</p>
            </div>
          ) : (
            <div className="divide-y divide-border/30">
              {racks.map((rack) => (
                <div
                  key={rack.id}
                  onClick={(e) => {
                    e.stopPropagation();
                    onSelectRack(rack.id);
                  }}
                  className={cn(
                    'flex items-center gap-2 px-3 py-1.5 text-xs cursor-pointer transition-colors',
                    selectedRackId === rack.id
                      ? 'bg-primary/10'
                      : 'hover:bg-muted/50'
                  )}
                >
                  <Server className="size-3 text-muted-foreground shrink-0" />
                  <span className="truncate">{rack.name}</span>
                  <span className="ml-auto text-muted-foreground">
                    {rack.used_u}/{rack.height_u}U
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onRemoveRack(rack.id);
                    }}
                    className="p-0.5 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 className="size-3" />
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center gap-1 px-2 py-1.5 border-t border-border/30">
            <Button
              variant="ghost"
              size="sm"
              className="h-6 text-xs px-2"
              onClick={(e) => {
                e.stopPropagation();
                onAddRack();
              }}
            >
              <Plus className="size-3 mr-1" />
              Rack
            </Button>
            <div className="ml-auto">
              <Button
                variant="ghost"
                size="sm"
                className="h-6 text-xs px-2 text-muted-foreground hover:text-destructive"
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete();
                }}
              >
                <Trash2 className="size-3" />
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
