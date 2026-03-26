import { Component } from '@angular/core';

@Component({
  selector: 'app-catalog',
  standalone: true,
  template: `
    <div class="coming-soon">
      <h1>Catalog</h1>
      <p>Coming soon — Hardware catalog management.</p>
    </div>
  `,
  styles: [`
    .coming-soon {
      padding: 2rem;
      text-align: center;
    }
  `],
})
export class CatalogComponent {}
