import {
  Component,
  inject,
  computed,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { trigger, state, style, transition, animate } from '@angular/animations';

import { KnowledgePanelService } from '../../../../core/knowledge-panel.service';
import { ArticleViewComponent } from '../article-view/article-view.component';

/**
 * KnowledgePanelComponent is the slide-out contextual help panel.
 * It reads state from KnowledgePanelService and renders the selected article
 * in compact mode. The user can pop out to the full viewer.
 */
@Component({
  selector: 'app-knowledge-panel',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatIconModule,
    MatTooltipModule,
    ArticleViewComponent,
  ],
  templateUrl: './knowledge-panel.component.html',
  styleUrl: './knowledge-panel.component.scss',
  animations: [
    trigger('panelSlide', [
      state('open', style({ transform: 'translateX(0)', opacity: 1 })),
      state('closed', style({ transform: 'translateX(100%)', opacity: 0 })),
      transition('closed => open', animate('200ms ease-out')),
      transition('open => closed', animate('150ms ease-in')),
    ]),
  ],
})
export class KnowledgePanelComponent {
  private readonly panelService = inject(KnowledgePanelService);
  private readonly router = inject(Router);

  readonly state = this.panelService.state;
  readonly isOpen = computed(() => this.state().isOpen);
  readonly articlePath = computed(() => this.state().articlePath);

  close(): void {
    this.panelService.close();
  }

  openFullViewer(): void {
    const path = this.articlePath();
    this.panelService.close();
    this.router.navigate(['/knowledge'], {
      queryParams: path ? { article: path } : undefined,
    });
  }
}
