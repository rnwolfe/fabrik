import { Component } from '@angular/core';

@Component({
  selector: 'app-knowledge',
  standalone: true,
  template: `
    <div class="coming-soon">
      <h1>Knowledge Base</h1>
      <p>Coming soon — Datacenter design knowledge base.</p>
    </div>
  `,
  styles: [`
    .coming-soon {
      padding: 2rem;
      text-align: center;
    }
  `],
})
export class KnowledgeComponent {}
