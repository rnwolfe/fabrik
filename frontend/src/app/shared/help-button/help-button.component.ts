import { Component, Input, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { KnowledgePanelService } from '../../core/knowledge-panel.service';

/**
 * HelpButtonComponent is a reusable help icon button that opens the
 * contextual knowledge panel for a specific article.
 *
 * Usage:
 *   <app-help-button article="networking/oversubscription"></app-help-button>
 */
@Component({
  selector: 'app-help-button',
  standalone: true,
  imports: [MatButtonModule, MatIconModule, MatTooltipModule],
  template: `
    <button
      mat-icon-button
      class="help-button"
      (click)="openHelp()"
      [matTooltip]="tooltip"
      [attr.aria-label]="'Open help for ' + article"
    >
      <mat-icon>help_outline</mat-icon>
    </button>
  `,
  styles: [`
    .help-button {
      color: var(--mat-sys-on-surface-variant, #666);

      &:hover {
        color: var(--mat-sys-primary, #6200ea);
      }
    }
  `],
})
export class HelpButtonComponent {
  /** The knowledge article path to open, e.g. "networking/oversubscription". */
  @Input({ required: true }) article!: string;
  /** Tooltip text shown on hover. */
  @Input() tooltip = 'Learn more';

  private readonly panelService = inject(KnowledgePanelService);

  openHelp(): void {
    this.panelService.open(this.article);
  }
}
