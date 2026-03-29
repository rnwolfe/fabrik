import { useMemo } from 'react';
import {
  Layers,
  Server,
  Zap,
  ArrowUpDown,
  Minus,
  Plus,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import DeviceModelPicker from '@/components/DeviceModelPicker';
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
  onSpineCountChange: (value: number) => void;
  onAssignSpine: (deviceModelId: number) => void;
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
  rackCount: number
): TopologyPlan | null {
  if (!leafModel || !spineModel) return null;

  const spineRadix = spineModel.port_count;
  const portGroupResult = deriveFromPortGroups(leafModel);

  let uplinks: number;
  let downlinks: number;
  let oversubscription: number;

  if (portGroupResult) {
    // Port groups define the physical port layout.
    // Spine count determines how many uplink ports are actually used.
    uplinks = Math.min(spineCount, portGroupResult.uplinks);
    downlinks = portGroupResult.downlinks;
    const downlinkBw = downlinks * portGroupResult.downlinkSpeed;
    const uplinkBw = uplinks * portGroupResult.uplinkSpeed;
    oversubscription = uplinkBw > 0 ? downlinkBw / uplinkBw : Infinity;
  } else {
    // Uniform ports — spine count directly sets the uplink/downlink split.
    uplinks = spineCount;
    downlinks = leafModel.port_count - uplinks;
    oversubscription = uplinks > 0 ? downlinks / uplinks : Infinity;
  }

  const maxLeaves = spineRadix;
  const leafCount = Math.min(rackCount || 1, maxLeaves);

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
  };
}

export default function BlockDetailPanel({
  block,
  aggs,
  racks,
  networkDevices,
  spineModelId,
  spineCount: spineCountProp,
  onSpineCountChange,
  onAssignSpine,
}: BlockDetailPanelProps) {
  const frontendAgg = aggs.find((a) => a.plane === 'front_end');
  const leafModel = frontendAgg
    ? networkDevices.find((d) => d.id === frontendAgg.device_model_id)
    : undefined;

  const rackCount = racks.length;
  const spineModel = spineModelId ? networkDevices.find((d) => d.id === spineModelId) : undefined;
  const portGroupResult = leafModel ? deriveFromPortGroups(leafModel) : null;
  const maxSpineCount = leafModel ? maxSpines(leafModel) : 0;

  // Default spine count: all available uplinks (non-blocking)
  const effectiveSpineCount = spineCountProp ?? maxSpineCount;

  const topology = useMemo(
    () => deriveTopology(leafModel, spineModel, effectiveSpineCount, rackCount),
    [leafModel, spineModel, effectiveSpineCount, rackCount]
  );

  const maxRacks = spineModel ? spineModel.port_count : 0;

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center gap-2">
        <Layers className="size-5 text-blue-500" />
        <h3 className="text-sm font-semibold">{block.name}</h3>
        <Badge variant="secondary" className="ml-auto text-[10px]">
          2-stage Clos
        </Badge>
      </div>

      {/* Spine model selector */}
      <div className="space-y-1.5">
        <Label className="text-xs">Spine Switch</Label>
        <DeviceModelPicker
          devices={networkDevices}
          value={spineModelId}
          onSelect={onAssignSpine}
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

      {/* Derived oversubscription */}
      {topology && (
        <div className="space-y-1.5">
          <Label className="text-xs flex items-center gap-1">
            <ArrowUpDown className="size-3" />
            Oversubscription
          </Label>
          <div className="rounded-md bg-muted/50 px-2.5 py-1.5">
            <p className="text-sm font-semibold">{topology.oversubscription.toFixed(1)}:1</p>
            <p className="text-[10px] text-muted-foreground">
              {portGroupResult
                ? `${portGroupResult.downlinks}×${portGroupResult.downlinkSpeed}G ↓ / ${effectiveSpineCount}×${portGroupResult.uplinkSpeed}G ↑`
                : `${topology.leaf_downlinks} downlinks / ${topology.leaf_uplinks} uplinks`}
            </p>
          </div>
        </div>
      )}

      {/* Topology stats */}
      {topology && (
        <div className="grid grid-cols-2 gap-2">
          <Stat icon={Server} label="Spines" value={topology.spine_count} />
          <Stat icon={Server} label="Leaves" value={`${rackCount}/${maxRacks}`} />
          <Stat icon={Zap} label="Host Ports" value={topology.total_host_ports} />
          <Stat icon={Layers} label="Switches" value={topology.total_switches} />
        </div>
      )}

      {/* Clos diagram */}
      {topology && (
        <div className="rounded-lg border border-border bg-muted/30 p-2">
          <ClosDiagram topology={topology} tier="frontend" />
        </div>
      )}

      {!leafModel && !spineModel && (
        <div className="rounded-lg border border-dashed border-border p-4 text-center">
          <p className="text-xs text-muted-foreground">
            Assign a spine switch model to see the topology diagram and capacity metrics.
          </p>
        </div>
      )}
    </div>
  );
}

function Stat({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Server;
  label: string;
  value: string | number;
}) {
  return (
    <div className="flex items-center gap-2 rounded-md bg-muted/50 px-2.5 py-1.5">
      <Icon className="size-3.5 text-muted-foreground" />
      <div>
        <p className="text-[10px] text-muted-foreground">{label}</p>
        <p className="text-sm font-semibold">{value}</p>
      </div>
    </div>
  );
}
