import { render, screen } from '@testing-library/angular';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ArticleViewComponent } from './article-view.component';
import { MarkdownRendererService } from '../../markdown-renderer.service';

const MOCK_ARTICLE = {
  path: 'networking/clos',
  title: 'Clos Fabric Fundamentals',
  category: 'networking',
  tags: ['clos', 'fabric'],
  content: '# Clos Fabric Fundamentals\n\nA Clos network is a multistage switching architecture.',
};

/** Stub MarkdownRendererService — avoids loading heavy libs (marked, mermaid, katex) in tests. */
class MockMarkdownRendererService {
  async render(markdown: string): Promise<string> {
    return Promise.resolve(`<p>${markdown.replace(/^#+\s*/gm, '')}</p>`);
  }
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async renderMermaid(_el: HTMLElement): Promise<void> {
    return Promise.resolve();
  }
  async loadKatex(): Promise<void> {
    return Promise.resolve();
  }
}

async function renderArticle(path = 'networking/clos') {
  const result = await render(ArticleViewComponent, {
    inputs: { articlePath: path },
    imports: [NoopAnimationsModule],
    providers: [
      provideHttpClient(),
      provideHttpClientTesting(),
      { provide: MarkdownRendererService, useClass: MockMarkdownRendererService },
    ],
  });
  const httpMock = TestBed.inject(HttpTestingController);
  return { ...result, httpMock };
}

describe('ArticleViewComponent', () => {
  afterEach(() => {
    try {
      TestBed.inject(HttpTestingController).verify();
    } catch { /* may already be verified */ }
  });

  it('shows loading state initially', async () => {
    const { httpMock } = await renderArticle();
    const spinner = document.querySelector('mat-spinner');
    expect(spinner).toBeTruthy();
    httpMock.expectOne('/api/knowledge/networking/clos').flush(MOCK_ARTICLE);
  });

  it('renders article title and category after load', async () => {
    const { httpMock } = await renderArticle();
    httpMock.expectOne('/api/knowledge/networking/clos').flush(MOCK_ARTICLE);

    await screen.findByText('Clos Fabric Fundamentals');
    expect(screen.getByText('Networking')).toBeTruthy();
  });

  it('renders article tags', async () => {
    const { httpMock } = await renderArticle();
    httpMock.expectOne('/api/knowledge/networking/clos').flush(MOCK_ARTICLE);

    await screen.findByText('clos');
    expect(screen.getByText('fabric')).toBeTruthy();
  });

  it('shows not-found state on 404', async () => {
    const { httpMock } = await renderArticle('does/not/exist');
    httpMock.expectOne('/api/knowledge/does/not/exist').flush(
      { error: 'article not found' },
      { status: 404, statusText: 'Not Found' }
    );

    await screen.findByText('Article not found');
    expect(screen.getByText(/does\/not\/exist/)).toBeTruthy();
  });

  it('shows error state on server error', async () => {
    const { httpMock } = await renderArticle();
    httpMock.expectOne('/api/knowledge/networking/clos').error(new ErrorEvent('Network error'));

    await screen.findByText('Failed to load article.');
  });
});
