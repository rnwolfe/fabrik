import { Component } from '@angular/core';

@Component({
  selector: 'app-topology',
  standalone: true,
  template: `
    <div class="coming-soon">
      <h1>Design</h1>
      <p>Coming soon — Clos fabric designer and topology visualization.</p>
    </div>
  `,
  styles: [`
    .coming-soon {
      padding: 2rem;
      text-align: center;
    }
  `],
})
export class TopologyComponent {}
