import { useMemo } from 'react';
import {
  Layers,
  Server,
  Zap,
  ArrowUpDown,
  Minus,
  Plus,
  Info,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import DeviceModelPicker from '@/components/DeviceModelPicker';
import { HelpLink } from '@/components/HelpLink';
import ClosDiagram from './ClosDiagram';
import type {
  Block,
  BlockAggregationSummary,
  RackSummary,
  DeviceModel,
  TopologyPlan,
} from '@/models';

interface BlockDetailPanelProps {
  block: Block;
  aggs: BlockAggregationSummary[];
  racks: RackSummary[];
  networkDevices: DeviceModel[];
  spineModelId?: number;
  spineCount: number | null;
  hostLinkSpeedGbps: number;
  onSpineCountChange: (value: number) => void;
  onHostLinkSpeedChange: (value: number) => void;
  onAssignSpine: (deviceModelId: number, initialSpineCount: number) => void;
  onAssignLeaf: (deviceModelId: number) => void;
}

/**
 * Derive uplink/downlink port assignment and oversubscription from port groups.
 * Heuristic: the port group with the highest speed is uplinks (to spines),
 * everything else is downlinks (to servers). When all groups have the same
 * speed, the largest group is downlinks.
 */
function deriveFromPortGroups(leafModel: DeviceModel): {
  downlinks: number;
  downlinkSpeed: number;
  uplinks: number;
  uplinkSpeed: number;
  oversubscription: number;
} | null {
  const groups = leafModel.port_groups;
  if (!groups || groups.length < 2) return null;

  // Sort by speed ascending; tie-break by count descending (larger group = downlink).
  const sorted = [...groups].sort((a, b) =>
    a.speed_gbps !== b.speed_gbps ? a.speed_gbps - b.speed_gbps : b.count - a.count
  );

  // Last group (highest speed) = uplinks, rest = downlinks.
  const uplinkGroup = sorted[sorted.length - 1];
  const downlinkGroups = sorted.slice(0, -1);

  const downlinks = downlinkGroups.reduce((sum, g) => sum + g.count, 0);
  const downlinkBw = downlinkGroups.reduce((sum, g) => sum + g.count * g.speed_gbps, 0);
  const uplinkBw = uplinkGroup.count * uplinkGroup.speed_gbps;

  if (uplinkBw === 0) return null;

  return {
    downlinks,
    downlinkSpeed: downlinkGroups[0].speed_gbps,
    uplinks: uplinkGroup.count,
    uplinkSpeed: uplinkGroup.speed_gbps,
    oversubscription: downlinkBw / uplinkBw,
  };
}

/**
 * Compute the max spine count allowed for this leaf model.
 * With port groups: max = uplink port count.
 * Without port groups: max = port_count - 1 (at least 1 downlink).
 */
function maxSpines(leafModel: DeviceModel): number {
  const pg = deriveFromPortGroups(leafModel);
  return pg ? pg.uplinks : Math.max(1, leafModel.port_count - 1);
}

function deriveTopology(
  leafModel: DeviceModel | undefined,
  spineModel: DeviceModel | undefined,
  spineCount: number,
  rackCount: number,
  hostLinkSpeedGbps: number
): TopologyPlan | null {
  if (!leafModel || !spineModel) return null;

  const spineRadix = spineModel.port_count;
  const portGroupResult = deriveFromPortGroups(leafModel);

  let uplinks: number;
  let downlinks: number;
  let oversubscription: number;
  let bandwidthOversubscription: number | undefined;
  let uplinkSpeedGbps: number | undefined;
  let downlinkSpeedGbps: number | undefined;

  if (portGroupResult) {
    // Port groups define the physical port layout.
    // Spine count determines how many uplink ports are actually used.
    uplinks = Math.min(spineCount, portGroupResult.uplinks);
    downlinks = portGroupResult.downlinks;

    const effectiveDownlinkSpeed = hostLinkSpeedGbps > 0 ? hostLinkSpeedGbps : portGroupResult.downlinkSpeed;
    const downlinkBw = downlinks * effectiveDownlinkSpeed;
    const uplinkBw = uplinks * portGroupResult.uplinkSpeed;
    oversubscription = uplinkBw > 0 ? downlinkBw / uplinkBw : Infinity;
    bandwidthOversubscription = oversubscription;
    uplinkSpeedGbps = portGroupResult.uplinkSpeed;
    downlinkSpeedGbps = effectiveDownlinkSpeed;
  } else {
    // Uniform ports — spine count directly sets the uplink/downlink split.
    uplinks = spineCount;
    downlinks = leafModel.port_count - uplinks;
    oversubscription = uplinks > 0 ? downlinks / uplinks : Infinity;
  }

  const leavesPerRack = 2;
  const maxLeaves = spineRadix;
  const leafCount = Math.min((rackCount || 1) * leavesPerRack, maxLeaves);

  return {
    stages: 2,
    radix: leafModel.port_count,
    spine_radix: spineRadix,
    oversubscription: Math.round(oversubscription * 10) / 10,
    leaf_count: leafCount,
    spine_count: spineCount,
    leaf_downlinks: downlinks,
    leaf_uplinks: uplinks,
    total_switches: leafCount + spineCount,
    total_host_ports: leafCount * downlinks,
    bandwidth_oversubscription: bandwidthOversubscription,
    host_link_speed_gbps: hostLinkSpeedGbps || undefined,
    uplink_speed_gbps: uplinkSpeedGbps,
    downlink_speed_gbps: downlinkSpeedGbps,
    bisection_bandwidth_gbps: uplinkSpeedGbps ? spineCount * uplinks * uplinkSpeedGbps : undefined,
  };
}

export default function BlockDetailPanel({
  block,
  aggs,
  racks,
  networkDevices,
  spineModelId,
  spineCount: spineCountProp,
  hostLinkSpeedGbps,
  onSpineCountChange,
  onHostLinkSpeedChange,
  onAssignSpine,
  onAssignLeaf,
}: BlockDetailPanelProps) {
  const frontendAgg = aggs.find((a) => a.plane === 'front_end');
  const leafModelId = frontendAgg?.device_model_id;
  const leafModel = leafModelId
    ? networkDevices.find((d) => d.id === leafModelId)
    : undefined;

  const rackCount = racks.length;
  const baseRackCount = racks.filter((r) => r.role === 'base').length;
  const computeRackCount = racks.filter((r) => r.role === 'compute').length;
  const spineModel = spineModelId ? networkDevices.find((d) => d.id === spineModelId) : undefined;
  const portGroupResult = leafModel ? deriveFromPortGroups(leafModel) : null;
  const maxSpineCount = leafModel ? maxSpines(leafModel) : 0;

  // Default to 2 spines — a minimal HA pair. User adjusts from there.
  const effectiveSpineCount = spineCountProp ?? Math.min(2, maxSpineCount);

  const topology = useMemo(
    () => deriveTopology(leafModel, spineModel, effectiveSpineCount, rackCount, hostLinkSpeedGbps),
    [leafModel, spineModel, effectiveSpineCount, rackCount, hostLinkSpeedGbps]
  );

  // Count placed servers in compute racks
  const serverCount = racks
    .filter((r) => r.role === 'compute')
    .flatMap((r) => r.devices ?? [])
    .filter((d) => d.role === 'server')
    .length;

  const maxRacks = spineModel ? Math.floor(spineModel.port_count / 2) : 0;

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center gap-2">
        <Layers className="size-5 text-blue-500" />
        <h3 className="text-sm font-semibold">{block.name}</h3>
        <Badge variant="secondary" className="ml-auto text-[10px]">
          2-stage Clos
        </Badge>
      </div>

      {/* Leaf model selector */}
      <div className="space-y-1.5">
        <Label className="text-xs">Leaf Switch</Label>
        <DeviceModelPicker
          devices={networkDevices}
          value={leafModelId}
          onSelect={(id) => onAssignLeaf(id)}
          placeholder="Select leaf model…"
          triggerClassName="h-8 text-xs"
          role="leaf"
        />
      </div>

      {/* Spine model selector */}
      <div className="space-y-1.5">
        <Label className="text-xs">Spine Switch</Label>
        <DeviceModelPicker
          devices={networkDevices}
          value={spineModelId}
          onSelect={(id) => onAssignSpine(id, effectiveSpineCount)}
          placeholder="Select spine model…"
          triggerClassName="h-8 text-xs"
        />
      </div>

      {/* Spine count */}
      {leafModel && (
        <div className="space-y-1.5">
          <Label className="text-xs flex items-center gap-1">
            <Server className="size-3" />
            Spine Count
            <HelpLink article="radix" anchor="radix-in-clos-fabrics" />
          </Label>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              className="size-7"
              disabled={effectiveSpineCount <= 1}
              onClick={() => onSpineCountChange(effectiveSpineCount - 1)}
            >
              <Minus className="size-3" />
            </Button>
            <span className="w-8 text-center font-mono text-sm font-semibold">
              {effectiveSpineCount}
            </span>
            <Button
              variant="outline"
              size="icon"
              className="size-7"
              disabled={effectiveSpineCount >= maxSpineCount}
              onClick={() => onSpineCountChange(effectiveSpineCount + 1)}
            >
              <Plus className="size-3" />
            </Button>
            <span className="text-[10px] text-muted-foreground">
              / {maxSpineCount} max
            </span>
          </div>
        </div>
      )}

      {/* Server link speed */}
      {leafModel && portGroupResult && (
        <div className="space-y-1.5">
          <Label className="text-xs flex items-center gap-1">
            <Zap className="size-3" />
            Server Link Speed
          </Label>
          <select
            className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-xs"
            value={hostLinkSpeedGbps}
            onChange={(e) => onHostLinkSpeedChange(Number(e.target.value))}
          >
            <option value={0}>Auto ({portGroupResult.downlinkSpeed}G from port groups)</option>
            <option value={10}>10 Gbps</option>
            <option value={25}>25 Gbps</option>
            <option value={50}>50 Gbps</option>
            <option value={100}>100 Gbps</option>
            <option value={200}>200 Gbps</option>
            <option value={400}>400 Gbps</option>
          </select>
        </div>
      )}

      {/* Derived oversubscription */}
      {topology && (
        <div className="space-y-1.5">
          <Label className="text-xs flex items-center gap-1">
            <ArrowUpDown className="size-3" />
            Oversubscription
            <HelpLink article="oversubscription" anchor="leaf-spine-ratio" />
          </Label>
          <div className="rounded-md bg-muted/50 px-2.5 py-1.5">
            <p className="text-sm font-semibold">{topology.oversubscription.toFixed(1)}:1</p>
            <p className="text-[10px] text-muted-foreground">
              {portGroupResult && topology.downlink_speed_gbps
                ? `${topology.leaf_downlinks}×${topology.downlink_speed_gbps}G ↓ / ${effectiveSpineCount}×${topology.uplink_speed_gbps}G ↑`
                : portGroupResult
                  ? `${portGroupResult.downlinks}×${portGroupResult.downlinkSpeed}G ↓ / ${effectiveSpineCount}×${portGroupResult.uplinkSpeed}G ↑`
                  : `${topology.leaf_downlinks} downlinks / ${topology.leaf_uplinks} uplinks`}
            </p>
          </div>
        </div>
      )}

      {/* Topology stats */}
      {topology && (
        <div className="grid grid-cols-2 gap-2">
          <Stat
            icon={Layers}
            label="Racks"
            value={baseRackCount > 0 || computeRackCount > 0
              ? `${baseRackCount}B + ${computeRackCount}C${maxRacks ? ` / ${maxRacks}` : ''}`
              : `${rackCount}${maxRacks ? `/${maxRacks}` : ''}`}
            tooltip="B = base (network infra), C = compute (servers)"
          />
          <Stat icon={Server} label="Leaves" value={topology.leaf_count} />
          <Stat icon={Server} label="Spines" value={topology.spine_count} />
          <Stat
            icon={Zap}
            label="Host Ports"
            value={serverCount > 0
              ? `${serverCount} servers / ${topology.total_host_ports} ports`
              : topology.total_host_ports}
            tooltip="Physical leaf downlink port ceiling (1:1 assumed). When servers are placed, shows server count vs. available ports."
          />
        </div>
      )}

      {/* Clos diagram */}
      {topology && (
        <div className="rounded-lg border border-border bg-muted/30 p-2">
          <ClosDiagram topology={topology} tier="frontend" />
        </div>
      )}

      {!leafModel && (
        <div className="rounded-lg border border-dashed border-border p-4 text-center">
          <p className="text-xs text-muted-foreground">
            Assign a leaf switch model to enable topology metrics and the capacity diagram.
          </p>
        </div>
      )}

      {leafModel && !spineModel && (
        <div className="grid grid-cols-2 gap-2">
          <Stat icon={Layers} label="Racks" value="—" />
          <Stat icon={Server} label="Leaves" value="—" />
          <Stat icon={Server} label="Spines" value="—" />
          <Stat icon={Zap} label="Host Ports" value="—" />
        </div>
      )}
    </div>
  );
}

function Stat({
  icon: Icon,
  label,
  value,
  tooltip,
}: {
  icon: typeof Server;
  label: string;
  value: string | number;
  tooltip?: string;
}) {
  return (
    <div className="flex items-center gap-2 rounded-md bg-muted/50 px-2.5 py-1.5">
      <Icon className="size-3.5 text-muted-foreground" />
      <div className="flex-1">
        <p className="text-[10px] text-muted-foreground flex items-center gap-1">
          {label}
          {tooltip && (
            <Tooltip>
              <TooltipTrigger>
                <Info className="size-2.5 cursor-help" />
              </TooltipTrigger>
              <TooltipContent className="max-w-[220px] text-xs">{tooltip}</TooltipContent>
            </Tooltip>
          )}
        </p>
        <p className="text-sm font-semibold">{value}</p>
      </div>
    </div>
  );
}
