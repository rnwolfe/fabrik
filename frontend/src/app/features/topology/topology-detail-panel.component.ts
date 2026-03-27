import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatChipsModule } from '@angular/material/chips';

import { NodeData, EdgeData, roleLabel } from './topology-graph.service';

export type DetailItem =
  | { kind: 'node'; data: NodeData }
  | { kind: 'edge'; data: EdgeData };

@Component({
  selector: 'app-topology-detail-panel',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    MatIconModule,
    MatButtonModule,
    MatDividerModule,
    MatTooltipModule,
    MatChipsModule,
  ],
  templateUrl: './topology-detail-panel.component.html',
  styleUrl: './topology-detail-panel.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class TopologyDetailPanelComponent {
  @Input() item: DetailItem | null = null;
  @Output() closed = new EventEmitter<void>();

  close(): void {
    this.closed.emit();
  }

  roleLabel = roleLabel;

  get nodeData(): NodeData | null {
    return this.item?.kind === 'node' ? this.item.data : null;
  }

  get edgeData(): EdgeData | null {
    return this.item?.kind === 'edge' ? this.item.data : null;
  }

  utilPercent(util: number | undefined): string {
    if (util === undefined) return 'N/A';
    return `${Math.round(util * 100)}%`;
  }

  utilPct(util: number | undefined): number {
    return util !== undefined ? Math.round(util * 100) : 0;
  }

  utilClass(util: number | undefined): string {
    if (util === undefined) return '';
    if (util >= 0.9) return 'util-critical';
    if (util >= 0.7) return 'util-warning';
    return 'util-healthy';
  }
}
