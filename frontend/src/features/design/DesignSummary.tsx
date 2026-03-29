import { Server, Layers, Zap, Box, Cpu, HardDrive, MemoryStick } from 'lucide-react';
import type { Block, RackSummary } from '@/models';

interface DesignSummaryProps {
  blocks: Block[];
  racksByBlock: Map<number, RackSummary[]>;
}

export default function DesignSummary({
  blocks,
  racksByBlock,
}: DesignSummaryProps) {
  let totalRacks = 0;
  let totalDevices = 0;
  let totalPowerW = 0;
  let totalU = 0;
  let usedU = 0;
  let totalVCPU = 0;
  let totalRAMGB = 0;
  let totalGPU = 0;

  for (const [, racks] of racksByBlock) {
    totalRacks += racks.length;
    for (const rack of racks) {
      totalU += rack.height_u;
      usedU += rack.used_u;
      totalPowerW += rack.used_watts_typical;
      if (rack.devices) {
        totalDevices += rack.devices.length;
        for (const dev of rack.devices) {
          totalVCPU += (dev.cpu_sockets ?? 0) * (dev.cores_per_socket ?? 0) * 2; // 2 threads per core
          totalRAMGB += dev.ram_gb ?? 0;
          totalGPU += dev.gpu_count ?? 0;
        }
      }
    }
  }

  const stages = blocks.length <= 1 ? 2 : 3;

  return (
    <div className="flex flex-col gap-4 p-4">
      <h3 className="text-sm font-semibold">Design Summary</h3>

      <div className="grid grid-cols-2 gap-2">
        <SummaryCard icon={Layers} label="Blocks" value={blocks.length} />
        <SummaryCard icon={Box} label="Racks" value={totalRacks} />
        <SummaryCard icon={Server} label="Devices" value={totalDevices} />
        <SummaryCard
          icon={Layers}
          label="Clos Stage"
          value={`${stages}-stage`}
        />
        <SummaryCard
          icon={Zap}
          label="Power (typ)"
          value={totalPowerW > 1000 ? `${(totalPowerW / 1000).toFixed(1)} kW` : `${totalPowerW} W`}
        />
        <SummaryCard
          icon={Box}
          label="Rack Space"
          value={`${usedU}/${totalU} U`}
        />
      </div>

      {(totalVCPU > 0 || totalRAMGB > 0 || totalGPU > 0) && (
        <>
          <h4 className="text-xs font-medium text-muted-foreground mt-2">Compute Capacity</h4>
          <div className="grid grid-cols-2 gap-2">
            {totalVCPU > 0 && <SummaryCard icon={Cpu} label="vCPU" value={totalVCPU.toLocaleString()} />}
            {totalRAMGB > 0 && (
              <SummaryCard
                icon={MemoryStick}
                label="RAM"
                value={totalRAMGB >= 1024 ? `${(totalRAMGB / 1024).toFixed(1)} TB` : `${totalRAMGB} GB`}
              />
            )}
            {totalGPU > 0 && <SummaryCard icon={HardDrive} label="GPUs" value={totalGPU.toLocaleString()} />}
          </div>
        </>
      )}

      {blocks.length === 0 && (
        <div className="rounded-lg border border-dashed border-border p-6 text-center">
          <p className="text-xs text-muted-foreground">
            Create your first block to start designing your fabric topology.
          </p>
        </div>
      )}
    </div>
  );
}

function SummaryCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Server;
  label: string;
  value: string | number;
}) {
  return (
    <div className="flex items-center gap-2.5 rounded-md border border-border bg-card px-3 py-2">
      <Icon className="size-4 text-muted-foreground shrink-0" />
      <div>
        <p className="text-[10px] text-muted-foreground leading-none">{label}</p>
        <p className="text-sm font-semibold mt-0.5">{value}</p>
      </div>
    </div>
  );
}
