/**
 * HierarchyTree unit tests.
 *
 * Note: Full component rendering tests would require jsdom + @testing-library/react.
 * The current test setup (vitest without a DOM environment) supports logic-only tests.
 * Install the following to enable component tests:
 *   npm install -D @testing-library/react @testing-library/user-event jsdom
 * Then add `environment: 'jsdom'` to the vitest config.
 */

import { describe, it, expect } from 'vitest';
import type { DesignHierarchy, HierarchySite, HierarchySuperBlock, HierarchyBlock } from '@/models';

// ─── Type shape tests ─────────────────────────────────────────────────────────
// Verify that the DesignHierarchy types compose correctly and the hierarchy
// structure matches what the API returns.

describe('DesignHierarchy types', () => {
  it('empty hierarchy has no sites', () => {
    const h: DesignHierarchy = { design_id: 1, sites: [] };
    expect(h.sites).toHaveLength(0);
  });

  it('single-site hierarchy structure is valid', () => {
    const block: HierarchyBlock = {
      id: 10,
      super_block_id: 5,
      name: 'Block A',
      description: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };
    const sb: HierarchySuperBlock = {
      id: 5,
      site_id: 2,
      name: 'Hall 1',
      description: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      blocks: [block],
    };
    const site: HierarchySite = {
      id: 2,
      design_id: 1,
      name: 'Site 1',
      description: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      super_blocks: [sb],
    };
    const h: DesignHierarchy = { design_id: 1, sites: [site] };

    expect(h.sites).toHaveLength(1);
    expect(h.sites[0].super_blocks).toHaveLength(1);
    expect(h.sites[0].super_blocks[0].blocks).toHaveLength(1);
    expect(h.sites[0].super_blocks[0].blocks[0].name).toBe('Block A');
  });

  it('multi-site hierarchy structure is valid', () => {
    const makeSite = (id: number, name: string): HierarchySite => ({
      id,
      design_id: 1,
      name,
      description: '',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      super_blocks: [],
    });

    const h: DesignHierarchy = {
      design_id: 1,
      sites: [makeSite(1, 'Site A'), makeSite(2, 'Site B')],
    };

    expect(h.sites).toHaveLength(2);
    expect(h.sites[0].name).toBe('Site A');
    expect(h.sites[1].name).toBe('Site B');
  });
});

// ─── allSuperBlocks derivation logic ─────────────────────────────────────────
// Replicate the allSuperBlocks derivation from HierarchyTree to test it in isolation.

function deriveAllSuperBlocks(hierarchy: DesignHierarchy) {
  return hierarchy.sites.flatMap((site) =>
    site.super_blocks.map((sb) => ({
      id: sb.id,
      name: sb.name,
      siteId: site.id,
      siteName: site.name,
    }))
  );
}

describe('allSuperBlocks derivation', () => {
  it('returns empty array when no sites', () => {
    const h: DesignHierarchy = { design_id: 1, sites: [] };
    expect(deriveAllSuperBlocks(h)).toHaveLength(0);
  });

  it('flattens super-blocks from multiple sites', () => {
    const h: DesignHierarchy = {
      design_id: 1,
      sites: [
        {
          id: 1,
          design_id: 1,
          name: 'Site A',
          description: '',
          created_at: '',
          updated_at: '',
          super_blocks: [
            { id: 10, site_id: 1, name: 'SB A1', description: '', created_at: '', updated_at: '', blocks: [] },
            { id: 11, site_id: 1, name: 'SB A2', description: '', created_at: '', updated_at: '', blocks: [] },
          ],
        },
        {
          id: 2,
          design_id: 1,
          name: 'Site B',
          description: '',
          created_at: '',
          updated_at: '',
          super_blocks: [
            { id: 20, site_id: 2, name: 'SB B1', description: '', created_at: '', updated_at: '', blocks: [] },
          ],
        },
      ],
    };

    const all = deriveAllSuperBlocks(h);
    expect(all).toHaveLength(3);
    expect(all[0]).toEqual({ id: 10, name: 'SB A1', siteId: 1, siteName: 'Site A' });
    expect(all[2]).toEqual({ id: 20, name: 'SB B1', siteId: 2, siteName: 'Site B' });
  });
});
