import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Server, Zap, Box, Minus, Plus } from 'lucide-react';
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

export default function RackElevation({ rackId }: RackElevationProps) {
  const queryClient = useQueryClient();
  const [serverModelId, setServerModelId] = useState<number | undefined>();
  const [serverCount, setServerCount] = useState(1);

  // Fetch the full RackSummary (includes devices, used_u, available_u).
  // allRacks comes from GET /api/racks which returns bare Rack (no summary fields),
  // so we cannot use it as initialData — always fetch the individual summary.
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

  const selectedServerModel = serverModels.find((d: DeviceModel) => d.id === serverModelId);

  const placeServersMutation = useMutation({
    mutationFn: ({ modelId, count }: { modelId: number; count: number }) =>
      racksApi.placeServerDevices(rackId, modelId, count),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['racks'] });
      queryClient.invalidateQueries({ queryKey: ['rack', rackId] });
    },
  });

  if (isLoading) {
    return (
      <div className="flex flex-col gap-4 p-4">
        <div className="h-6 w-32 animate-pulse rounded bg-muted" />
        <div className="grid grid-cols-3 gap-2">
          {[1, 2, 3].map((i) => <div key={i} className="h-12 animate-pulse rounded-md bg-muted" />)}
        </div>
        <div className="h-64 animate-pulse rounded-lg bg-muted" />
      </div>
    );
  }
  if (!rack) return null;

  // Build a map of position → device for rendering
  const devicesByPos = new Map<number, DeviceSummary>();
  for (const dev of rack.devices ?? []) {
    devicesByPos.set(dev.position, dev);
  }

  const heightU = rack.height_u;
  const powerPct = rack.power_capacity_w > 0
    ? Math.min(100, Math.round((rack.used_watts_typical / rack.power_capacity_w) * 100))
    : 0;

  // Max servers that fit given selected server model height
  const maxServers = selectedServerModel
    ? Math.floor((rack.available_u + (rack.devices ?? []).filter(d => d.role === 'server').reduce((s, d) => s + d.height_u, 0)) / selectedServerModel.height_u)
    : 0;

  return (
    <div className="flex flex-col gap-4 p-4">
      {/* Header */}
      <div className="flex items-center gap-2">
        <Box className="size-5 text-muted-foreground" />
        <h3 className="text-sm font-semibold">{rack.name}</h3>
        <Badge variant={rack.role === 'base' ? 'default' : 'secondary'} className="ml-auto text-[10px]">
          {rack.role === 'base' ? 'Base (NET)' : 'Compute'}
        </Badge>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-3 gap-2">
        <div className="rounded-md bg-muted/50 px-2.5 py-1.5 text-center">
          <p className="text-[10px] text-muted-foreground">Used</p>
          <p className="text-sm font-semibold font-mono">{rack.used_u}/{rack.height_u}U</p>
        </div>
        <div className="rounded-md bg-muted/50 px-2.5 py-1.5 text-center">
          <p className="text-[10px] text-muted-foreground">Free</p>
          <p className="text-sm font-semibold font-mono">{rack.available_u}U</p>
        </div>
        <div className={cn('rounded-md px-2.5 py-1.5 text-center', powerPct > 90 ? 'bg-red-500/10' : powerPct > 75 ? 'bg-amber-500/10' : 'bg-muted/50')}>
          <p className="text-[10px] text-muted-foreground flex items-center justify-center gap-0.5"><Zap className="size-2.5" />Power</p>
          <p className={cn('text-sm font-semibold font-mono', powerPct > 90 ? 'text-red-600' : powerPct > 75 ? 'text-amber-600' : '')}>{powerPct}%</p>
        </div>
      </div>

      {/* Rack elevation diagram */}
      <div className="rounded-lg border border-border bg-muted/10 overflow-hidden">
        {/* Top of rack label */}
        <div className="border-b border-border px-2 py-1 bg-muted/30">
          <p className="text-[9px] font-medium text-muted-foreground uppercase tracking-wider text-center">Top of Rack</p>
        </div>

        {/* U slots — rendered top to bottom (position 1 = top) */}
        <div className="flex flex-col">
          {Array.from({ length: heightU }, (_, i) => i + 1).map((u) => {
            const dev = devicesByPos.get(u);
            if (dev) {
              // Device occupies this slot — render it spanning its height
              const colorClass = ROLE_COLORS[dev.role] ?? ROLE_COLORS.other;
              return (
                <div
                  key={u}
                  className={cn('flex items-center px-2 gap-1.5 border-b border-border/40 font-mono text-[10px] font-medium', colorClass)}
                  style={{ height: `${Math.max(20, dev.height_u * 20)}px` }}
                >
                  <span className="text-[9px] opacity-60 w-4 text-right shrink-0">{u}</span>
                  <Server className="size-2.5 shrink-0 opacity-70" />
                  <span className="truncate">{dev.name}</span>
                  <span className="ml-auto opacity-60">{ROLE_LABELS[dev.role] ?? dev.role}</span>
                </div>
              );
            }
            // Check if this U is covered by a multi-U device starting above
            const coveredBy = Array.from(devicesByPos.entries()).find(
              ([pos, d]) => pos < u && pos + d.height_u > u
            );
            if (coveredBy) return null; // rendered as part of the device block above

            return (
              <div
                key={u}
                className="flex items-center px-2 gap-1.5 border-b border-border/40 h-5"
                style={{ height: '20px' }}
              >
                <span className="text-[9px] text-muted-foreground/40 font-mono w-4 text-right shrink-0">{u}</span>
                <div className="flex-1 border-b border-dashed border-border/30" />
              </div>
            );
          })}
        </div>

        {/* Bottom of rack label */}
        <div className="border-t border-border px-2 py-1 bg-muted/30">
          <p className="text-[9px] font-medium text-muted-foreground uppercase tracking-wider text-center">Bottom of Rack</p>
        </div>
      </div>

      {/* Server placement — only show for compute racks */}
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
              {placeServersMutation.isPending ? 'Placing…' : `Place ${serverCount} Server${serverCount !== 1 ? 's' : ''}`}
            </Button>
          )}

          {serverModels.length === 0 && (
            <p className="text-xs text-muted-foreground text-center py-2">
              No server models in catalog. Add a server model to place servers.
            </p>
          )}
        </div>
      )}
    </div>
  );
}
