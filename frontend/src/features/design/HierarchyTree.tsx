import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ChevronDown,
  ChevronRight,
  Plus,
  Trash2,
  Check,
  X,
  Building2,
  Layers2,
  Cpu,
  MoreHorizontal,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuSubContent,
} from '@/components/ui/dropdown-menu';
import { sitesApi } from '@/api/sites';
import { superblocksApi } from '@/api/superblocks';
import { blocksApi } from '@/api/blocks';
import type {
  DesignHierarchy,
  HierarchySite,
  HierarchySuperBlock,
  HierarchyBlock,
} from '@/models';

// ─── Inline editable label ───────────────────────────────────────────────────

interface InlineLabelProps {
  value: string;
  onSave: (name: string) => void;
  className?: string;
}

function InlineLabel({ value, onSave, className }: InlineLabelProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);

  const commit = () => {
    const trimmed = draft.trim();
    if (trimmed && trimmed !== value) onSave(trimmed);
    setEditing(false);
  };

  if (editing) {
    return (
      <div className="flex items-center gap-1 flex-1" onClick={(e) => e.stopPropagation()}>
        <Input
          autoFocus
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') commit();
            if (e.key === 'Escape') { setDraft(value); setEditing(false); }
          }}
          className="h-5 text-xs px-1 py-0 border-primary"
        />
        <button onClick={commit} className="p-0.5 rounded hover:bg-emerald-100 text-emerald-600">
          <Check className="size-3" />
        </button>
        <button onClick={() => { setDraft(value); setEditing(false); }} className="p-0.5 rounded hover:bg-muted text-muted-foreground">
          <X className="size-3" />
        </button>
      </div>
    );
  }

  return (
    <span
      className={cn('truncate cursor-text', className)}
      onDoubleClick={(e) => { e.stopPropagation(); setDraft(value); setEditing(true); }}
      title="Double-click to rename"
    >
      {value}
    </span>
  );
}

// ─── Inline create row ────────────────────────────────────────────────────────

interface InlineCreateProps {
  placeholder: string;
  onSave: (name: string) => void;
  onCancel: () => void;
}

function InlineCreate({ placeholder, onSave, onCancel }: InlineCreateProps) {
  const [draft, setDraft] = useState('');

  const commit = () => {
    const trimmed = draft.trim();
    if (trimmed) onSave(trimmed);
  };

  return (
    <div className="flex items-center gap-1 px-1" onClick={(e) => e.stopPropagation()}>
      <Input
        autoFocus
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        placeholder={placeholder}
        onKeyDown={(e) => {
          if (e.key === 'Enter') commit();
          if (e.key === 'Escape') onCancel();
        }}
        className="h-5 text-xs px-1 py-0"
      />
      <button onClick={commit} disabled={!draft.trim()} className="p-0.5 rounded hover:bg-emerald-100 text-emerald-600 disabled:opacity-40">
        <Check className="size-3" />
      </button>
      <button onClick={onCancel} className="p-0.5 rounded hover:bg-muted text-muted-foreground">
        <X className="size-3" />
      </button>
    </div>
  );
}

// ─── Block row ────────────────────────────────────────────────────────────────

interface BlockRowProps {
  block: HierarchyBlock;
  allSuperBlocks: { id: number; name: string; siteId: number; siteName: string }[];
  isSelected: boolean;
  onSelect: () => void;
  designId: number;
  superBlockName?: string;
}

function BlockRow({ block, allSuperBlocks, isSelected, onSelect, designId, superBlockName }: BlockRowProps) {
  const queryClient = useQueryClient();

  const reparentMutation = useMutation({
    mutationFn: (targetId: number) => blocksApi.reparent(block.id, targetId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] });
      queryClient.invalidateQueries({ queryKey: ['blocks'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => blocksApi.delete(block.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] });
      queryClient.invalidateQueries({ queryKey: ['blocks'] });
    },
  });

  const otherSuperBlocks = allSuperBlocks.filter((sb) => sb.id !== block.super_block_id);

  return (
    <div
      className={cn(
        'flex items-center gap-1.5 px-2 py-1 rounded text-xs cursor-pointer group transition-colors',
        isSelected ? 'bg-primary/10 text-primary' : 'hover:bg-muted/50'
      )}
      onClick={onSelect}
    >
      <Cpu className="size-3 shrink-0 text-blue-500" />
      <span className="truncate flex-1">{block.name}</span>
      {superBlockName && (
        <span className="text-[9px] text-muted-foreground bg-muted px-1 rounded shrink-0">
          {superBlockName}
        </span>
      )}

      <div className="opacity-0 group-hover:opacity-100 transition-opacity" onClick={(e) => e.stopPropagation()}>
        <DropdownMenu>
          <DropdownMenuTrigger className="p-0.5 rounded hover:bg-muted inline-flex items-center">
            <MoreHorizontal className="size-3 text-muted-foreground" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="text-xs">
            {otherSuperBlocks.length > 0 && (
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>Move to super-block…</DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                  {otherSuperBlocks.map((sb) => (
                    <DropdownMenuItem
                      key={sb.id}
                      onClick={() => reparentMutation.mutate(sb.id)}
                      disabled={reparentMutation.isPending}
                    >
                      <span className="text-muted-foreground mr-1">{sb.siteName} /</span>
                      {sb.name}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuSubContent>
              </DropdownMenuSub>
            )}
            {otherSuperBlocks.length > 0 && <DropdownMenuSeparator />}
            <DropdownMenuItem
              className="text-destructive focus:text-destructive"
              onClick={() => deleteMutation.mutate()}
              disabled={deleteMutation.isPending}
            >
              <Trash2 className="size-3 mr-1" />
              Delete block
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}

// ─── SuperBlock section ───────────────────────────────────────────────────────

interface SuperBlockSectionProps {
  sb: HierarchySuperBlock;
  allSuperBlocks: { id: number; name: string; siteId: number; siteName: string }[];
  selectedBlockId: number | null;
  onSelectBlock: (id: number) => void;
  designId: number;
}

function SuperBlockSection({ sb, allSuperBlocks, selectedBlockId, onSelectBlock, designId }: SuperBlockSectionProps) {
  const [expanded, setExpanded] = useState(true);
  const [addingBlock, setAddingBlock] = useState(false);
  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: (name: string) => superblocksApi.update(sb.id, { name }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] }),
  });

  const deleteMutation = useMutation({
    mutationFn: () => superblocksApi.delete(sb.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] }),
  });

  const createBlockMutation = useMutation({
    mutationFn: (name: string) => blocksApi.create({ super_block_id: sb.id, name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] });
      queryClient.invalidateQueries({ queryKey: ['blocks'] });
      setAddingBlock(false);
    },
  });

  return (
    <div className="ml-4 mt-0.5">
      {/* SuperBlock header */}
      <div className="flex items-center gap-1 group rounded px-1 py-0.5 hover:bg-muted/40">
        <button
          className="p-0.5 rounded hover:bg-muted"
          onClick={() => setExpanded((v) => !v)}
        >
          {expanded ? (
            <ChevronDown className="size-3 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-3 text-muted-foreground" />
          )}
        </button>
        <Layers2 className="size-3 shrink-0 text-violet-500" />
        <InlineLabel
          value={sb.name}
          onSave={(name) => updateMutation.mutate(name)}
          className="text-xs font-medium flex-1"
        />
        <span className="text-[10px] text-muted-foreground/60 mr-1">{sb.blocks.length}</span>

        <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center">
          <button
            className="p-0.5 rounded hover:bg-muted text-muted-foreground"
            title="Add block"
            onClick={(e) => { e.stopPropagation(); setAddingBlock(true); setExpanded(true); }}
          >
            <Plus className="size-3" />
          </button>
          <button
            className="p-0.5 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive"
            title="Delete super-block"
            onClick={(e) => { e.stopPropagation(); deleteMutation.mutate(); }}
            disabled={deleteMutation.isPending}
          >
            <Trash2 className="size-3" />
          </button>
        </div>
      </div>

      {/* Blocks */}
      {expanded && (
        <div className="ml-3 mt-0.5 space-y-0.5">
          {sb.blocks.map((block) => (
            <BlockRow
              key={block.id}
              block={block}
              allSuperBlocks={allSuperBlocks}
              isSelected={selectedBlockId === block.id}
              onSelect={() => onSelectBlock(block.id)}
              designId={designId}
            />
          ))}
          {addingBlock && (
            <InlineCreate
              placeholder="Block name…"
              onSave={(name) => createBlockMutation.mutate(name)}
              onCancel={() => setAddingBlock(false)}
            />
          )}
          {sb.blocks.length === 0 && !addingBlock && (
            <p className="text-[10px] text-muted-foreground/50 px-2 py-0.5 italic">No blocks</p>
          )}
        </div>
      )}
    </div>
  );
}

// ─── Site section ─────────────────────────────────────────────────────────────

interface SiteSectionProps {
  site: HierarchySite;
  allSuperBlocks: { id: number; name: string; siteId: number; siteName: string }[];
  selectedBlockId: number | null;
  onSelectBlock: (id: number) => void;
  designId: number;
}

function SiteSection({ site, allSuperBlocks, selectedBlockId, onSelectBlock, designId }: SiteSectionProps) {
  const [expanded, setExpanded] = useState(true);
  const [addingSuperBlock, setAddingSuperBlock] = useState(false);
  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: (name: string) => sitesApi.update(site.id, { name }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] }),
  });

  const deleteMutation = useMutation({
    mutationFn: () => sitesApi.delete(site.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] }),
  });

  const createSuperBlockMutation = useMutation({
    mutationFn: (name: string) => superblocksApi.create(site.id, { name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] });
      setAddingSuperBlock(false);
    },
  });

  return (
    <div className="mt-1">
      {/* Site header */}
      <div className="flex items-center gap-1 group rounded px-1 py-0.5 hover:bg-muted/40">
        <button
          className="p-0.5 rounded hover:bg-muted"
          onClick={() => setExpanded((v) => !v)}
        >
          {expanded ? (
            <ChevronDown className="size-3.5 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-3.5 text-muted-foreground" />
          )}
        </button>
        <Building2 className="size-3.5 shrink-0 text-emerald-500" />
        <InlineLabel
          value={site.name}
          onSave={(name) => updateMutation.mutate(name)}
          className="text-xs font-semibold flex-1"
        />

        <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5">
          <button
            className="p-0.5 rounded hover:bg-muted text-muted-foreground"
            title="Add super-block"
            onClick={(e) => { e.stopPropagation(); setAddingSuperBlock(true); setExpanded(true); }}
          >
            <Plus className="size-3" />
          </button>
          <button
            className="p-0.5 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive"
            title="Delete site"
            onClick={(e) => { e.stopPropagation(); deleteMutation.mutate(); }}
            disabled={deleteMutation.isPending}
          >
            <Trash2 className="size-3" />
          </button>
        </div>
      </div>

      {/* Super-blocks */}
      {expanded && (
        <div>
          {site.super_blocks.map((sb) => (
            <SuperBlockSection
              key={sb.id}
              sb={sb}
              allSuperBlocks={allSuperBlocks}
              selectedBlockId={selectedBlockId}
              onSelectBlock={onSelectBlock}
              designId={designId}
            />
          ))}
          {addingSuperBlock && (
            <div className="ml-4 mt-0.5">
              <InlineCreate
                placeholder="Super-block name…"
                onSave={(name) => createSuperBlockMutation.mutate(name)}
                onCancel={() => setAddingSuperBlock(false)}
              />
            </div>
          )}
          {site.super_blocks.length === 0 && !addingSuperBlock && (
            <p className="text-[10px] text-muted-foreground/50 px-5 py-0.5 italic">No super-blocks</p>
          )}
        </div>
      )}
    </div>
  );
}

// ─── HierarchyTree ────────────────────────────────────────────────────────────

interface HierarchyTreeProps {
  designId: number;
  selectedBlockId: number | null;
  onSelectBlock: (id: number) => void;
}

export default function HierarchyTree({ designId, selectedBlockId, onSelectBlock }: HierarchyTreeProps) {
  const [addingSite, setAddingSite] = useState(false);
  const queryClient = useQueryClient();

  const { data: hierarchy, isLoading } = useQuery<DesignHierarchy>({
    queryKey: ['hierarchy', designId],
    queryFn: () => sitesApi.getHierarchy(designId),
    refetchInterval: 30_000,
    enabled: !!designId,
  });

  const createSiteMutation = useMutation({
    mutationFn: (name: string) => sitesApi.create(designId, { name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hierarchy', designId] });
      setAddingSite(false);
    },
  });

  if (isLoading) {
    return (
      <div className="px-3 py-4 text-xs text-muted-foreground text-center">Loading…</div>
    );
  }

  const sites = hierarchy?.sites ?? [];

  // Flatten all super-blocks for the "move to" target picker
  const allSuperBlocks = sites.flatMap((site) =>
    site.super_blocks.map((sb) => ({
      id: sb.id,
      name: sb.name,
      siteId: site.id,
      siteName: site.name,
    }))
  );

  return (
    <div className="flex flex-col">
      {/* Section header */}
      <div className="flex items-center justify-between px-3 py-2">
        <span className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
          Hierarchy
        </span>
        <Button
          variant="ghost"
          size="sm"
          className="h-5 w-5 p-0"
          title="Add site"
          onClick={() => setAddingSite(true)}
        >
          <Plus className="size-3" />
        </Button>
      </div>

      <div className="px-2 pb-4 space-y-0.5">
        {sites.map((site) => (
          <SiteSection
            key={site.id}
            site={site}
            allSuperBlocks={allSuperBlocks}
            selectedBlockId={selectedBlockId}
            onSelectBlock={onSelectBlock}
            designId={designId}
          />
        ))}

        {addingSite && (
          <div className="mt-1">
            <InlineCreate
              placeholder="Site name…"
              onSave={(name) => createSiteMutation.mutate(name)}
              onCancel={() => setAddingSite(false)}
            />
          </div>
        )}

        {sites.length === 0 && !addingSite && (
          <div className="px-2 py-3 text-center">
            <Building2 className="size-6 text-muted-foreground/30 mx-auto mb-1.5" />
            <p className="text-[10px] text-muted-foreground/60">
              No sites yet. Add one to organize your blocks.
            </p>
            <Button
              variant="outline"
              size="sm"
              className="mt-2 text-xs h-6"
              onClick={() => setAddingSite(true)}
            >
              <Plus className="size-3 mr-1" />
              Add Site
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
