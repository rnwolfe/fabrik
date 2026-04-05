import { useState, useRef, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Server, Zap, Box, Minus, Plus, Cpu, ChevronLeft, GripVertical } from 'lucide-react';
import { racksApi } from '@/api/racks';
import { catalogApi } from '@/api/catalog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import DeviceModelPicker from '@/components/DeviceModelPicker';
import { cn } from '@/lib/utils';
import type { DeviceModel, DeviceSummary } from '@/models';

// Role → display color
const ROLE_COLORS: Record<string, string> = {
  leaf: 'bg-teal-600 text-white',
  spine: 'bg-blue-600 text-white',
  server: 'bg-slate-500 text-white',
  management_tor: 'bg-purple-600 text-white',
  management_agg: 'bg-purple-700 text-white',
  other: 'bg-gray-500 text-white',
};

const ROLE_LABELS: Record<string, string> = {
  leaf: 'Leaf',
  spine: 'Spine',
  server: 'Server',
  management_tor: 'Mgmt ToR',
  management_agg: 'Mgmt Agg',
  other: 'Other',
};

interface RackElevationProps {
  rackId: number;
}

// ─── Device detail panel ──────────────────────────────────────────────────────

function DeviceDetail({ device, onBack }: { device: DeviceSummary; onBack: () => void }) {
  const colorClass = ROLE_COLORS[device.role] ?? ROLE_COLORS.other;
  return (
    <div className="space-y-4">
      <button
        onClick={onBack}
        className="flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground transition-colors"
      >
        <ChevronLeft className="size-3" />
        Back to rack
      </button>

      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <Cpu className="size-4 text-muted-foreground" />
          <span className="text-sm font-semibold">{device.name}</span>
        </div>
        <Badge className={cn('text-[10px]', colorClass)}>{ROLE_LABELS[device.role] ?? device.role}</Badge>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div className="rounded-md bg-muted/50 px-2.5 py-2">
          <p className="text-[10px] text-muted-foreground">Position</p>
          <p className="text-sm font-semibold font-mono">U{device.position}</p>
        </div>
        <div className="rounded-md bg-muted/50 px-2.5 py-2">
          <p className="text-[10px] text-muted-foreground">Height</p>
          <p className="text-sm font-semibold font-mono">{device.height_u}U</p>
        </div>
        {device.power_watts_typical > 0 && (
          <div className="rounded-md bg-muted/50 px-2.5 py-2">
            <p className="text-[10px] text-muted-foreground">Power (typ)</p>
            <p className="text-sm font-semibold font-mono">{device.power_watts_typical}W</p>
          </div>
        )}
        {device.power_watts_max > 0 && (
          <div className="rounded-md bg-muted/50 px-2.5 py-2">
            <p className="text-[10px] text-muted-foreground">Power (max)</p>
            <p className="text-sm font-semibold font-mono">{device.power_watts_max}W</p>
          </div>
        )}
      </div>

      {(device.model_vendor || device.model_name) && (
        <div className="rounded-md border bg-muted/10 px-3 py-2 space-y-0.5">
          <p className="text-[10px] text-muted-foreground uppercase tracking-wider">Model</p>
          {device.model_vendor && <p className="text-xs text-muted-foreground">{device.model_vendor}</p>}
          {device.model_name && <p className="text-sm font-medium">{device.model_name}</p>}
        </div>
      )}
    </div>
  );
}

// ─── Rack config panel (default right-side content) ───────────────────────────

function RackConfig({
  rack,
  serverModels,
}: {
  rack: { role: string; height_u: number; used_u: number; available_u: number; power_capacity_w: number; used_watts_typical: number; devices?: DeviceSummary[] };
  serverModels: DeviceModel[];
}) {
  const queryClient = useQueryClient();
  const [serverModelId, setServerModelId] = useState<number | undefined>();
  const [serverCount, setServerCount] = useState(1);

  const selectedServerModel = serverModels.find((d) => d.id === serverModelId);
  const maxServers = selectedServerModel
    ? Math.floor(
        (rack.available_u +
          (rack.devices ?? []).filter((d) => d.role === 'server').reduce((s, d) => s + d.height_u, 0)) /
          selectedServerModel.height_u
      )
    : 0;

  const placeServersMutation = useMutation({
    mutationFn: ({ modelId, count }: { modelId: number; count: number }) =>
      racksApi.placeServerDevices((rack as any).id, modelId, count),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['racks'] });
      queryClient.invalidateQueries({ queryKey: ['rack', (rack as any).id] });
    },
  });

  const powerPct =
    rack.power_capacity_w > 0
      ? Math.min(100, Math.round((rack.used_watts_typical / rack.power_capacity_w) * 100))
      : 0;

  return (
    <div className="space-y-4">
      {/* Stats */}
      <div className="grid grid-cols-3 gap-2">
        <div className="rounded-md bg-muted/50 px-2.5 py-1.5 text-center">
          <p className="text-[10px] text-muted-foreground">Used</p>
          <p className="text-sm font-semibold font-mono">{rack.used_u}/{rack.height_u}U</p>
        </div>
        <div className="rounded-md bg-muted/50 px-2.5 py-1.5 text-center">
          <p className="text-[10px] text-muted-foreground">Free</p>
          <p className="text-sm font-semibold font-mono">{rack.available_u}U</p>
        </div>
        <div
          className={cn(
            'rounded-md px-2.5 py-1.5 text-center',
            powerPct > 90 ? 'bg-red-500/10' : powerPct > 75 ? 'bg-amber-500/10' : 'bg-muted/50'
          )}
        >
          <p className="text-[10px] text-muted-foreground flex items-center justify-center gap-0.5">
            <Zap className="size-2.5" />
            Power
          </p>
          <p
            className={cn(
              'text-sm font-semibold font-mono',
              powerPct > 90 ? 'text-red-600' : powerPct > 75 ? 'text-amber-600' : ''
            )}
          >
            {powerPct}%
          </p>
        </div>
      </div>

      {/* Server placement — compute racks only */}
      {rack.role === 'compute' && (
        <div className="space-y-3 border-t pt-3">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">Server Placement</p>

          <div className="space-y-1.5">
            <Label className="text-xs">Server Model</Label>
            <DeviceModelPicker
              devices={serverModels}
              value={serverModelId}
              onSelect={(id) => {
                setServerModelId(id);
                setServerCount(1);
              }}
              placeholder="Select server model…"
              triggerClassName="h-8 text-xs"
            />
          </div>

          {serverModelId && (
            <div className="space-y-1.5">
              <Label className="text-xs flex items-center gap-1">
                <Server className="size-3" />
                Server Count
              </Label>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="icon"
                  className="size-7"
                  disabled={serverCount <= 0}
                  onClick={() => setServerCount((c) => Math.max(0, c - 1))}
                >
                  <Minus className="size-3" />
                </Button>
                <span className="w-8 text-center font-mono text-sm font-semibold">{serverCount}</span>
                <Button
                  variant="outline"
                  size="icon"
                  className="size-7"
                  disabled={serverCount >= maxServers}
                  onClick={() => setServerCount((c) => c + 1)}
                >
                  <Plus className="size-3" />
                </Button>
                <span className="text-[10px] text-muted-foreground">/ {maxServers} max</span>
              </div>
            </div>
          )}

          {serverModelId && (
            <Button
              size="sm"
              className="w-full text-xs"
              disabled={placeServersMutation.isPending}
              onClick={() => {
                if (!serverModelId) return;
                placeServersMutation.mutate({ modelId: serverModelId, count: serverCount });
              }}
            >
              {placeServersMutation.isPending
                ? 'Placing…'
                : `Place ${serverCount} Server${serverCount !== 1 ? 's' : ''}`}
            </Button>
          )}

          {serverModels.length === 0 && (
            <p className="text-xs text-muted-foreground text-center py-2">
              No server models in catalog. Add a server model to place servers.
            </p>
          )}
        </div>
      )}

      {/* Hint */}
      <p className="text-[10px] text-muted-foreground/50 text-center pt-2">
        Click a device in the rack to view its details.
      </p>
    </div>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

const MIN_VIZ_WIDTH = 140;
const MAX_VIZ_WIDTH = 480;
const DEFAULT_VIZ_WIDTH = 220;

export default function RackElevation({ rackId }: RackElevationProps) {
  const [selectedDevice, setSelectedDevice] = useState<DeviceSummary | null>(null);
  const [vizWidth, setVizWidth] = useState(DEFAULT_VIZ_WIDTH);
  const dragRef = useRef<{ startX: number; startWidth: number } | null>(null);

  const onDragStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    dragRef.current = { startX: e.clientX, startWidth: vizWidth };
    const onMove = (e: MouseEvent) => {
      if (!dragRef.current) return;
      const next = Math.max(MIN_VIZ_WIDTH, Math.min(MAX_VIZ_WIDTH, dragRef.current.startWidth + e.clientX - dragRef.current.startX));
      setVizWidth(next);
    };
    const onUp = () => {
      dragRef.current = null;
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    };
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  }, [vizWidth]);

  const { data: rack, isLoading } = useQuery({
    queryKey: ['rack', rackId],
    queryFn: () => racksApi.get(rackId),
  });

  const { data: catalog } = useQuery({
    queryKey: ['catalog'],
    queryFn: catalogApi.list,
  });

  const serverModels = (catalog ?? []).filter(
    (d: DeviceModel) => d.device_model_type === 'server' && !d.archived_at
  );

  // Clear device selection when switching racks.
  if (rack && selectedDevice && !(rack.devices ?? []).some((d) => d.id === selectedDevice.id)) {
    setSelectedDevice(null);
  }

  if (isLoading) {
    return (
      <div className="flex h-full">
        <div className="shrink-0 border-r animate-pulse bg-muted/20" style={{ width: vizWidth }} />
        <div className="flex-1 p-4 space-y-3">
          <div className="h-5 w-32 animate-pulse rounded bg-muted" />
          <div className="grid grid-cols-3 gap-2">
            {[1, 2, 3].map((i) => <div key={i} className="h-12 animate-pulse rounded-md bg-muted" />)}
          </div>
        </div>
      </div>
    );
  }
  if (!rack) return null;

  // Build a map of top-slot → device.
  // position = bottom U slot; top slot = position + height_u - 1.
  const devicesByTopSlot = new Map<number, DeviceSummary>();
  for (const dev of rack.devices ?? []) {
    devicesByTopSlot.set(dev.position + dev.height_u - 1, dev);
  }

  const heightU = rack.height_u;

  return (
    <div className="flex h-full overflow-hidden">
      {/* ── Left: Rack visualization ─────────────────────────────────────── */}
      <div className="shrink-0 border-r flex flex-col overflow-hidden bg-muted/10" style={{ width: vizWidth }}>
        {/* Top label */}
        <div className="border-b border-border px-2 py-1 bg-muted/30 shrink-0">
          <p className="text-[9px] font-medium text-muted-foreground uppercase tracking-wider text-center">
            Top of Rack
          </p>
        </div>

        {/* Slot column — flex fills height, each slot gets proportional space */}
        <div className="flex-1 flex flex-col min-h-0">
          {Array.from({ length: heightU }, (_, i) => heightU - i).map((u) => {
            const dev = devicesByTopSlot.get(u);

            if (dev) {
              const colorClass = ROLE_COLORS[dev.role] ?? ROLE_COLORS.other;
              const isSelected = selectedDevice?.id === dev.id;
              return (
                <button
                  key={u}
                  style={{ flexGrow: dev.height_u, flexShrink: dev.height_u, flexBasis: 0 }}
                  className={cn(
                    'flex items-center px-1.5 gap-1 border-b border-border/40 font-mono text-[9px] font-medium w-full text-left transition-opacity',
                    colorClass,
                    isSelected ? 'ring-2 ring-inset ring-white/60' : 'hover:opacity-90'
                  )}
                  onClick={() => setSelectedDevice(isSelected ? null : dev)}
                >
                  <span className="opacity-60 shrink-0 w-3 text-right">{u}</span>
                  <span className="truncate">{dev.name}</span>
                </button>
              );
            }

            // Check if covered by a multi-U device above
            const coveredBy = Array.from(devicesByTopSlot.entries()).find(
              ([topSlot, d]) => topSlot > u && d.position <= u
            );
            if (coveredBy) return null;

            return (
              <div
                key={u}
                style={{ flexGrow: 1, flexShrink: 1, flexBasis: 0 }}
                className="flex items-center px-1.5 gap-1 border-b border-border/20"
              >
                <span className="text-[8px] text-muted-foreground/30 font-mono w-3 text-right shrink-0">{u}</span>
                <div className="flex-1 border-b border-dashed border-border/20" />
              </div>
            );
          })}
        </div>

        {/* Bottom label */}
        <div className="border-t border-border px-2 py-1 bg-muted/30 shrink-0">
          <p className="text-[9px] font-medium text-muted-foreground uppercase tracking-wider text-center">
            Bottom of Rack
          </p>
        </div>
      </div>

      {/* ── Resize handle ────────────────────────────────────────────────── */}
      <div
        onMouseDown={onDragStart}
        className="w-1 shrink-0 hover:w-1.5 bg-border hover:bg-primary/40 cursor-col-resize flex items-center justify-center transition-all group"
        title="Drag to resize"
      >
        <GripVertical className="size-3 text-muted-foreground/30 group-hover:text-primary/60 pointer-events-none" />
      </div>

      {/* ── Right: Config / detail panel ─────────────────────────────────── */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Rack header — always visible */}
        <div className="border-b border-border px-4 py-2.5 shrink-0 flex items-center gap-2">
          <Box className="size-4 text-muted-foreground shrink-0" />
          <h3 className="text-sm font-semibold truncate">{rack.name}</h3>
          <Badge
            variant={rack.role === 'base' ? 'default' : 'secondary'}
            className="ml-auto shrink-0 text-[10px]"
          >
            {rack.role === 'base' ? 'Base (NET)' : 'Compute'}
          </Badge>
        </div>

        {/* Scrollable content area */}
        <div className="flex-1 overflow-y-auto p-4">
          {selectedDevice ? (
            <DeviceDetail device={selectedDevice} onBack={() => setSelectedDevice(null)} />
          ) : (
            <RackConfig rack={{ ...rack, id: rackId } as any} serverModels={serverModels} />
          )}
        </div>
      </div>
    </div>
  );
}
