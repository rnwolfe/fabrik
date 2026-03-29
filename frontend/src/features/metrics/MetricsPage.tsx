import { useQuery } from '@tanstack/react-query';
import { metricsApi } from '@/api/metrics';
import { designsApi } from '@/api/designs';
import { useDesign } from '@/contexts/DesignContext';
import { PageHeader } from '@/components/PageHeader';
import { EmptyState } from '@/components/EmptyState';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { ProgressTrack, ProgressIndicator } from '@/components/ui/progress';
import {
  BarChart3,
  AlertTriangle,
  Zap,
  Cpu,
  HardDrive,
  Server,
  Network,
  ChevronDown,
} from 'lucide-react';
import type { FabricTier } from '@/models';

function oversubColor(ratio: number) {
  if (ratio <= 2) return 'text-green-600 dark:text-green-400';
  if (ratio <= 4) return 'text-amber-600 dark:text-amber-400';
  return 'text-red-600 dark:text-red-400';
}

function oversubBgColor(ratio: number) {
  if (ratio <= 2) return 'bg-green-500';
  if (ratio <= 4) return 'bg-amber-500';
  return 'bg-red-500';
}

export default function MetricsPage() {
  const { activeDesignId, setActiveDesignId } = useDesign();

  const { data: designs } = useQuery({
    queryKey: ['designs'],
    queryFn: designsApi.list,
  });

  const { data: metrics, isLoading } = useQuery({
    queryKey: ['metrics', activeDesignId],
    queryFn: () => metricsApi.getDesignMetrics(activeDesignId!),
    enabled: !!activeDesignId,
    refetchInterval: 30_000,
  });

  const tierLabel: Record<FabricTier, string> = {
    frontend: 'Frontend',
    backend: 'Backend',
  };

  return (
    <div className="mx-auto max-w-5xl">
      <PageHeader
        title="Metrics"
        subtitle="Design performance and capacity analytics"
        actions={
          designs && designs.length > 0 ? (
            <Select
              value={activeDesignId ? String(activeDesignId) : ''}
              onValueChange={(v) => setActiveDesignId((v && v !== '') ? Number(v) : null)}
            >
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="Select design…">
                  {(value: string) => designs?.find((d) => String(d.id) === value)?.name ?? 'Select design…'}
                </SelectValue>
                <ChevronDown className="size-4" />
              </SelectTrigger>
              <SelectContent>
                {designs.map((d) => (
                  <SelectItem key={d.id} value={String(d.id)}>{d.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : null
        }
      />

      {!activeDesignId ? (
        <EmptyState
          icon={BarChart3}
          title="No design selected"
          description="Select a design to view its topology metrics and capacity data."
        />
      ) : isLoading ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-24 animate-pulse rounded-xl bg-muted" />
            ))}
          </div>
          <div className="h-48 animate-pulse rounded-xl bg-muted" />
        </div>
      ) : !metrics || metrics.empty ? (
        <EmptyState
          icon={BarChart3}
          title="No data yet"
          description="Add fabrics to this design to see topology metrics."
        />
      ) : (
        <div className="space-y-6">
          {/* Summary cards */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <SummaryCard
              icon={Server}
              label="Total Hosts"
              value={metrics.total_hosts.toLocaleString()}
            />
            <SummaryCard
              icon={Network}
              label="Total Switches"
              value={metrics.total_switches.toLocaleString()}
            />
            <SummaryCard
              icon={BarChart3}
              label="Bisection BW"
              value={`${metrics.bisection_bandwidth_gbps.toFixed(1)} Gbps`}
            />
            <SummaryCard
              icon={Zap}
              label="Power Utilization"
              value={`${metrics.power.utilization_pct.toFixed(1)}%`}
              valueClass={
                metrics.power.utilization_pct > 90
                  ? 'text-red-600 dark:text-red-400'
                  : metrics.power.utilization_pct > 75
                  ? 'text-amber-600 dark:text-amber-400'
                  : undefined
              }
            />
          </div>

          {/* Choke point warning */}
          {metrics.choke_point && (
            <div className="flex items-start gap-3 rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-950/30">
              <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-400" />
              <div>
                <p className="text-sm font-medium text-amber-700 dark:text-amber-300">
                  Choke Point Detected
                </p>
                <p className="mt-0.5 text-sm text-amber-600 dark:text-amber-400">
                  <strong>{metrics.choke_point.fabric_name}</strong> (
                  {tierLabel[metrics.choke_point.tier]}) has the worst oversubscription at{' '}
                  <strong>{metrics.choke_point.ratio.toFixed(2)}:1</strong>
                </p>
              </div>
            </div>
          )}

          {/* Fabric metrics */}
          <section>
            <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Fabric Metrics
            </h2>
            <div className="rounded-xl border border-border bg-card overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Fabric</TableHead>
                    <TableHead>Tier</TableHead>
                    <TableHead className="text-right">Stages</TableHead>
                    <TableHead className="text-right">Leaf→Spine Oversub</TableHead>
                    <TableHead className="text-right">Spine→SS Oversub</TableHead>
                    <TableHead className="text-right">Switches</TableHead>
                    <TableHead className="text-right">Host Ports</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {metrics.fabrics.map((f) => (
                    <TableRow key={f.fabric_id}>
                      <TableCell className="font-medium">{f.fabric_name}</TableCell>
                      <TableCell>
                        <span className={`rounded-full px-2 py-0.5 text-[10px] font-medium ${
                          f.tier === 'frontend'
                            ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400'
                            : 'bg-purple-500/10 text-purple-600 dark:text-purple-400'
                        }`}>
                          {tierLabel[f.tier]}
                        </span>
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{f.stages}</TableCell>
                      <TableCell className="text-right">
                        <span className={`font-mono text-sm ${oversubColor(f.leaf_spine_oversubscription)}`}>
                          {f.leaf_spine_oversubscription.toFixed(2)}:1
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        {f.spine_super_spine_oversubscription !== undefined ? (
                          <span className={`font-mono text-sm ${oversubColor(f.spine_super_spine_oversubscription)}`}>
                            {f.spine_super_spine_oversubscription.toFixed(2)}:1
                          </span>
                        ) : (
                          <span className="text-muted-foreground">—</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm">{f.total_switches}</TableCell>
                      <TableCell className="text-right font-mono text-sm">{f.total_host_ports.toLocaleString()}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            <div className="mt-2 flex gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <span className="size-2 rounded-full bg-green-500 inline-block" /> ≤2:1 (excellent)
              </span>
              <span className="flex items-center gap-1">
                <span className="size-2 rounded-full bg-amber-500 inline-block" /> ≤4:1 (acceptable)
              </span>
              <span className="flex items-center gap-1">
                <span className="size-2 rounded-full bg-red-500 inline-block" /> &gt;4:1 (high)
              </span>
            </div>
          </section>

          {/* Power */}
          <section>
            <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Power
            </h2>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Zap className="size-4 text-muted-foreground" />
                  Power Budget
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-4 mb-3">
                  <div>
                    <p className="text-xs text-muted-foreground">Draw</p>
                    <p className="font-mono font-medium">{(metrics.power.total_draw_w / 1000).toFixed(1)} kW</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Capacity</p>
                    <p className="font-mono font-medium">{(metrics.power.total_capacity_w / 1000).toFixed(1)} kW</p>
                  </div>
                  <div className="ml-auto">
                    <p className="text-xs text-muted-foreground">Utilization</p>
                    <p className={`font-mono font-medium ${
                      metrics.power.utilization_pct > 90 ? 'text-red-600 dark:text-red-400' :
                      metrics.power.utilization_pct > 75 ? 'text-amber-600 dark:text-amber-400' : ''
                    }`}>
                      {metrics.power.utilization_pct.toFixed(1)}%
                    </p>
                  </div>
                </div>
                <ProgressTrack className="h-2">
                  <ProgressIndicator
                    className={
                      metrics.power.utilization_pct > 90 ? 'bg-red-500' :
                      metrics.power.utilization_pct > 75 ? 'bg-amber-500' : 'bg-green-500'
                    }
                    style={{ width: `${Math.min(metrics.power.utilization_pct, 100)}%` }}
                  />
                </ProgressTrack>
              </CardContent>
            </Card>
          </section>

          {/* Resource Capacity */}
          <section>
            <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-muted-foreground">
              Resource Capacity
            </h2>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <ResourceCard icon={Cpu} label="vCPU" value={metrics.capacity.total_vcpu.toLocaleString()} />
              <ResourceCard icon={HardDrive} label="RAM" value={`${metrics.capacity.total_ram_gb.toLocaleString()} GB`} />
              <ResourceCard icon={HardDrive} label="Storage" value={`${metrics.capacity.total_storage_tb.toFixed(1)} TB`} />
              <ResourceCard icon={Server} label="GPUs" value={metrics.capacity.total_gpu_count.toLocaleString()} />
            </div>
          </section>

          {/* Port Utilization */}
          {metrics.port_utilization && metrics.port_utilization.length > 0 && (
            <section>
              <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-muted-foreground">
                Port Utilization
              </h2>
              <div className="rounded-xl border border-border bg-card overflow-hidden">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Fabric</TableHead>
                      <TableHead>Tier</TableHead>
                      <TableHead className="text-right">Allocated</TableHead>
                      <TableHead className="text-right">Available</TableHead>
                      <TableHead className="text-right">Total</TableHead>
                      <TableHead className="w-[120px]">Utilization</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {metrics.port_utilization.map((p) => {
                      const utilPct = p.total_ports > 0 ? (p.allocated_ports / p.total_ports) * 100 : 0;
                      return (
                        <TableRow key={`${p.fabric_id}-${p.tier_name}`}>
                          <TableCell className="font-medium">{p.fabric_name}</TableCell>
                          <TableCell className="text-muted-foreground text-sm">{p.tier_name}</TableCell>
                          <TableCell className="text-right font-mono text-sm">{p.allocated_ports.toLocaleString()}</TableCell>
                          <TableCell className="text-right font-mono text-sm">{p.available_ports.toLocaleString()}</TableCell>
                          <TableCell className="text-right font-mono text-sm">{p.total_ports.toLocaleString()}</TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <ProgressTrack className="h-1.5 flex-1">
                                <ProgressIndicator
                                  className={oversubBgColor(utilPct > 80 ? 5 : utilPct > 60 ? 3 : 1)}
                                  style={{ width: `${utilPct}%` }}
                                />
                              </ProgressTrack>
                              <span className="w-10 text-right text-xs text-muted-foreground">
                                {utilPct.toFixed(0)}%
                              </span>
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}

function SummaryCard({
  icon: Icon,
  label,
  value,
  valueClass,
}: {
  icon: typeof Server;
  label: string;
  value: string;
  valueClass?: string;
}) {
  return (
    <Card size="sm">
      <CardContent className="pt-3">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Icon className="size-3.5" />
          <span className="text-xs">{label}</span>
        </div>
        <p className={`mt-1 font-mono text-lg font-semibold ${valueClass ?? ''}`}>{value}</p>
      </CardContent>
    </Card>
  );
}

function ResourceCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Server;
  label: string;
  value: string;
}) {
  return (
    <Card size="sm">
      <CardContent className="pt-3">
        <div className="flex items-center gap-1.5 text-muted-foreground">
          <Icon className="size-3.5" />
          <span className="text-xs">{label}</span>
        </div>
        <p className="mt-1 font-mono text-base font-semibold">{value}</p>
      </CardContent>
    </Card>
  );
}
