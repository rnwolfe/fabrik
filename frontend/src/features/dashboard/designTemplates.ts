import { sitesApi } from '@/api/sites';
import { superblocksApi } from '@/api/superblocks';
import { blocksApi } from '@/api/blocks';

export interface DesignTemplate {
  id: string;
  name: string;
  description: string;
  /** Returns void on success. Throws on unrecoverable error. */
  scaffold: (designId: number) => Promise<void>;
}

/**
 * Blank template — no hierarchy is created.
 * This is the default / current behaviour.
 */
const blankTemplate: DesignTemplate = {
  id: 'blank',
  name: 'Blank',
  description: 'Start with an empty canvas. Add sites, pods, and blocks manually.',
  scaffold: async (_designId: number) => {
    // no-op
  },
};

/**
 * 2-stage Clos — 1 site → 1 superblock → 1 block.
 * Each block is an independent leaf–spine fabric.
 */
const twoStageClos: DesignTemplate = {
  id: '2-stage-clos',
  name: '2-stage Clos',
  description: 'Single pod with one leaf–spine block. The simplest production-grade Clos fabric.',
  scaffold: async (designId: number) => {
    const site = await sitesApi.create(designId, { name: 'Site 1' });
    const superblock = await superblocksApi.create(site.id, { name: 'Pod A' });
    await blocksApi.create({ super_block_id: superblock.id, name: 'Block 1' });
  },
};

/**
 * 3-stage Clos — 1 site → 1 superblock → 1 block, sized for a super-spine tier.
 * Structurally identical to 2-stage at creation time; the "3-stage" designation
 * signals that super-spine aggregation will be assigned at the superblock level.
 */
const threeStageClos: DesignTemplate = {
  id: '3-stage-clos',
  name: '3-stage Clos',
  description:
    'Single pod with one block and a super-spine tier. Scales beyond a single leaf–spine plane.',
  scaffold: async (designId: number) => {
    const site = await sitesApi.create(designId, { name: 'Site 1' });
    const superblock = await superblocksApi.create(site.id, {
      name: 'Pod A',
      description: 'Super-spine aggregation pod',
    });
    await blocksApi.create({ super_block_id: superblock.id, name: 'Block 1' });
  },
};

/**
 * Pod-based fabric — 1 site → 2 superblocks → 2 blocks each (4 blocks total).
 * Models a multi-pod datacenter with isolated failure domains.
 */
const podBasedFabric: DesignTemplate = {
  id: 'pod-based',
  name: 'Pod-based fabric',
  description:
    'Two independent pods, each with two leaf–spine blocks. Ideal for large-scale multi-pod DCs.',
  scaffold: async (designId: number) => {
    const site = await sitesApi.create(designId, { name: 'Site 1' });

    const podA = await superblocksApi.create(site.id, { name: 'Pod A' });
    await blocksApi.create({ super_block_id: podA.id, name: 'Block 1' });
    await blocksApi.create({ super_block_id: podA.id, name: 'Block 2' });

    const podB = await superblocksApi.create(site.id, { name: 'Pod B' });
    await blocksApi.create({ super_block_id: podB.id, name: 'Block 3' });
    await blocksApi.create({ super_block_id: podB.id, name: 'Block 4' });
  },
};

export const DESIGN_TEMPLATES: DesignTemplate[] = [
  blankTemplate,
  twoStageClos,
  threeStageClos,
  podBasedFabric,
];
