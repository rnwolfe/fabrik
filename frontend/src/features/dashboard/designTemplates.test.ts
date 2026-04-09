/// <reference types="vitest/globals" />
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { DESIGN_TEMPLATES } from './designTemplates';

// ── Mock API modules ─────────────────────────────────────────────────────────

const mockSiteCreate = vi.fn();
const mockSuperblockCreate = vi.fn();
const mockBlockCreate = vi.fn();

vi.mock('@/api/sites', () => ({
  sitesApi: {
    create: (...args: unknown[]) => mockSiteCreate(...args),
  },
}));

vi.mock('@/api/superblocks', () => ({
  superblocksApi: {
    create: (...args: unknown[]) => mockSuperblockCreate(...args),
  },
}));

vi.mock('@/api/blocks', () => ({
  blocksApi: {
    create: (...args: unknown[]) => mockBlockCreate(...args),
  },
  scaffoldApi: {},
  superBlocksApi: {},
}));

// ── Helpers ──────────────────────────────────────────────────────────────────

function getTemplate(id: string) {
  const t = DESIGN_TEMPLATES.find((t) => t.id === id);
  if (!t) throw new Error(`Template "${id}" not found`);
  return t;
}

beforeEach(() => {
  vi.clearAllMocks();
});

// ── Template list ────────────────────────────────────────────────────────────

describe('DESIGN_TEMPLATES', () => {
  it('exports exactly 4 templates', () => {
    expect(DESIGN_TEMPLATES).toHaveLength(4);
  });

  it('includes blank, 2-stage-clos, 3-stage-clos, and pod-based', () => {
    const ids = DESIGN_TEMPLATES.map((t) => t.id);
    expect(ids).toContain('blank');
    expect(ids).toContain('2-stage-clos');
    expect(ids).toContain('3-stage-clos');
    expect(ids).toContain('pod-based');
  });

  it('blank is the first entry (default)', () => {
    expect(DESIGN_TEMPLATES[0].id).toBe('blank');
  });

  it('every template has id, name, description, and scaffold function', () => {
    for (const t of DESIGN_TEMPLATES) {
      expect(typeof t.id).toBe('string');
      expect(typeof t.name).toBe('string');
      expect(typeof t.description).toBe('string');
      expect(typeof t.scaffold).toBe('function');
    }
  });
});

// ── blank ────────────────────────────────────────────────────────────────────

describe('blank template', () => {
  it('scaffold makes no API calls', async () => {
    await getTemplate('blank').scaffold(1);
    expect(mockSiteCreate).not.toHaveBeenCalled();
    expect(mockSuperblockCreate).not.toHaveBeenCalled();
    expect(mockBlockCreate).not.toHaveBeenCalled();
  });

  it('scaffold resolves without error', async () => {
    await expect(getTemplate('blank').scaffold(42)).resolves.toBeUndefined();
  });
});

// ── 2-stage Clos ─────────────────────────────────────────────────────────────

describe('2-stage Clos template', () => {
  beforeEach(() => {
    mockSiteCreate.mockResolvedValue({ id: 10, name: 'Site 1' });
    mockSuperblockCreate.mockResolvedValue({ id: 20, name: 'Pod A' });
    mockBlockCreate.mockResolvedValue({ block: { id: 30, name: 'Block 1' } });
  });

  it('creates 1 site, 1 superblock, 1 block on success path', async () => {
    await getTemplate('2-stage-clos').scaffold(1);
    expect(mockSiteCreate).toHaveBeenCalledTimes(1);
    expect(mockSuperblockCreate).toHaveBeenCalledTimes(1);
    expect(mockBlockCreate).toHaveBeenCalledTimes(1);
  });

  it('creates site named "Site 1"', async () => {
    await getTemplate('2-stage-clos').scaffold(1);
    expect(mockSiteCreate).toHaveBeenCalledWith(1, expect.objectContaining({ name: 'Site 1' }));
  });

  it('creates superblock named "Pod A" under the site', async () => {
    await getTemplate('2-stage-clos').scaffold(1);
    expect(mockSuperblockCreate).toHaveBeenCalledWith(10, expect.objectContaining({ name: 'Pod A' }));
  });

  it('creates block named "Block 1" under the superblock', async () => {
    await getTemplate('2-stage-clos').scaffold(1);
    expect(mockBlockCreate).toHaveBeenCalledWith(
      expect.objectContaining({ super_block_id: 20, name: 'Block 1' }),
    );
  });

  it('does not pre-assign leaf_model_id or spine_model_id', async () => {
    await getTemplate('2-stage-clos').scaffold(1);
    const call = mockBlockCreate.mock.calls[0][0] as Record<string, unknown>;
    expect(call.leaf_model_id).toBeUndefined();
    expect(call.spine_model_id).toBeUndefined();
  });

  it('throws when site creation fails (partial failure path)', async () => {
    mockSiteCreate.mockRejectedValue(new Error('network error'));
    await expect(getTemplate('2-stage-clos').scaffold(1)).rejects.toThrow('network error');
  });

  it('throws when superblock creation fails (partial failure path)', async () => {
    mockSuperblockCreate.mockRejectedValue(new Error('superblock error'));
    await expect(getTemplate('2-stage-clos').scaffold(1)).rejects.toThrow('superblock error');
    expect(mockBlockCreate).not.toHaveBeenCalled();
  });

  it('throws when block creation fails (partial failure path)', async () => {
    mockBlockCreate.mockRejectedValue(new Error('block error'));
    await expect(getTemplate('2-stage-clos').scaffold(1)).rejects.toThrow('block error');
  });
});

// ── 3-stage Clos ─────────────────────────────────────────────────────────────

describe('3-stage Clos template', () => {
  beforeEach(() => {
    mockSiteCreate.mockResolvedValue({ id: 11, name: 'Site 1' });
    mockSuperblockCreate.mockResolvedValue({ id: 21, name: 'Pod A' });
    mockBlockCreate.mockResolvedValue({ block: { id: 31, name: 'Block 1' } });
  });

  it('creates 1 site, 1 superblock, 1 block on success path', async () => {
    await getTemplate('3-stage-clos').scaffold(2);
    expect(mockSiteCreate).toHaveBeenCalledTimes(1);
    expect(mockSuperblockCreate).toHaveBeenCalledTimes(1);
    expect(mockBlockCreate).toHaveBeenCalledTimes(1);
  });

  it('creates site named "Site 1"', async () => {
    await getTemplate('3-stage-clos').scaffold(2);
    expect(mockSiteCreate).toHaveBeenCalledWith(2, expect.objectContaining({ name: 'Site 1' }));
  });

  it('creates superblock named "Pod A"', async () => {
    await getTemplate('3-stage-clos').scaffold(2);
    expect(mockSuperblockCreate).toHaveBeenCalledWith(11, expect.objectContaining({ name: 'Pod A' }));
  });

  it('does not pre-assign devices', async () => {
    await getTemplate('3-stage-clos').scaffold(2);
    const call = mockBlockCreate.mock.calls[0][0] as Record<string, unknown>;
    expect(call.leaf_model_id).toBeUndefined();
    expect(call.spine_model_id).toBeUndefined();
  });

  it('throws on partial failure', async () => {
    mockSiteCreate.mockRejectedValue(new Error('site error'));
    await expect(getTemplate('3-stage-clos').scaffold(2)).rejects.toThrow('site error');
    expect(mockSuperblockCreate).not.toHaveBeenCalled();
  });
});

// ── Pod-based fabric ──────────────────────────────────────────────────────────

describe('pod-based fabric template', () => {
  beforeEach(() => {
    mockSiteCreate.mockResolvedValue({ id: 12, name: 'Site 1' });
    mockSuperblockCreate
      .mockResolvedValueOnce({ id: 22, name: 'Pod A' })
      .mockResolvedValueOnce({ id: 23, name: 'Pod B' });
    mockBlockCreate
      .mockResolvedValueOnce({ block: { id: 40, name: 'Block 1' } })
      .mockResolvedValueOnce({ block: { id: 41, name: 'Block 2' } })
      .mockResolvedValueOnce({ block: { id: 42, name: 'Block 3' } })
      .mockResolvedValueOnce({ block: { id: 43, name: 'Block 4' } });
  });

  it('creates 1 site, 2 superblocks, 4 blocks on success path', async () => {
    await getTemplate('pod-based').scaffold(3);
    expect(mockSiteCreate).toHaveBeenCalledTimes(1);
    expect(mockSuperblockCreate).toHaveBeenCalledTimes(2);
    expect(mockBlockCreate).toHaveBeenCalledTimes(4);
  });

  it('creates site named "Site 1"', async () => {
    await getTemplate('pod-based').scaffold(3);
    expect(mockSiteCreate).toHaveBeenCalledWith(3, expect.objectContaining({ name: 'Site 1' }));
  });

  it('creates superblocks named "Pod A" and "Pod B"', async () => {
    await getTemplate('pod-based').scaffold(3);
    expect(mockSuperblockCreate).toHaveBeenCalledWith(12, expect.objectContaining({ name: 'Pod A' }));
    expect(mockSuperblockCreate).toHaveBeenCalledWith(12, expect.objectContaining({ name: 'Pod B' }));
  });

  it('creates 2 blocks under Pod A and 2 under Pod B', async () => {
    await getTemplate('pod-based').scaffold(3);
    const calls = mockBlockCreate.mock.calls as Array<[Record<string, unknown>, ...unknown[]]>;
    const podACalls = calls.filter((c) => c[0].super_block_id === 22);
    const podBCalls = calls.filter((c) => c[0].super_block_id === 23);
    expect(podACalls).toHaveLength(2);
    expect(podBCalls).toHaveLength(2);
  });

  it('uses named blocks (Block 1–4), not generic names', async () => {
    await getTemplate('pod-based').scaffold(3);
    const calls = mockBlockCreate.mock.calls as Array<[Record<string, unknown>, ...unknown[]]>;
    const names = calls.map((c) => c[0].name);
    expect(names).toEqual(['Block 1', 'Block 2', 'Block 3', 'Block 4']);
  });

  it('does not pre-assign devices on any block', async () => {
    await getTemplate('pod-based').scaffold(3);
    const calls = mockBlockCreate.mock.calls as Array<[Record<string, unknown>, ...unknown[]]>;
    for (const call of calls) {
      expect(call[0].leaf_model_id).toBeUndefined();
      expect(call[0].spine_model_id).toBeUndefined();
    }
  });

  it('throws on partial failure (site error)', async () => {
    mockSiteCreate.mockRejectedValue(new Error('network error'));
    await expect(getTemplate('pod-based').scaffold(3)).rejects.toThrow('network error');
    expect(mockSuperblockCreate).not.toHaveBeenCalled();
    expect(mockBlockCreate).not.toHaveBeenCalled();
  });

  it('throws on partial failure (second superblock error)', async () => {
    // Reset implementations so that the beforeEach queue is replaced
    mockSuperblockCreate.mockReset();
    mockBlockCreate.mockReset();
    mockSuperblockCreate
      .mockResolvedValueOnce({ id: 22, name: 'Pod A' })
      .mockRejectedValueOnce(new Error('pod B error'));
    mockBlockCreate
      .mockResolvedValueOnce({ block: { id: 40, name: 'Block 1' } })
      .mockResolvedValueOnce({ block: { id: 41, name: 'Block 2' } });
    await expect(getTemplate('pod-based').scaffold(3)).rejects.toThrow('pod B error');
  });
});
