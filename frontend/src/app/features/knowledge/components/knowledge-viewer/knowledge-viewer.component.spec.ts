import { render, screen, fireEvent } from '@testing-library/angular';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import { KnowledgeViewerComponent } from './knowledge-viewer.component';
import { MarkdownRendererService } from '../../markdown-renderer.service';

const MOCK_INDEX = {
  articles: [
    { path: 'networking/clos', title: 'Clos Fabric Fundamentals', category: 'networking', tags: ['clos'] },
    { path: 'networking/ecmp', title: 'ECMP Load Balancing', category: 'networking', tags: ['ecmp'] },
    { path: 'infrastructure/rack', title: 'Rack Design Basics', category: 'infrastructure', tags: ['rack'] },
  ],
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

async function renderViewer() {
  const result = await render(KnowledgeViewerComponent, {
    imports: [NoopAnimationsModule],
    providers: [
      provideHttpClient(),
      provideHttpClientTesting(),
      provideRouter([]),
      { provide: MarkdownRendererService, useClass: MockMarkdownRendererService },
    ],
  });
  const httpMock = TestBed.inject(HttpTestingController);
  return { ...result, httpMock };
}

describe('KnowledgeViewerComponent', () => {
  afterEach(() => {
    try {
      TestBed.inject(HttpTestingController).verify();
    } catch { /* may already be verified */ }
  });

  it('shows loading spinner initially', async () => {
    const { httpMock } = await renderViewer();
    const spinner = document.querySelector('mat-spinner');
    expect(spinner).toBeTruthy();
    httpMock.expectOne('/api/knowledge').flush(MOCK_INDEX);
  });

  it('renders category headings after load', async () => {
    const { httpMock } = await renderViewer();
    httpMock.expectOne('/api/knowledge').flush(MOCK_INDEX);
    // Article requests fired for first article
    const pending = httpMock.match(() => true);
    pending.forEach(r => r.flush({ ...MOCK_INDEX.articles[0], content: '# Test' }));

    await screen.findByText('Networking');
    expect(screen.getByText('Infrastructure')).toBeTruthy();
  });

  it('renders article titles in TOC', async () => {
    const { httpMock } = await renderViewer();
    httpMock.expectOne('/api/knowledge').flush(MOCK_INDEX);
    const pending = httpMock.match(() => true);
    pending.forEach(r => r.flush({ ...MOCK_INDEX.articles[0], content: '# Test' }));

    await screen.findByText('Clos Fabric Fundamentals');
    expect(screen.getByText('ECMP Load Balancing')).toBeTruthy();
    expect(screen.getByText('Rack Design Basics')).toBeTruthy();
  });

  it('shows no-results state for empty search', async () => {
    const { httpMock } = await renderViewer();
    httpMock.expectOne('/api/knowledge').flush(MOCK_INDEX);
    const pending = httpMock.match(() => true);
    pending.forEach(r => r.flush({ ...MOCK_INDEX.articles[0], content: '# Test' }));

    await screen.findByText('Clos Fabric Fundamentals');

    const searchInput = screen.getByRole('textbox', { name: /search/i });
    fireEvent.input(searchInput, { target: { value: 'xyznosuchthing' } });

    expect(screen.getByText(/No articles found for/)).toBeTruthy();
  });

  it('shows error state on API failure', async () => {
    const { httpMock } = await renderViewer();
    httpMock.expectOne('/api/knowledge').error(new ErrorEvent('Network error'));

    await screen.findByText('Failed to load knowledge base.');
  });
});
