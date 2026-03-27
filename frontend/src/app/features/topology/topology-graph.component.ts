import {
  Component,
  ElementRef,
  Input,
  OnChanges,
  OnDestroy,
  OnInit,
  Output,
  EventEmitter,
  ViewChild,
  signal,
  inject,
  SimpleChanges,
  NgZone,
  ChangeDetectionStrategy,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatChipsModule } from '@angular/material/chips';

// Cytoscape is loaded dynamically to avoid SSR issues
import cytoscape from 'cytoscape';
// @ts-expect-error – cytoscape-dagre has no default export type
import cytoscapeDagre from 'cytoscape-dagre';

import { FabricResponse } from '../../models/fabric';
import { TopologyGraphService, NodeData, EdgeData } from './topology-graph.service';

cytoscape.use(cytoscapeDagre);

export interface NodeClickEvent {
  data: NodeData;
}

export interface EdgeClickEvent {
  data: EdgeData;
}

@Component({
  selector: 'app-topology-graph',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatIconModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
    MatChipsModule,
  ],
  templateUrl: './topology-graph.component.html',
  styleUrl: './topology-graph.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class TopologyGraphComponent implements OnInit, OnChanges, OnDestroy {
  @ViewChild('graphContainer', { static: true })
  graphContainerRef!: ElementRef<HTMLDivElement>;

  @Input() fabrics: FabricResponse[] = [];
  @Input() expandAll = false;

  @Output() nodeClick = new EventEmitter<NodeClickEvent>();
  @Output() edgeClick = new EventEmitter<EdgeClickEvent>();

  private readonly _svc = inject(TopologyGraphService);
  private readonly _zone = inject(NgZone);

  private _cy: cytoscape.Core | null = null;
  /** collapse state per fabric id: true = collapsed */
  readonly collapseState = signal(new Map<number, boolean>());

  isRendering = signal(false);
  hoveredNodeId = signal<string | null>(null);
  nodeWarning = signal<string | null>(null);

  ngOnInit(): void {
    this._initCytoscape();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['fabrics'] || changes['expandAll']) {
      if (changes['expandAll']) {
        this._applyExpandAll(this.expandAll);
      }
      if (this._cy) {
        this._renderGraph();
      }
    }
  }

  ngOnDestroy(): void {
    this._cy?.destroy();
    this._cy = null;
  }

  private _applyExpandAll(expandAll: boolean): void {
    const newMap = new Map<number, boolean>();
    this.fabrics.forEach(f => newMap.set(f.id, !expandAll));
    this.collapseState.set(newMap);
  }

  toggleCollapse(fabricId: number): void {
    const map = new Map(this.collapseState());
    map.set(fabricId, !(map.get(fabricId) ?? true));
    this.collapseState.set(map);
    this._renderGraph();
  }

  private _initCytoscape(): void {
    this._zone.runOutsideAngular(() => {
      try {
        this._cy = cytoscape({
          container: this.graphContainerRef.nativeElement,
          elements: [],
          style: this._buildStyle(),
          layout: { name: 'preset' },
          minZoom: 0.1,
          maxZoom: 4,
          wheelSensitivity: 0.3,
          userZoomingEnabled: true,
          userPanningEnabled: true,
          boxSelectionEnabled: false,
        });

        this._attachEvents();
      } catch {
        // Canvas not available in test environments — skip initialization
      }
    });

    if (this.fabrics.length && this._cy) {
      this._renderGraph();
    }
  }

  private _renderGraph(): void {
    if (!this._cy) {
      this.isRendering.set(false);
      return;
    }

    this.isRendering.set(true);
    const totalNodes = this._svc.countExpandedNodes(this.fabrics, this.collapseState());
    if (totalNodes > 500) {
      this.nodeWarning.set(
        `Warning: ${totalNodes} nodes – rendering may be slow. Consider collapsing some fabrics.`,
      );
    } else {
      this.nodeWarning.set(null);
    }

    const { nodes, edges } = this._svc.buildElements(this.fabrics, this.collapseState());

    this._zone.runOutsideAngular(() => {
      const cy = this._cy!;
      cy.elements().remove();
      cy.add([...nodes, ...edges]);

      cy.layout(this._buildLayout()).run();
      cy.ready(() => {
        cy.fit(undefined, 40);
        this._zone.run(() => this.isRendering.set(false));
      });
    });
  }

  fitGraph(): void {
    this._zone.runOutsideAngular(() => this._cy?.fit(undefined, 40));
  }

  private _attachEvents(): void {
    if (!this._cy) return;
    const cy = this._cy;

    cy.on('tap', 'node', evt => {
      const data = evt.target.data() as NodeData;
      this._zone.run(() => {
        this.nodeClick.emit({ data });
      });
    });

    cy.on('tap', 'edge', evt => {
      const data = evt.target.data() as EdgeData;
      this._zone.run(() => {
        this.edgeClick.emit({ data });
      });
    });

    cy.on('mouseover', 'node', evt => {
      const id = evt.target.id() as string;
      this._zone.run(() => this.hoveredNodeId.set(id));
    });

    cy.on('mouseout', 'node', () => {
      this._zone.run(() => this.hoveredNodeId.set(null));
    });
  }

  private _buildLayout(): cytoscape.LayoutOptions {
    return {
      name: 'dagre',
      rankDir: 'BT',
      nodeSep: 60,
      rankSep: 80,
      padding: 40,
      animate: false,
    } as cytoscape.LayoutOptions;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private _buildStyle(): any[] {
    return [
      {
        selector: 'node',
        style: {
          label: 'data(label)',
          'text-valign': 'center',
          'text-halign': 'center',
          'text-wrap': 'wrap',
          'text-max-width': '80px',
          'font-size': '11px',
          width: 70,
          height: 40,
          shape: 'round-rectangle',
          'background-color': '#90caf9',
          'border-width': 2,
          'border-color': '#42a5f5',
          color: '#0d1b2a',
        },
      },
      {
        selector: 'node[tier = "backend"]',
        style: {
          'background-color': '#a5d6a7',
          'border-color': '#66bb6a',
        },
      },
      // Role-based border accents for frontend
      {
        selector: 'node[role = "spine"][tier = "frontend"]',
        style: { 'background-color': '#64b5f6', 'border-color': '#1976d2' },
      },
      {
        selector: 'node[role = "super-spine"][tier = "frontend"]',
        style: { 'background-color': '#4dd0e1', 'border-color': '#00838f' },
      },
      {
        selector: 'node[role = "leaf"][tier = "frontend"]',
        style: { 'background-color': '#90caf9', 'border-color': '#42a5f5' },
      },
      // Role-based for backend
      {
        selector: 'node[role = "spine"][tier = "backend"]',
        style: { 'background-color': '#81c784', 'border-color': '#388e3c' },
      },
      {
        selector: 'node[role = "super-spine"][tier = "backend"]',
        style: { 'background-color': '#80cbc4', 'border-color': '#00695c' },
      },
      {
        selector: 'node[role = "leaf"][tier = "backend"]',
        style: { 'background-color': '#a5d6a7', 'border-color': '#66bb6a' },
      },
      // Utilization color coding
      {
        selector: 'node[utilLevel = "warning"]',
        style: { 'border-color': '#f57c00', 'border-width': 3 },
      },
      {
        selector: 'node[utilLevel = "critical"]',
        style: { 'border-color': '#c62828', 'border-width': 4 },
      },
      // Group nodes (collapsed)
      {
        selector: 'node[?isGroup]',
        style: {
          width: 90,
          height: 50,
          'font-size': '12px',
          'font-weight': 'bold',
          'border-style': 'dashed',
        },
      },
      {
        selector: 'edge',
        style: {
          width: 1.5,
          'line-color': '#b0bec5',
          'target-arrow-color': '#b0bec5',
          'target-arrow-shape': 'none',
          'curve-style': 'bezier',
          opacity: 0.7,
        },
      },
      {
        selector: ':selected',
        style: {
          'border-color': '#ff6f00',
          'border-width': 3,
          'line-color': '#ff6f00',
        },
      },
    ];
  }
}
