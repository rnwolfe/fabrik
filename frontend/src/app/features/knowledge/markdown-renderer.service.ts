import { Injectable, PLATFORM_ID, inject } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import DOMPurify from 'dompurify';

/**
 * MarkdownRendererService converts Markdown (with Mermaid, KaTeX, and
 * syntax-highlighted code blocks) to HTML.
 *
 * It uses:
 *   - marked  — Markdown → HTML
 *   - highlight.js — syntax highlighting for fenced code blocks
 *   - mermaid — diagram rendering (post-process DOM)
 *   - katex   — math rendering (inline $...$ and block $$...$$)
 *   - dompurify — HTML sanitization before the output leaves this service
 */
@Injectable({ providedIn: 'root' })
export class MarkdownRendererService {
  private readonly platformId = inject(PLATFORM_ID);
  private _mermaidCounter = 0;

  /**
   * Converts Markdown to HTML.
   * Code blocks with language "mermaid" are wrapped in a div for later
   * post-processing by renderMermaid().
   * Math expressions ($...$ and $$...$$ ) are rendered with KaTeX.
   * The final HTML is sanitized with DOMPurify before being returned.
   */
  async render(markdown: string): Promise<string> {
    if (!isPlatformBrowser(this.platformId)) {
      return '';
    }

    const { marked, Renderer } = await import('marked');
    const hljs = (await import('highlight.js')).default;

    // Pre-process math before Markdown parsing (KaTeX)
    const processedMd = this.preProcessMath(markdown);

    const renderer = new Renderer();

    // Override code rendering: mermaid blocks get special treatment.
    renderer.code = ({ text, lang }: { text: string; lang?: string }) => {
      if (lang === 'mermaid') {
        // Escape the Mermaid source when inserting into a wrapper element
        // to prevent XSS via crafted diagram definitions.
        return `<div class="mermaid-diagram">${this.escapeHtml(text)}</div>`;
      }
      if (lang && hljs.getLanguage(lang)) {
        try {
          const highlighted = hljs.highlight(text, { language: lang }).value;
          return `<pre><code class="hljs language-${this.escapeAttr(lang)}">${highlighted}</code></pre>`;
        } catch {
          // fall through to default
        }
      }
      return `<pre><code class="hljs">${hljs.highlightAuto(text).value}</code></pre>`;
    };

    // Override links to flag internal knowledge links for resolution.
    // Attribute values are HTML-escaped to prevent XSS.
    renderer.link = ({ href, title, text }: { href: string; title?: string | null; text: string }) => {
      const safeHref = this.escapeAttr(href ?? '');
      if (href && !href.startsWith('http') && !href.startsWith('#')) {
        // Internal link — mark with data attribute for Angular router
        const t = title ? ` title="${this.escapeAttr(title)}"` : '';
        return `<a href="${safeHref}"${t} data-knowledge-link="true">${text}</a>`;
      }
      const t = title ? ` title="${this.escapeAttr(title)}"` : '';
      return `<a href="${safeHref}"${t} target="_blank" rel="noopener noreferrer">${text}</a>`;
    };

    marked.use({ renderer });

    let html = await marked(processedMd);
    html = this.postProcessMath(html);

    // Final DOMPurify pass — catches anything that slipped through the renderers.
    // ADD_ATTR: allow target so external links keep target="_blank" as set by
    // the renderer. rel="noopener noreferrer" is always present to mitigate
    // tab-napping even when target is allowed.
    html = DOMPurify.sanitize(html, {
      USE_PROFILES: { html: true },
      ADD_ATTR: ['target', 'data-knowledge-link'],
    });

    return html;
  }

  /**
   * Initialises Mermaid and renders all .mermaid-diagram elements in the
   * given container. Call this after the rendered HTML is inserted into the DOM.
   * Mermaid reads graphDef from textContent (not innerHTML) so the escaped
   * source is decoded automatically by the browser before Mermaid sees it.
   */
  async renderMermaid(container: HTMLElement): Promise<void> {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }

    const elements = container.querySelectorAll<HTMLElement>('.mermaid-diagram');
    if (elements.length === 0) {
      return;
    }

    try {
      const mermaid = (await import('mermaid')).default;
      mermaid.initialize({ startOnLoad: false, theme: 'default' });

      for (const el of Array.from(elements)) {
        // textContent decodes the HTML entities we inserted, giving Mermaid
        // the original diagram definition string.
        const graphDef = el.textContent || '';
        const id = `mermaid-${++this._mermaidCounter}`;
        try {
          const { svg } = await mermaid.render(id, graphDef);
          el.innerHTML = svg;
        } catch (err) {
          el.innerHTML = `<div class="mermaid-error">Diagram error: ${this.escapeHtml(String(err))}</div>`;
        }
      }
    } catch (err) {
      if (typeof ngDevMode !== 'undefined' && ngDevMode) {
        console.warn('MarkdownRendererService: mermaid render failed', err);
      }
    }
  }

  /** Pre-processes KaTeX math: replaces $...$ and $$...$$ with placeholders. */
  private preProcessMath(md: string): string {
    // Replace $$...$$ (display math) with placeholder.
    md = md.replace(/\$\$([\s\S]+?)\$\$/g, (_, expr) => {
      return `<katex-display>${this.escapeHtml(expr.trim())}</katex-display>`;
    });
    // Replace $...$ (inline math) with placeholder.
    md = md.replace(/\$([^$\n]+?)\$/g, (_, expr) => {
      return `<katex-inline>${this.escapeHtml(expr.trim())}</katex-inline>`;
    });
    return md;
  }

  /** Post-processes the rendered HTML to replace KaTeX placeholders with rendered math. */
  private postProcessMath(html: string): string {
    // Try to import katex synchronously via the cached module.
    // KaTeX rendering is done inline during this post-process step.
    try {
      // Use a synchronous require-style import by accessing the global.
      // If katex is not yet loaded, we fall back to raw LaTeX display.
      const katex = (window as Window & { __fabrik_katex?: { renderToString(expr: string, opts: object): string } }).__fabrik_katex;
      if (!katex) {
        return html;
      }

      html = html.replace(/<katex-display>([\s\S]+?)<\/katex-display>/g, (_, expr) => {
        try {
          return katex.renderToString(this.unescapeHtml(expr), { displayMode: true, throwOnError: false });
        } catch {
          return `<span class="katex-error">$$${this.unescapeHtml(expr)}$$</span>`;
        }
      });

      html = html.replace(/<katex-inline>([^<]+?)<\/katex-inline>/g, (_, expr) => {
        try {
          return katex.renderToString(this.unescapeHtml(expr), { displayMode: false, throwOnError: false });
        } catch {
          return `<span class="katex-error">$${this.unescapeHtml(expr)}$</span>`;
        }
      });
    } catch {
      // katex not available yet — leave placeholders as raw text
    }
    return html;
  }

  /**
   * Loads KaTeX asynchronously and stores it in window for synchronous access
   * during HTML post-processing. Call this once on app init or before rendering.
   */
  async loadKatex(): Promise<void> {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }
    const katex = (await import('katex')).default;
    (window as Window & { __fabrik_katex?: unknown }).__fabrik_katex = katex;
  }

  /** Escapes a string for safe insertion as HTML text content. */
  private escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  /** Escapes a string for safe use as an HTML attribute value (inside double quotes). */
  private escapeAttr(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/'/g, '&#x27;');
  }

  private unescapeHtml(text: string): string {
    return text
      .replace(/&amp;/g, '&')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&quot;/g, '"');
  }
}
