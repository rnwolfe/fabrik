import { render, screen } from '@testing-library/angular';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideAnimations } from '@angular/platform-browser/animations';

import { KnowledgePanelComponent } from './knowledge-panel.component';
import { KnowledgePanelService } from '../../../../core/knowledge-panel.service';
import { MarkdownRendererService } from '../../markdown-renderer.service';

const MOCK_ARTICLE = {
  path: 'networking/clos',
  title: 'Clos Fabric Fundamentals',
  category: 'networking',
  tags: ['clos'],
  content: '# Clos Basics\n\nBody.',
};

class MockMarkdownRendererService {
  async render(markdown: string): Promise<string> {
    return Promise.resolve(`<p>${markdown}</p>`);
  }
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async renderMermaid(_el: HTMLElement): Promise<void> {
    return Promise.resolve();
  }
  async loadKatex(): Promise<void> {
    return Promise.resolve();
  }
}

describe('KnowledgePanelComponent', () => {
  let panelService: KnowledgePanelService;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await render(KnowledgePanelComponent, {
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideAnimations(),
        provideRouter([{ path: 'knowledge', component: KnowledgePanelComponent }]),
        { provide: MarkdownRendererService, useClass: MockMarkdownRendererService },
      ],
    });
    panelService = TestBed.inject(KnowledgePanelService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    try { httpMock.verify(); } catch { /* ok */ }
  });

  it('panel is not visible when closed', () => {
    expect(panelService.state().isOpen).toBe(false);
    const backdrop = document.querySelector('.panel-backdrop');
    expect(backdrop).toBeNull();
  });

  it('shows panel when open', async () => {
    panelService.open('networking/clos');
    await new Promise(r => setTimeout(r, 50));

    const req = httpMock.match('/api/knowledge/networking/clos');
    if (req.length > 0) {
      req.forEach(r => r.flush(MOCK_ARTICLE));
    }

    expect(screen.getByRole('complementary')).toBeTruthy();
  });

  it('has close button when open', async () => {
    panelService.open('networking/clos');
    await new Promise(r => setTimeout(r, 50));

    const req = httpMock.match('/api/knowledge/networking/clos');
    if (req.length > 0) {
      req.forEach(r => r.flush(MOCK_ARTICLE));
    }

    const closeBtn = screen.getByRole('button', { name: /close help panel/i });
    expect(closeBtn).toBeTruthy();
  });

  it('has open-in-full-viewer button', async () => {
    panelService.open('networking/clos');
    await new Promise(r => setTimeout(r, 50));

    const req = httpMock.match('/api/knowledge/networking/clos');
    if (req.length > 0) {
      req.forEach(r => r.flush(MOCK_ARTICLE));
    }

    const popoutBtn = screen.getByRole('button', { name: /open in full/i });
    expect(popoutBtn).toBeTruthy();
  });
});
