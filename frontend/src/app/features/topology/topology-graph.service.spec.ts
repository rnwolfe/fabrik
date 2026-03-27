import { describe, it, expect } from 'vitest';
import {
  fabricToElements,
  roleLabel,
  TopologyGraphService,
  NodeData,
  EdgeData,
} from './topology-graph.service';
import { FabricResponse } from '../../models/fabric';

const mockFabricFrontend: FabricResponse = {
  id: 1,
  design_id: 0,
  name: 'fe-fabric',
  tier: 'frontend',
  stages: 2,
  radix: 4,
  oversubscription: 1.0,
  description: '',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  topology: {
    stages: 2,
    radix: 4,
    oversubscription: 1.0,
    leaf_count: 2,
    spine_count: 2,
    leaf_uplinks: 2,
    leaf_downlinks: 2,
    total_switches: 4,
    total_host_ports: 4,
  },
  metrics: {
    total_switches: 4,
    total_host_ports: 4,
    oversubscription_ratio: 1.0,
  },
};

const mockFabricBackend: FabricResponse = {
  ...mockFabricFrontend,
  id: 2,
  name: 'be-fabric',
  tier: 'backend',
};

const mockFabric3Stage: FabricResponse = {
  ...mockFabricFrontend,
  id: 3,
  name: '3stage-fabric',
  topology: {
    ...mockFabricFrontend.topology,
    stages: 3,
    super_spine_count: 1,
    total_switches: 5,
  },
};

const mockFabric5Stage: FabricResponse = {
  ...mockFabricFrontend,
  id: 4,
  name: '5stage-fabric',
  topology: {
    ...mockFabricFrontend.topology,
    stages: 5,
    super_spine_count: 1,
    agg1_count: 2,
    agg2_count: 1,
    total_switches: 8,
  },
};

describe('fabricToElements', () => {
  it('collapsed mode produces one node per layer for 2-stage', () => {
    const { nodes, edges } = fabricToElements(mockFabricFrontend, true);
    expect(nodes.length).toBe(2);
    expect(edges.length).toBe(1);
  });

  it('collapsed mode produces 3 nodes for 3-stage', () => {
    const { nodes, edges } = fabricToElements(mockFabric3Stage, true);
    expect(nodes.length).toBe(3);
    expect(edges.length).toBe(2);
  });

  it('collapsed mode produces 5 nodes for 5-stage', () => {
    const { nodes } = fabricToElements(mockFabric5Stage, true);
    expect(nodes.length).toBe(5);
  });

  it('collapsed nodes have isGroup=true', () => {
    const { nodes } = fabricToElements(mockFabricFrontend, true);
    nodes.forEach(n => {
      expect((n.data as NodeData).isGroup).toBe(true);
    });
  });

  it('collapsed nodes carry correct tier', () => {
    const { nodes } = fabricToElements(mockFabricFrontend, true);
    nodes.forEach(n => expect((n.data as NodeData).tier).toBe('frontend'));

    const { nodes: beNodes } = fabricToElements(mockFabricBackend, true);
    beNodes.forEach(n => expect((n.data as NodeData).tier).toBe('backend'));
  });

  it('expanded mode produces individual nodes', () => {
    // 2 leaf + 2 spine = 4 nodes
    const { nodes } = fabricToElements(mockFabricFrontend, false);
    expect(nodes.length).toBe(4);
  });

  it('expanded mode edges connect adjacent layers', () => {
    // 2 leaf × 2 spine = 4 edges
    const { edges } = fabricToElements(mockFabricFrontend, false);
    expect(edges.length).toBe(4);
  });

  it('expanded nodes do not have isGroup set', () => {
    const { nodes } = fabricToElements(mockFabricFrontend, false);
    nodes.forEach(n => {
      expect((n.data as NodeData).isGroup).toBeFalsy();
    });
  });

  it('forces collapse when totalSwitches exceeds maxNodes', () => {
    // maxNodes=2, total switches=4 → should collapse
    const { nodes } = fabricToElements(mockFabricFrontend, false, 2);
    // In collapsed mode there should be 2 group nodes
    expect(nodes.every(n => (n.data as NodeData).isGroup)).toBe(true);
  });

  it('nodes have utilization and utilLevel', () => {
    const { nodes } = fabricToElements(mockFabricFrontend, true);
    nodes.forEach(n => {
      const d = n.data as NodeData;
      expect(d.utilization).toBeTypeOf('number');
      expect(['healthy', 'warning', 'critical']).toContain(d.utilLevel);
    });
  });

  it('edge data includes fabricId', () => {
    const { edges } = fabricToElements(mockFabricFrontend, true);
    edges.forEach(e => {
      expect((e.data as EdgeData).fabricId).toBe(1);
    });
  });

  it('collapsed node label includes count', () => {
    const { nodes } = fabricToElements(mockFabricFrontend, true);
    const leafNode = nodes.find(n => (n.data as NodeData).role === 'leaf');
    expect(leafNode).toBeDefined();
    expect((leafNode!.data as NodeData).label).toContain('2');
  });

  it('nodes include fabricName', () => {
    const { nodes } = fabricToElements(mockFabricFrontend, true);
    nodes.forEach(n => {
      expect((n.data as NodeData).fabricName).toBe('fe-fabric');
    });
  });
});

describe('roleLabel', () => {
  it('returns correct label for each role', () => {
    expect(roleLabel('leaf')).toBe('Leaf');
    expect(roleLabel('spine')).toBe('Spine');
    expect(roleLabel('super-spine')).toBe('SuperSpine');
    expect(roleLabel('agg1')).toBe('Agg-1');
    expect(roleLabel('agg2')).toBe('Agg-2');
    expect(roleLabel('host')).toBe('Host');
  });
});

describe('TopologyGraphService', () => {
  const svc = new TopologyGraphService();

  it('buildElements merges elements from multiple fabrics', () => {
    const collapseState = new Map([[1, true], [2, true]]);
    const { nodes } = svc.buildElements(
      [mockFabricFrontend, mockFabricBackend],
      collapseState,
    );
    // 2 nodes per fabric (collapsed 2-stage) = 4
    expect(nodes.length).toBe(4);
  });

  it('buildElements defaults to collapsed when fabricId not in map', () => {
    const collapseState = new Map<number, boolean>();
    const { nodes } = svc.buildElements([mockFabricFrontend], collapseState);
    expect(nodes.every(n => (n.data as NodeData).isGroup)).toBe(true);
  });

  it('countExpandedNodes returns layer count for collapsed fabrics', () => {
    const collapseState = new Map([[1, true]]);
    const count = svc.countExpandedNodes([mockFabricFrontend], collapseState);
    expect(count).toBe(2); // 2-stage → 2 layers
  });

  it('countExpandedNodes returns switch count for expanded fabrics', () => {
    const collapseState = new Map([[1, false]]);
    const count = svc.countExpandedNodes([mockFabricFrontend], collapseState);
    expect(count).toBe(4); // 2 leaf + 2 spine
  });

  it('countExpandedNodes sums across fabrics', () => {
    const collapseState = new Map([[1, true], [2, false]]);
    const count = svc.countExpandedNodes(
      [mockFabricFrontend, mockFabricBackend],
      collapseState,
    );
    // collapsed=2, expanded=4
    expect(count).toBe(6);
  });
});
