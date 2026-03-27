import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

/**
 * RacksComponent is the shell component for the racks feature.
 * It delegates all content to child routes via RouterOutlet.
 */
@Component({
  selector: 'app-racks',
  standalone: true,
  imports: [RouterOutlet],
  template: `<router-outlet />`,
})
export class RacksComponent {}
