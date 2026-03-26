import { Component } from '@angular/core';

@Component({
  selector: 'app-metrics',
  standalone: true,
  template: `
    <div class="coming-soon">
      <h1>Metrics</h1>
      <p>Coming soon — Oversubscription, power, and capacity metrics.</p>
    </div>
  `,
  styles: [`
    .coming-soon {
      padding: 2rem;
      text-align: center;
    }
  `],
})
export class MetricsComponent {}
