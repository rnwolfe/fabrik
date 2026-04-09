import { useQuery } from '@tanstack/react-query';
import { useState, useEffect } from 'react';
import {
  AlertTriangle,
  BarChart3,
  ChevronDown,
  ChevronUp,
  Network,
  RefreshCw,
  Zap,
} from 'lucide-react';
import { metricsApi } from '@/api/metrics';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { oversubColor, utilizationColor, powerUtilizationColor } from '@/features/metrics/colors';

const LS_KEY = 'fabrik:designMetricsStripCollapsed';

function getDefaultCollapsed(): boolean {
  const viewportDefault = typeof window !== 'undefined' && window.innerWidth < 1280;
  try {
    const stored = localStorage.getItem(LS_KEY);
    if (stored !== null) return stored === 'true';
  } catch {
    return viewportDefault;
  }
  return viewportDefault;
}

interface MetricPillProps {
  label: string;
  value: string;
  valueClass?: string;
  icon?: React.ReactNode;
  tooltip?: string;
}

function MetricPill({ label, value, valueClass, icon, tooltip }: MetricPillProps) {
  const pill = (
    <div className="flex items-center gap-2 rounded-md border border-border bg-card px-3 py-1.5">
      {icon && <span className="shrink-0 text-muted-foreground">{icon}</span>}
      <div className="flex flex-col leading-tight">
        <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {label}
        </span>
        <span className={`font-mono text-sm font-semibold ${valueClass ?? ''}`}>{value}</span>
      </div>
    </div>
  );

  if (!tooltip) return pill;

  return (
    <Tooltip>
      <TooltipTrigger>{pill}</TooltipTrigger>
      <TooltipContent>{tooltip}</TooltipContent>
    </Tooltip>
  );
}

function MetricPillSkeleton() {
  return (
    <div className="flex items-center gap-2 rounded-md border border-border bg-card px-3 py-1.5">
      <Skeleton className="size-4 rounded" />
      <div className="flex flex-col gap-1">
        <Skeleton className="h-2.5 w-16 rounded" />
        <Skeleton className="h-4 w-12 rounded" />
      </div>
    </div>
  );
}

interface MetricsStripProps {
  designId: number;
}

export default function MetricsStrip({ designId }: MetricsStripProps) {
  const [collapsed, setCollapsed] = useState(getDefaultCollapsed);

  useEffect(() => {
    try {
      localStorage.setItem(LS_KEY, String(collapsed));
    } catch {
      // ignore storage failures
    }
  }, [collapsed]);

  const {
    data: metrics,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['metrics', designId],
    queryFn: () => metricsApi.getDesignMetrics(designId),
    enabled: true,
    refetchInterval: 30_000,
  });

  return (
    <TooltipProvider>
      <div
        data-testid="metrics-strip"
        className="flex shrink-0 items-center gap-2 border-b border-border bg-muted/10 px-4 py-2"
      >
        {/* Toggle button */}
        <Button
          variant="ghost"
          size="sm"
          className="h-7 w-7 shrink-0 p-0"
          onClick={() => setCollapsed((c) => !c)}
          aria-label={collapsed ? 'Expand metrics strip' : 'Collapse metrics strip'}
          title={collapsed ? 'Show metrics' : 'Hide metrics'}
        >
          {collapsed ? (
            <ChevronDown className="size-3.5" />
          ) : (
            <ChevronUp className="size-3.5" />
          )}
        </Button>

        <span className="shrink-0 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/70">
          Metrics
        </span>

        {!collapsed && (
          <>
            {isLoading ? (
              <div className="flex gap-2 overflow-x-auto">
                {[1, 2, 3, 4, 5].map((i) => (
                  <MetricPillSkeleton key={i} />
                ))}
              </div>
            ) : isError ? (
              <div className="flex items-center gap-2 text-sm text-destructive">
                <AlertTriangle className="size-4 shrink-0" />
                <span>Failed to load metrics</span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 px-2 text-xs"
                  onClick={() => refetch()}
                >
                  <RefreshCw className="mr-1 size-3" />
                  Retry
                </Button>
              </div>
            ) : !metrics || metrics.empty ? (
              <p className="text-xs text-muted-foreground">
                Add a block to see metrics
              </p>
            ) : (
              <div className="flex gap-2 overflow-x-auto">
                {/* Worst leaf→spine oversubscription */}
                {metrics.choke_point ? (
                  <MetricPill
                    label="Worst Oversub"
                    value={`${metrics.choke_point.ratio.toFixed(2)}:1`}
                    valueClass={oversubColor(metrics.choke_point.ratio)}
                    icon={<AlertTriangle className="size-3.5" />}
                    tooltip={`Choke point: ${metrics.choke_point.fabric_name}`}
                  />
                ) : (
                  (() => {
                    const worstRatio = metrics.fabrics.reduce(
                      (max, f) => Math.max(max, f.leaf_spine_oversubscription),
                      0
                    );
                    return (
                      <MetricPill
                        label="Worst Oversub"
                        value={worstRatio > 0 ? `${worstRatio.toFixed(2)}:1` : '—'}
                        valueClass={worstRatio > 0 ? oversubColor(worstRatio) : undefined}
                        icon={<Network className="size-3.5" />}
                      />
                    );
                  })()
                )}

                {/* Total host ports */}
                <MetricPill
                  label="Host Ports"
                  value={metrics.total_hosts.toLocaleString()}
                  icon={<Network className="size-3.5" />}
                />

                {/* Bisection BW */}
                <MetricPill
                  label="Bisection BW"
                  value={`${metrics.bisection_bandwidth_gbps.toFixed(1)} Gbps`}
                  icon={<BarChart3 className="size-3.5" />}
                />

                {/* Power budget % */}
                <MetricPill
                  label="Power"
                  value={`${metrics.power.utilization_pct.toFixed(1)}%`}
                  valueClass={powerUtilizationColor(metrics.power.utilization_pct)}
                  icon={<Zap className="size-3.5" />}
                  tooltip={`${(metrics.power.total_draw_w / 1000).toFixed(1)} kW / ${(metrics.power.total_capacity_w / 1000).toFixed(1)} kW`}
                />

                {/* Port utilization % (average across all fabrics) */}
                {metrics.port_utilization && metrics.port_utilization.length > 0 && (() => {
                  const totalPorts = metrics.port_utilization.reduce((s, p) => s + p.total_ports, 0);
                  const allocPorts = metrics.port_utilization.reduce((s, p) => s + p.allocated_ports, 0);
                  const utilPct = totalPorts > 0 ? (allocPorts / totalPorts) * 100 : 0;
                  return (
                    <MetricPill
                      label="Port Util"
                      value={`${utilPct.toFixed(1)}%`}
                      valueClass={utilizationColor(utilPct)}
                      icon={<Network className="size-3.5" />}
                      tooltip={`${allocPorts.toLocaleString()} / ${totalPorts.toLocaleString()} ports`}
                    />
                  );
                })()}
              </div>
            )}
          </>
        )}
      </div>
    </TooltipProvider>
  );
}
