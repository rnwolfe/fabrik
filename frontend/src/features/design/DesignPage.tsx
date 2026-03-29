import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
  Plus,
  Server,
  ChevronDown,
  Layers,
} from 'lucide-react';
import { blocksApi, scaffoldApi } from '@/api/blocks';
import { racksApi } from '@/api/racks';
import { catalogApi } from '@/api/catalog';
import { designsApi } from '@/api/designs';
import { useDesign } from '@/contexts/DesignContext';
import { EmptyState } from '@/components/EmptyState';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import DeviceModelPicker, { PickerGuardProvider, usePickerGuard } from '@/components/DeviceModelPicker';
import BlockCard from './BlockCard';
import BlockDetailPanel from './BlockDetailPanel';
import DesignSummary from './DesignSummary';
import type { Block, BlockAggregationSummary, RackSummary, DeviceModel } from '@/models';

// ─── New Block form schema ───────────────────────────────────────────────────

const newBlockSchema = z.object({
  name: z.string().min(1, 'Name required'),
  leaf_model_id: z.preprocess(
    (v) => (v === '' || v == null ? undefined : Number(v)),
    z.number().int().optional()
  ),
});
type NewBlockForm = z.infer<typeof newBlockSchema>;

// ─── Add Rack form ───────────────────────────────────────────────────────────

const addRackSchema = z.object({
  name: z.string().min(1, 'Name required'),
  rack_type_id: z.preprocess(
    (v) => (v === '' || v == null ? undefined : Number(v)),
    z.number().int().optional()
  ),
});
type AddRackForm = z.infer<typeof addRackSchema>;

// ─── Main page ───────────────────────────────────────────────────────────────

export default function DesignPage() {
  const { activeDesignId, setActiveDesignId } = useDesign();
  const queryClient = useQueryClient();

  // Selection state
  const [selectedBlockId, setSelectedBlockId] = useState<number | null>(null);
  const [selectedRackId, setSelectedRackId] = useState<number | null>(null);

  // Dialog state
  const [blockDialogOpen, setBlockDialogOpen] = useState(false);
  const [rackDialogBlockId, setRackDialogBlockId] = useState<number | null>(null);

  // Per-block spine count (local state until we persist it)
  const [blockSpineCount, setBlockSpineCount] = useState<Map<number, number>>(new Map());
  // Per-block spine model (local state — not yet persisted in backend)
  const [blockSpineModel, setBlockSpineModel] = useState<Map<number, number>>(new Map());

  // ── Data queries ─────────────────────────────────────────────────────────

  const { data: designs } = useQuery({ queryKey: ['designs'], queryFn: designsApi.list });
  const { data: catalogDevices } = useQuery({ queryKey: ['catalog'], queryFn: catalogApi.list });
  const { data: rackTypes } = useQuery({ queryKey: ['rackTypes'], queryFn: racksApi.listTypes });

  const networkDevices = (catalogDevices ?? []).filter(
    (d: DeviceModel) => d.device_model_type === 'network' && !d.archived_at
  );

  // Get scaffold (default site + super-block) for the active design
  const { data: scaffold } = useQuery({
    queryKey: ['scaffold', activeDesignId],
    queryFn: () => scaffoldApi.get(activeDesignId!),
    enabled: !!activeDesignId,
  });

  const superBlockId = scaffold?.super_block_id;

  // List blocks for this design's super-block
  const { data: blocks } = useQuery({
    queryKey: ['blocks', superBlockId],
    queryFn: () => blocksApi.list(superBlockId!),
    enabled: !!superBlockId,
  });

  // List all racks (unfiltered — we'll group by block_id client-side)
  const { data: allRacks } = useQuery({
    queryKey: ['racks'],
    queryFn: () => racksApi.list(),
  });

  // List aggregations for each block
  const blockIds = (blocks ?? []).map((b: Block) => b.id);
  const { data: aggsByBlock } = useQuery({
    queryKey: ['aggs', ...blockIds],
    queryFn: async () => {
      const map = new Map<number, BlockAggregationSummary[]>();
      for (const id of blockIds) {
        try {
          const aggs = await blocksApi.listAggregations(id);
          map.set(id, aggs);
        } catch {
          map.set(id, []);
        }
      }
      return map;
    },
    enabled: blockIds.length > 0,
  });

  // Group racks by block_id
  const racksByBlock = new Map<number, RackSummary[]>();
  for (const rack of allRacks ?? []) {
    const bid = rack.block_id;
    if (bid) {
      const arr = racksByBlock.get(bid) ?? [];
      arr.push(rack);
      racksByBlock.set(bid, arr);
    }
  }

  // Auto-select first block when blocks load. Calling setState inside an effect
  // is intentional here — we're syncing UI selection state with fetched data.
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if ((blocks ?? []).length > 0 && !selectedBlockId) {
      setSelectedBlockId(blocks![0].id);
    }
    if ((blocks ?? []).length === 0) {
      setSelectedBlockId(null);
      setSelectedRackId(null);
    }
  }, [blocks, selectedBlockId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const selectedBlock = (blocks ?? []).find((b: Block) => b.id === selectedBlockId) ?? null;

  // ── Mutations ────────────────────────────────────────────────────────────

  const createBlockMutation = useMutation({
    mutationFn: (data: { super_block_id: number; name: string }) => blocksApi.create(data),
    onSuccess: (block) => {
      queryClient.invalidateQueries({ queryKey: ['blocks'] });
      setSelectedBlockId(block.id);
      setBlockDialogOpen(false);
      blockReset();
    },
  });

  const addRackMutation = useMutation({
    mutationFn: async ({ blockId, name, rackTypeId }: { blockId: number; name: string; rackTypeId?: number }) => {
      // First create the rack
      const rt = rackTypeId ? rackTypes?.find((t) => t.id === rackTypeId) : null;
      const rack = await racksApi.create({
        name,
        rack_type_id: rackTypeId,
        height_u: rt?.height_u ?? 42,
        power_capacity_w: rt?.power_capacity_w ?? 10000,
      });
      // Then add it to the block
      await blocksApi.addRack({ rack_id: rack.id, block_id: blockId });
      return rack;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['racks'] });
      queryClient.invalidateQueries({ queryKey: ['aggs'] });
      setRackDialogBlockId(null);
      rackReset();
    },
  });

  const removeRackMutation = useMutation({
    mutationFn: (rackId: number) => blocksApi.removeRack(rackId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['racks'] });
      queryClient.invalidateQueries({ queryKey: ['aggs'] });
      setSelectedRackId(null);
    },
  });

  const assignAggMutation = useMutation({
    mutationFn: ({ blockId, deviceModelId }: { blockId: number; deviceModelId: number }) =>
      blocksApi.assignAggregation(blockId, 'front_end', deviceModelId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['aggs'] });
    },
  });

  // ── Forms ────────────────────────────────────────────────────────────────

  const {
    register: blockRegister,
    handleSubmit: blockHandleSubmit,
    reset: blockReset,
    setValue: blockSetValue,
  } = useForm<NewBlockForm>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(newBlockSchema) as any,
  });

  const {
    register: rackRegister,
    handleSubmit: rackHandleSubmit,
    reset: rackReset,
  } = useForm<AddRackForm>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(addRackSchema) as any,
  });

  // ── Handlers ─────────────────────────────────────────────────────────────

  const handleCreateBlock = blockHandleSubmit((data) => {
    if (!superBlockId) return;
    createBlockMutation.mutate({
      super_block_id: superBlockId,
      name: data.name,
    });
    // If a leaf model was selected, assign it as the frontend aggregation
    if (data.leaf_model_id) {
      // We'll do this after block creation in onSuccess
    }
  });

  const handleAddRack = rackHandleSubmit((data) => {
    if (!rackDialogBlockId) return;
    addRackMutation.mutate({
      blockId: rackDialogBlockId,
      name: data.name,
      rackTypeId: data.rack_type_id,
    });
  });

  const handleAssignSpine = useCallback(
    (blockId: number, deviceModelId: number) => {
      setBlockSpineModel((prev) => new Map(prev).set(blockId, deviceModelId));
      // Also assign as frontend aggregation so port tracking works
      assignAggMutation.mutate({ blockId, deviceModelId });
    },
    [assignAggMutation]
  );

  const handleSpineCountChange = useCallback((blockId: number, value: number) => {
    setBlockSpineCount((prev) => new Map(prev).set(blockId, value));
  }, []);

  // ── No design selected ─────────────────────────────────────────────────

  if (!activeDesignId) {
    return (
      <div className="flex h-full items-center justify-center">
        <EmptyState
          icon={Server}
          title="No design selected"
          description="Select or create a design from the dashboard to start building your fabric topology."
          action={
            designs && designs.length > 0 ? (
              <div className="flex flex-col items-center gap-2">
                <p className="text-xs text-muted-foreground">Quick select:</p>
                <div className="flex flex-wrap justify-center gap-2">
                  {designs.slice(0, 5).map((d) => (
                    <Button key={d.id} variant="outline" size="sm" onClick={() => setActiveDesignId(d.id)}>
                      {d.name}
                    </Button>
                  ))}
                </div>
              </div>
            ) : undefined
          }
        />
      </div>
    );
  }

  const activeDesign = designs?.find((d) => d.id === activeDesignId);

  // ── Main layout ────────────────────────────────────────────────────────

  return (
    <div className="flex h-full flex-col -m-6 overflow-hidden">
      {/* Top bar */}
      <div className="flex shrink-0 items-center gap-3 border-b border-border bg-background/95 px-4 py-2 backdrop-blur">
        {designs && designs.length > 1 ? (
          <Select value={String(activeDesignId)} onValueChange={(v) => v && setActiveDesignId(Number(v))}>
            <SelectTrigger className="h-7 w-auto gap-1.5 border-none bg-muted/60 px-2.5 text-sm font-medium shadow-none">
              <SelectValue>
                {(value: string) => designs?.find((d) => String(d.id) === value)?.name ?? 'Design'}
              </SelectValue>
              <ChevronDown className="size-3.5" />
            </SelectTrigger>
            <SelectContent>
              {designs.map((d) => (
                <SelectItem key={d.id} value={String(d.id)}>
                  {d.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <span className="text-sm font-medium">{activeDesign?.name ?? 'Design'}</span>
        )}

        <div className="ml-auto flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="h-7 text-xs gap-1"
            onClick={() => {
              blockReset({ name: `Block ${(blocks ?? []).length + 1}` });
              setBlockDialogOpen(true);
            }}
          >
            <Plus className="size-3" />
            New Block
          </Button>
        </div>
      </div>

      {/* Two-panel layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left panel: Block canvas */}
        <div className="flex flex-col w-80 shrink-0 border-r border-border overflow-y-auto bg-muted/20">
          <div className="px-3 py-2">
            <span className="text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
              Blocks
            </span>
          </div>

          {(blocks ?? []).length === 0 ? (
            <div className="flex flex-1 items-center justify-center p-4">
              <div className="text-center">
                <Layers className="size-8 text-muted-foreground/40 mx-auto mb-2" />
                <p className="text-xs text-muted-foreground">
                  No blocks yet. Create one to start your design.
                </p>
                <Button
                  variant="outline"
                  size="sm"
                  className="mt-3 text-xs"
                  onClick={() => {
                    blockReset({ name: 'Block 1' });
                    setBlockDialogOpen(true);
                  }}
                >
                  <Plus className="size-3 mr-1" />
                  Create Block
                </Button>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-2 px-2 pb-4">
              {(blocks ?? []).map((block: Block) => (
                <BlockCard
                  key={block.id}
                  block={block}
                  aggs={aggsByBlock?.get(block.id) ?? []}
                  racks={racksByBlock.get(block.id) ?? []}
                  isSelected={selectedBlockId === block.id}
                  selectedRackId={selectedRackId}
                  onSelect={() => {
                    setSelectedBlockId(block.id);
                    setSelectedRackId(null);
                  }}
                  onSelectRack={(rackId) => {
                    setSelectedBlockId(block.id);
                    setSelectedRackId(rackId);
                  }}
                  onAddRack={() => {
                    const racks = racksByBlock.get(block.id) ?? [];
                    rackReset({ name: `Rack ${racks.length + 1}` });
                    setRackDialogBlockId(block.id);
                  }}
                  onRemoveRack={(rackId) => removeRackMutation.mutate(rackId)}
                  onDelete={() => {
                    // TODO: add block delete API
                  }}
                />
              ))}
            </div>
          )}
        </div>

        {/* Right panel: Detail */}
        <div className="flex-1 overflow-y-auto">
          {selectedBlock ? (
            <BlockDetailPanel
              block={selectedBlock}
              aggs={aggsByBlock?.get(selectedBlock.id) ?? []}
              racks={racksByBlock.get(selectedBlock.id) ?? []}
              networkDevices={networkDevices}
              spineModelId={blockSpineModel.get(selectedBlock.id)}
              spineCount={blockSpineCount.get(selectedBlock.id) ?? null}
              onSpineCountChange={(v) => handleSpineCountChange(selectedBlock.id, v)}
              onAssignSpine={(id) => handleAssignSpine(selectedBlock.id, id)}
            />
          ) : (
            <DesignSummary
              blocks={blocks ?? []}
              racksByBlock={racksByBlock}
            />
          )}
        </div>
      </div>

      {/* ── New Block Dialog ─────────────────────────────────────────────── */}
      <PickerGuardProvider>
        <NewBlockDialogInner
          open={blockDialogOpen}
          onOpenChange={setBlockDialogOpen}
          onSubmit={handleCreateBlock}
          register={blockRegister}
          isPending={createBlockMutation.isPending}
          networkDevices={networkDevices}
          onSelectLeaf={(id) => blockSetValue('leaf_model_id', id)}
        />
      </PickerGuardProvider>

      {/* ── Add Rack Dialog ──────────────────────────────────────────────── */}
      <Dialog open={rackDialogBlockId !== null} onOpenChange={(open) => !open && setRackDialogBlockId(null)}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Add Rack</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleAddRack} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="rack-name">Name</Label>
              <Input id="rack-name" {...rackRegister('name')} />
            </div>
            {(rackTypes ?? []).length > 0 && (
              <div className="space-y-1.5">
                <Label>Rack Template</Label>
                <Select onValueChange={(v) => rackReset({ rack_type_id: Number(v) } as Partial<AddRackForm>)}>
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select template (optional)…">
                      {(value: string) => {
                        const rt = (rackTypes ?? []).find((t) => String(t.id) === value);
                        return rt ? `${rt.name} (${rt.height_u}U, ${rt.power_capacity_w}W)` : 'Select template (optional)…';
                      }}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {(rackTypes ?? []).map((rt) => (
                      <SelectItem key={rt.id} value={String(rt.id)}>
                        {rt.name} ({rt.height_u}U, {rt.power_capacity_w}W)
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setRackDialogBlockId(null)}>
                Cancel
              </Button>
              <Button type="submit" disabled={addRackMutation.isPending}>
                {addRackMutation.isPending ? 'Adding…' : 'Add Rack'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Extracted so usePickerGuard hook can be called inside PickerGuardProvider
function NewBlockDialogInner({
  open,
  onOpenChange,
  onSubmit,
  register,
  isPending,
  networkDevices,
  onSelectLeaf,
}: {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  onSubmit: (e: React.FormEvent) => void;
  register: ReturnType<typeof useForm<NewBlockForm>>['register'];
  isPending: boolean;
  networkDevices: DeviceModel[];
  onSelectLeaf: (id: number) => void;
}) {
  const pickerOpen = usePickerGuard();

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        // Don't close the dialog while the picker popover is open
        if (!v && pickerOpen) return;
        onOpenChange(v);
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New Block</DialogTitle>
        </DialogHeader>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="block-name">Name</Label>
            <Input id="block-name" {...register('name')} />
          </div>
          <div className="space-y-1.5">
            <Label>Leaf Switch</Label>
            <DeviceModelPicker
              devices={networkDevices}
              onSelect={onSelectLeaf}
              placeholder="Select leaf model…"
            />
            <p className="text-[11px] text-muted-foreground">
              The leaf switch model determines port allocation and oversubscription.
            </p>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? 'Creating…' : 'Create Block'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
