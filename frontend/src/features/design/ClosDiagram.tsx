import { useState } from 'react';
import type { FabricResponse, TopologyPlan } from '@/models';

// ─── Layout constants ────────────────────────────────────────────────────────
const SVG_W = 1000;
const PAD_X = 100;   // left margin for tier labels
const PAD_Y = 48;    // top / bottom margin
const TIER_GAP = 120; // vertical distance between tier centre-lines
const NODE_W = 86;
const NODE_H = 32;
const MAX_VISIBLE = 8; // max nodes drawn per tier before collapsing

// ─── Colour palette (works on both light & dark canvas) ────────────────────
const TIER_COLORS: Record<string, { fill: string; stroke: string; text: string }> = {
  super_spine: { fill: '#6d28d9', stroke: '#7c3aed', text: '#ffffff' },
  spine:       { fill: '#1d4ed8', stroke: '#2563eb', text: '#ffffff' },
  leaf:        { fill: '#0f766e', stroke: '#0d9488', text: '#ffffff' },
  host:        { fill: 'transparent', stroke: '#64748b', text: '#94a3b8' },
};

// ─── Types ───────────────────────────────────────────────────────────────────
interface Tier {
  role: 'super_spine' | 'spine' | 'leaf' | 'host';
  label: string;
  count: number;
  sublabel?: string;
}

interface PlacedNode {
  x: number;
  y: number;
  label: string;
  sublabel?: string;
  isEllipsis?: boolean;
  role: Tier['role'];
}

interface PlacedTier {
  tier: Tier;
  cy: number;
  nodes: PlacedNode[];
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function buildTiers(topology: TopologyPlan, tier: string): Tier[] {
  const tiers: Tier[] = [];
  if (topology.super_spine_count) {
    tiers.push({ role: 'super_spine', label: 'Super-Spine', count: topology.super_spine_count });
  }
  if (topology.agg1_count) {
    tiers.push({ role: 'spine', label: 'Agg-1', count: topology.agg1_count });
  }
  if (topology.agg2_count) {
    tiers.push({ role: 'spine', label: 'Agg-2', count: topology.agg2_count });
  }
  tiers.push({ role: 'spine', label: 'Spine', count: topology.spine_count });
  tiers.push({
    role: 'leaf',
    label: 'Leaf',
    count: topology.leaf_count,
    sublabel: `${topology.leaf_downlinks} downlinks`,
  });
  tiers.push({
    role: 'host',
    label: 'Hosts',
    count: topology.total_host_ports,
    sublabel: `${topology.total_host_ports.toLocaleString()} ports`,
  });
  void tier;
  return tiers;
}

function placeNodes(tier: Tier, cy: number): PlacedNode[] {
  const { count, role, sublabel } = tier;
  const isHost = role === 'host';

  if (isHost) {
    // Hosts: single representative bar
    return [{
      x: SVG_W / 2 - NODE_W / 2,
      y: cy - NODE_H / 2,
      label: `${count.toLocaleString()} host ports`,
      role,
      isEllipsis: false,
    }];
  }

  const visible = count <= MAX_VISIBLE
    ? count
    : MAX_VISIBLE; // we'll show MAX_VISIBLE - 1 real + 1 ellipsis

  const actualVisible = count <= MAX_VISIBLE ? count : MAX_VISIBLE - 1;
  const gap = 12;
  const totalW = visible * NODE_W + (visible - 1) * gap;
  const drawableW = SVG_W - PAD_X - 40;
  const startX = PAD_X + (drawableW - totalW) / 2;

  const nodes: PlacedNode[] = [];
  for (let i = 0; i < actualVisible; i++) {
    nodes.push({
      x: startX + i * (NODE_W + gap),
      y: cy - NODE_H / 2,
      label: `${tier.label} ${i + 1}`,
      sublabel,
      role,
    });
  }

  if (count > MAX_VISIBLE) {
    const remaining = count - (MAX_VISIBLE - 1);
    nodes.push({
      x: startX + (MAX_VISIBLE - 1) * (NODE_W + gap),
      y: cy - NODE_H / 2,
      label: `+${remaining} more`,
      role,
      isEllipsis: true,
    });
  }

  return nodes;
}

function placedTiers(topology: TopologyPlan, fabricTier: string): PlacedTier[] {
  const tiers = buildTiers(topology, fabricTier);
  const startY = PAD_Y + NODE_H / 2;

  return tiers.map((tier, i) => {
    const cy = startY + i * TIER_GAP;
    return { tier, cy, nodes: placeNodes(tier, cy) };
  });
}

function svgHeight(tierCount: number) {
  return PAD_Y * 2 + NODE_H + (tierCount - 1) * TIER_GAP;
}

// ─── Connection drawing ───────────────────────────────────────────────────────

function ConnectionLines({
  upper,
  lower,
  oversubLabel,
}: {
  upper: PlacedTier;
  lower: PlacedTier;
  oversubLabel?: string;
}) {
  const un = upper.nodes;
  const ln = lower.nodes;

  if (lower.tier.role === 'host') {
    // Host tier: draw vertical drops from each leaf to a horizontal bus
    const busY = lower.nodes[0].y;
    const paths: string[] = un.map((n) => {
      const cx = n.x + NODE_W / 2;
      return `M ${cx} ${n.y + NODE_H} L ${cx} ${busY}`;
    });
    return (
      <g>
        {paths.map((d, i) => (
          <path key={i} d={d} stroke="#475569" strokeWidth="1" strokeOpacity="0.4" fill="none" strokeDasharray="3 3" />
        ))}
      </g>
    );
  }

  // Determine if we should draw individual lines or a band
  const totalLines = un.length * ln.length;
  const midY = (upper.cy + NODE_H / 2 + lower.cy - NODE_H / 2) / 2;

  if (totalLines <= 40) {
    // Individual lines
    const paths: string[] = [];
    for (const u of un) {
      for (const l of ln) {
        const x1 = u.x + NODE_W / 2;
        const y1 = u.y + NODE_H;
        const x2 = l.x + NODE_W / 2;
        const y2 = l.y;
        paths.push(`M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}`);
      }
    }
    return (
      <g>
        {paths.map((d, i) => (
          <path key={i} d={d} stroke="#94a3b8" strokeWidth="1" strokeOpacity="0.35" fill="none" />
        ))}
        {oversubLabel && (
          <text
            x={SVG_W / 2}
            y={midY}
            textAnchor="middle"
            fontSize="10"
            fill="#94a3b8"
            className="select-none"
          >
            {oversubLabel}
          </text>
        )}
      </g>
    );
  }

  // Connectivity band for dense fabrics
  const uLeft = un[0].x + NODE_W / 2;
  const uRight = un[un.length - 1].x + NODE_W / 2;
  const lLeft = ln[0].x + NODE_W / 2;
  const lRight = ln[ln.length - 1].x + NODE_W / 2;
  const y1 = upper.cy + NODE_H / 2;
  const y2 = lower.cy - NODE_H / 2;

  return (
    <g>
      <path
        d={`M ${uLeft} ${y1} L ${uRight} ${y1} L ${lRight} ${y2} L ${lLeft} ${y2} Z`}
        fill="#3b82f6"
        fillOpacity="0.07"
        stroke="#3b82f6"
        strokeOpacity="0.2"
        strokeWidth="0.5"
      />
      {oversubLabel && (
        <text
          x={SVG_W / 2}
          y={(y1 + y2) / 2 + 4}
          textAnchor="middle"
          fontSize="10"
          fill="#94a3b8"
          className="select-none"
        >
          {oversubLabel}
        </text>
      )}
    </g>
  );
}

// ─── Node ────────────────────────────────────────────────────────────────────

function Node({
  node,
  hovered,
  onEnter,
  onLeave,
}: {
  node: PlacedNode;
  hovered: boolean;
  onEnter: () => void;
  onLeave: () => void;
}) {
  const colors = TIER_COLORS[node.role];
  const isHost = node.role === 'host';
  const r = isHost ? 4 : 6;

  if (isHost) {
    // Wide bus bar
    return (
      <g onMouseEnter={onEnter} onMouseLeave={onLeave}>
        <rect
          x={PAD_X}
          y={node.y}
          width={SVG_W - PAD_X - 40}
          height={NODE_H}
          rx={r}
          fill="none"
          stroke={colors.stroke}
          strokeWidth="1.5"
          strokeDasharray="6 4"
          opacity="0.5"
        />
        <text
          x={SVG_W / 2}
          y={node.y + NODE_H / 2 + 4}
          textAnchor="middle"
          fontSize="11"
          fill={colors.text}
          className="select-none font-mono"
        >
          {node.label}
        </text>
      </g>
    );
  }

  return (
    <g
      onMouseEnter={onEnter}
      onMouseLeave={onLeave}
      style={{ cursor: 'default' }}
    >
      <rect
        x={node.x}
        y={node.y}
        width={NODE_W}
        height={NODE_H}
        rx={r}
        fill={node.isEllipsis ? 'none' : colors.fill}
        stroke={hovered ? '#ffffff' : colors.stroke}
        strokeWidth={hovered ? 2 : node.isEllipsis ? 1.5 : 0}
        strokeDasharray={node.isEllipsis ? '4 3' : undefined}
        opacity={node.isEllipsis ? 0.6 : hovered ? 1 : 0.9}
      />
      <text
        x={node.x + NODE_W / 2}
        y={node.y + NODE_H / 2 + 4}
        textAnchor="middle"
        fontSize={node.isEllipsis ? '10' : '11'}
        fill={node.isEllipsis ? '#94a3b8' : colors.text}
        fontWeight="500"
        className="select-none"
      >
        {node.isEllipsis ? node.label : node.label.replace(/^.*? /, '')}
      </text>
    </g>
  );
}

// ─── Tier label ──────────────────────────────────────────────────────────────

function TierLabel({ placed }: { placed: PlacedTier }) {
  const colors = TIER_COLORS[placed.tier.role];
  return (
    <g>
      <text
        x={PAD_X - 12}
        y={placed.cy + 4}
        textAnchor="end"
        fontSize="11"
        fontWeight="600"
        fill={placed.tier.role === 'host' ? '#64748b' : colors.fill}
        className="select-none uppercase tracking-wider"
        style={{ letterSpacing: '0.06em' }}
      >
        {placed.tier.label}
      </text>
      <text
        x={PAD_X - 12}
        y={placed.cy + 16}
        textAnchor="end"
        fontSize="10"
        fill="#64748b"
        className="select-none"
      >
        ×{placed.tier.count.toLocaleString()}
      </text>
    </g>
  );
}

// ─── Main diagram ─────────────────────────────────────────────────────────────

interface ClosDiagramProps {
  fabric?: FabricResponse;
  topology?: TopologyPlan;
  tier?: string;
}

export default function ClosDiagram({ fabric, topology: topologyProp, tier: tierProp }: ClosDiagramProps) {
  const [hoveredNode, setHoveredNode] = useState<string | null>(null);
  const topology = fabric?.topology ?? topologyProp;
  const tier = fabric?.tier ?? tierProp ?? 'frontend';
  const metrics = fabric?.metrics;
  const diagramName = fabric?.name ?? 'Block';
  if (!topology) return null;

  const placed = placedTiers(topology, tier);
  const height = svgHeight(placed.length);

  // Build oversubscription labels for inter-tier connections
  const oversubLabels: (string | undefined)[] = placed.map((_, i) => {
    if (i === placed.length - 2) {
      // spine→leaf (or last real tier before hosts)
      const ratio = metrics?.leaf_spine_oversubscription ?? topology.oversubscription;
      return ratio != null ? `${ratio.toFixed(1)}:1 oversubscription` : undefined;
    }
    if (i === placed.length - 3 && metrics?.spine_super_spine_oversubscription != null) {
      return `${metrics.spine_super_spine_oversubscription.toFixed(1)}:1 oversubscription`;
    }
    return undefined;
  });

  return (
    <svg
      viewBox={`0 0 ${SVG_W} ${height}`}
      className="w-full"
      preserveAspectRatio="xMidYMid meet"
      aria-label={`${diagramName} topology diagram`}
    >
      {/* Subtle tier band backgrounds */}
      {placed.map((p) =>
        p.tier.role !== 'host' ? (
          <rect
            key={`band-${p.tier.role}-${p.cy}`}
            x={PAD_X}
            y={p.cy - TIER_GAP / 2}
            width={SVG_W - PAD_X - 40}
            height={TIER_GAP}
            fill={TIER_COLORS[p.tier.role].fill}
            fillOpacity="0.03"
            rx="8"
          />
        ) : null
      )}

      {/* Connection lines between tiers */}
      {placed.slice(0, -1).map((upper, i) => (
        <ConnectionLines
          key={`conn-${i}`}
          upper={upper}
          lower={placed[i + 1]}
          oversubLabel={oversubLabels[i]}
        />
      ))}

      {/* Tier labels */}
      {placed.map((p) => (
        <TierLabel key={`label-${p.tier.label}-${p.cy}`} placed={p} />
      ))}

      {/* Nodes */}
      {placed.map((p) =>
        p.nodes.map((node, ni) => {
          const key = `${p.tier.label}-${ni}`;
          return (
            <Node
              key={key}
              node={node}
              hovered={hoveredNode === key}
              onEnter={() => setHoveredNode(key)}
              onLeave={() => setHoveredNode(null)}
            />
          );
        })
      )}
    </svg>
  );
}
