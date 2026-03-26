import { TestBed } from '@angular/core/testing';
import { PLATFORM_ID } from '@angular/core';
import { MarkdownRendererService } from './markdown-renderer.service';

describe('MarkdownRendererService', () => {
  let service: MarkdownRendererService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        { provide: PLATFORM_ID, useValue: 'browser' },
      ],
    });
    service = TestBed.inject(MarkdownRendererService);
  });

  it('renders headings', async () => {
    const html = await service.render('# Hello World');
    expect(html).toContain('<h1');
    expect(html).toContain('Hello World');
  });

  it('renders bold and italic text', async () => {
    const html = await service.render('**bold** and _italic_');
    expect(html).toContain('<strong>bold</strong>');
    expect(html).toContain('<em>italic</em>');
  });

  it('renders code blocks with syntax highlighting', async () => {
    const html = await service.render('```javascript\nconst x = 1;\n```');
    expect(html).toContain('hljs');
  });

  it('wraps mermaid blocks in .mermaid-diagram', async () => {
    const html = await service.render('```mermaid\ngraph TD\n  A --> B\n```');
    expect(html).toContain('class="mermaid-diagram"');
  });

  it('marks internal links with data-knowledge-link attribute', async () => {
    const html = await service.render('[Clos Basics](networking/clos)');
    expect(html).toContain('data-knowledge-link="true"');
    expect(html).toContain('href="networking/clos"');
  });

  it('adds target=_blank to external links', async () => {
    const html = await service.render('[External](https://example.com)');
    expect(html).toContain('target="_blank"');
  });

  it('returns empty string on server platform', async () => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      providers: [
        { provide: PLATFORM_ID, useValue: 'server' },
      ],
    });
    const serverService = TestBed.inject(MarkdownRendererService);
    const html = await serverService.render('# Hello');
    expect(html).toBe('');
  });
});
