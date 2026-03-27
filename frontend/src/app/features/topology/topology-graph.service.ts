import { Injectable } from '@angular/core';
import type { ElementDefinition } from 'cytoscape';
import { FabricResponse, TopologyPlan } from '../../models/fabric';
import { DeviceModel } from '../../models/device-model';

export type NodeRole = 'host' | 'leaf' | 'spine' | 'super-spine' | 'agg1' | 'agg2';
export type UtilLevel = 'healthy' | 'warning' | 'critical';

export interface NodeData {
  id: string;
  label: string;
  role: NodeRole;
  tier: 'frontend' | 'backend';
  /** collapsed group node */
  isGroup?: boolean;
  /** index within the layer */
  index: number;
  count?: number;
  deviceModel?: DeviceModel;
  portCount?: number;
  powerWattsTypical?: number;
  utilization?: number;
  utilLevel?: UtilLevel;
  /** parent group id for compound graph */
  parent?: string;
  fabricId: number;
  fabricName: string;
}

export interface EdgeData {
  id: string;
  source: string;
  target: string;
  speed?: string;
  srcPort?: string;
  dstPort?: string;
  fabricId: number;
}

export interface TopoElements {
  nodes: ElementDefinition[];
  edges: ElementDefinition[];
}

function utilLevel(util: number | undefined): UtilLevel {
  if (util === undefined) return 'healthy';
  if (util >= 0.9) return 'critical';
  if (util >= 0.7) return 'warning';
  return 'healthy';
}

function mockUtil(role: NodeRole): number {
  // Produce varied mock utilization values seeded by role for demo purposes
  const seeds: Record<NodeRole, number> = {
    host: 0.45,
    leaf: 0.62,
    spine: 0.55,
    'super-spine': 0.38,
    agg1: 0.71,
    agg2: 0.83,
  };
  return seeds[role];
}

/**
 * Transform a FabricResponse into Cytoscape node/edge element definitions.
 * When collapsed=true each tier is represented by a single summary node.
 * When collapsed=false individual switch nodes are generated (up to the counts
 * in the TopologyPlan).
 */
export function fabricToElements(
  fabric: FabricResponse,
  collapsed: boolean,
  maxNodes = 500,
): TopoElements {
  const plan: TopologyPlan = fabric.topology;
  const tier = fabric.tier;
  const fid = fabric.id;

  const nodes: ElementDefinition[] = [];
  const edges: ElementDefinition[] = [];

  // Layer definitions in bottom-up order (hosts at bottom)
  interface LayerDef {
    role: NodeRole;
    count: number;
    model?: DeviceModel;
    portsPerSwitch?: number;
    powerPerSwitch?: number;
  }

  const layers: LayerDef[] = [
    {
      role: 'leaf',
      count: plan.leaf_count,
      model: fabric.leaf_model,
      portsPerSwitch: fabric.leaf_model?.port_count ?? plan.radix,
      powerPerSwitch: fabric.leaf_model?.power_watts_typical ?? undefined,
    },
    {
      role: 'spine',
      count: plan.spine_count,
      model: fabric.spine_model,
      portsPerSwitch: fabric.spine_model?.port_count ?? plan.radix,
      powerPerSwitch: fabric.spine_model?.power_watts_typical,
    },
  ];

  if (plan.super_spine_count) {
    layers.push({
      role: 'super-spine',
      count: plan.super_spine_count,
      model: fabric.super_spine_model,
      portsPerSwitch: fabric.super_spine_model?.port_count ?? plan.radix,
      powerPerSwitch: fabric.super_spine_model?.power_watts_typical,
    });
  }
  if (plan.agg1_count) {
    layers.push({ role: 'agg1', count: plan.agg1_count });
  }
  if (plan.agg2_count) {
    layers.push({ role: 'agg2', count: plan.agg2_count });
  }

  const totalSwitches = layers.reduce((s, l) => s + l.count, 0);

  if (collapsed || totalSwitches > maxNodes) {
    // One summary node per layer
    layers.forEach((layer, idx) => {
      const nodeId = `f${fid}-${layer.role}-group`;
      const util = mockUtil(layer.role);
      const data: NodeData = {
        id: nodeId,
        label: `${roleLabel(layer.role)}\n×${layer.count}`,
        role: layer.role,
        tier,
        isGroup: true,
        index: idx,
        count: layer.count,
        deviceModel: layer.model,
        portCount: layer.portsPerSwitch,
        powerWattsTypical: layer.powerPerSwitch,
        utilization: util,
        utilLevel: utilLevel(util),
        fabricId: fid,
        fabricName: fabric.name,
      };
      nodes.push({ data });
    });

    // Edges between adjacent layers
    for (let i = 0; i < layers.length - 1; i++) {
      const lower = layers[i];
      const upper = layers[i + 1];
      const eid = `f${fid}-edge-${lower.role}-${upper.role}`;
      const edgeData: EdgeData = {
        id: eid,
        source: `f${fid}-${lower.role}-group`,
        target: `f${fid}-${upper.role}-group`,
        fabricId: fid,
      };
      edges.push({ data: edgeData });
    }
  } else {
    // Individual nodes per switch
    const allNodeIds: Record<NodeRole, string[]> = {
      host: [],
      leaf: [],
      spine: [],
      'super-spine': [],
      agg1: [],
      agg2: [],
    };

    layers.forEach((layer) => {
      for (let i = 0; i < layer.count; i++) {
        const nodeId = `f${fid}-${layer.role}-${i}`;
        allNodeIds[layer.role].push(nodeId);
        const util = mockUtil(layer.role);
        const data: NodeData = {
          id: nodeId,
          label: `${roleLabel(layer.role)}-${i + 1}`,
          role: layer.role,
          tier,
          index: i,
          deviceModel: layer.model,
          portCount: layer.portsPerSwitch,
          powerWattsTypical: layer.powerPerSwitch,
          utilization: util,
          utilLevel: utilLevel(util),
          fabricId: fid,
          fabricName: fabric.name,
        };
        nodes.push({ data });
      }
    });

    // Full mesh edges between adjacent layers
    for (let li = 0; li < layers.length - 1; li++) {
      const lowerRole = layers[li].role;
      const upperRole = layers[li + 1].role;
      const lowerIds = allNodeIds[lowerRole];
      const upperIds = allNodeIds[upperRole];

      // Limit edges to avoid overwhelming the graph: each lower connects to each upper
      lowerIds.forEach((srcId, si) => {
        upperIds.forEach((dstId, di) => {
          const eid = `f${fid}-e-${lowerRole}${si}-${upperRole}${di}`;
          const edgeData: EdgeData = {
            id: eid,
            source: srcId,
            target: dstId,
            fabricId: fid,
          };
          edges.push({ data: edgeData });
        });
      });
    }
  }

  return { nodes, edges };
}

export function roleLabel(role: NodeRole): string {
  const labels: Record<NodeRole, string> = {
    host: 'Host',
    leaf: 'Leaf',
    spine: 'Spine',
    'super-spine': 'SuperSpine',
    agg1: 'Agg-1',
    agg2: 'Agg-2',
  };
  return labels[role];
}

export function roleSortKey(role: NodeRole): number {
  return ['host', 'leaf', 'spine', 'super-spine', 'agg1', 'agg2'].indexOf(role);
}

@Injectable({ providedIn: 'root' })
export class TopologyGraphService {
  /**
   * Convert an array of fabrics into Cytoscape elements.
   * Each fabric contributes a set of nodes and edges at a given position offset.
   */
  buildElements(
    fabrics: FabricResponse[],
    collapseState: Map<number, boolean>,
  ): TopoElements {
    const allNodes: ElementDefinition[] = [];
    const allEdges: ElementDefinition[] = [];

    fabrics.forEach(fabric => {
      const collapsed = collapseState.get(fabric.id) ?? true;
      const { nodes, edges } = fabricToElements(fabric, collapsed);
      allNodes.push(...nodes);
      allEdges.push(...edges);
    });

    return { nodes: allNodes, edges: allEdges };
  }

  /**
   * Count the total number of expanded nodes across all fabrics.
   */
  countExpandedNodes(fabrics: FabricResponse[], collapseState: Map<number, boolean>): number {
    return fabrics.reduce((sum, fabric) => {
      const collapsed = collapseState.get(fabric.id) ?? true;
      if (collapsed) {
        // One node per layer
        const plan = fabric.topology;
        let layers = 2;
        if (plan.super_spine_count) layers++;
        if (plan.agg1_count) layers++;
        if (plan.agg2_count) layers++;
        return sum + layers;
      }
      return sum + fabric.topology.total_switches;
    }, 0);
  }
}
