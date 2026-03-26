import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { KnowledgePanelComponent } from './features/knowledge/components/knowledge-panel/knowledge-panel.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, KnowledgePanelComponent],
  template: `
    <router-outlet></router-outlet>
    <app-knowledge-panel></app-knowledge-panel>
  `,
  styles: [`
    :host {
      display: block;
      height: 100dvh;
    }
  `],
})
export class App {}
