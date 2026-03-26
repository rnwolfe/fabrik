import { render, screen, fireEvent } from '@testing-library/angular';
import { TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { HelpButtonComponent } from './help-button.component';
import { KnowledgePanelService } from '../../core/knowledge-panel.service';

describe('HelpButtonComponent', () => {
  it('renders a help button', async () => {
    await render(HelpButtonComponent, {
      inputs: { article: 'networking/clos' },
      imports: [NoopAnimationsModule],
    });

    const btn = screen.getByRole('button', { name: /open help/i });
    expect(btn).toBeTruthy();
  });

  it('opens the knowledge panel when clicked', async () => {
    await render(HelpButtonComponent, {
      inputs: { article: 'networking/clos' },
      imports: [NoopAnimationsModule],
    });

    const panelService = TestBed.inject(KnowledgePanelService);
    expect(panelService.state().isOpen).toBe(false);

    const btn = screen.getByRole('button', { name: /open help/i });
    fireEvent.click(btn);

    expect(panelService.state().isOpen).toBe(true);
    expect(panelService.state().articlePath).toBe('networking/clos');
  });

  it('toggles panel closed when clicking same article twice', async () => {
    await render(HelpButtonComponent, {
      inputs: { article: 'networking/clos' },
      imports: [NoopAnimationsModule],
    });

    const panelService = TestBed.inject(KnowledgePanelService);
    const btn = screen.getByRole('button', { name: /open help/i });

    fireEvent.click(btn);
    expect(panelService.state().isOpen).toBe(true);

    fireEvent.click(btn);
    expect(panelService.state().isOpen).toBe(false);
  });

  it('uses custom tooltip when provided', async () => {
    const { fixture } = await render(HelpButtonComponent, {
      inputs: { article: 'networking/clos', tooltip: 'Learn about Clos' },
      imports: [NoopAnimationsModule],
    });

    const component = fixture.componentInstance;
    expect(component.tooltip).toBe('Learn about Clos');
  });
});
