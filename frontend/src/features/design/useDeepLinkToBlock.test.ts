import { describe, it, expect } from 'vitest';

/**
 * Unit tests for the deep-link-to-block URL building logic.
 *
 * The full hook (useDeepLinkToBlock) wraps react-router's useNavigate and
 * cannot be tested in isolation without a router context. These tests cover
 * the URL-building contract and the block-resolution logic that mirrors the
 * DesignPage useEffect.
 */

// ── Pure URL builder ─────────────────────────────────────────────────────────

function buildBlockUrl(blockId: number): string {
  return `/design?block=${blockId}`;
}

// ── Deep-link resolution (mirrors DesignPage useEffect logic) ────────────────

interface Block { id: number; name: string; }

type DeepLinkResult =
  | { kind: 'found'; blockId: number }
  | { kind: 'not-found'; message: string }
  | { kind: 'malformed' };

function resolveDeepLink(param: string | null, blocks: Block[]): DeepLinkResult {
  if (!param) return { kind: 'not-found', message: 'no param' };

  const id = Number(param);
  if (isNaN(id) || !Number.isInteger(id)) {
    return { kind: 'malformed' };
  }

  const found = blocks.find((b) => b.id === id);
  if (!found) {
    return { kind: 'not-found', message: `Block #${id} not found in this design.` };
  }

  return { kind: 'found', blockId: found.id };
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('buildBlockUrl', () => {
  it('returns the correct deep-link path for a given block ID', () => {
    expect(buildBlockUrl(1)).toBe('/design?block=1');
    expect(buildBlockUrl(42)).toBe('/design?block=42');
    expect(buildBlockUrl(999)).toBe('/design?block=999');
  });
});

describe('resolveDeepLink', () => {
  const blocks: Block[] = [
    { id: 1, name: 'Block A' },
    { id: 2, name: 'Block B' },
    { id: 10, name: 'Block C' },
  ];

  it('happy path: resolves an existing block by ID', () => {
    const result = resolveDeepLink('1', blocks);
    expect(result).toEqual({ kind: 'found', blockId: 1 });
  });

  it('happy path: resolves another existing block', () => {
    const result = resolveDeepLink('10', blocks);
    expect(result).toEqual({ kind: 'found', blockId: 10 });
  });

  it('missing block: returns not-found with descriptive message', () => {
    const result = resolveDeepLink('99', blocks);
    expect(result).toEqual({
      kind: 'not-found',
      message: 'Block #99 not found in this design.',
    });
  });

  it('wrong-design scenario: block exists in another design (ID not in list)', () => {
    // Simulates navigating to a design that does not contain the deep-linked block.
    const result = resolveDeepLink('5', blocks);
    expect(result.kind).toBe('not-found');
    if (result.kind === 'not-found') {
      expect(result.message).toContain('5');
    }
  });

  it('malformed param: non-numeric string returns malformed', () => {
    const result = resolveDeepLink('abc', blocks);
    expect(result).toEqual({ kind: 'malformed' });
  });

  it('malformed param: empty string treated as no param', () => {
    const result = resolveDeepLink('', blocks);
    // Number('') = 0 which is an integer; block 0 doesn't exist → not-found
    expect(result.kind).toBe('not-found');
  });

  it('null param: treated as not-found', () => {
    const result = resolveDeepLink(null, blocks);
    expect(result.kind).toBe('not-found');
  });
});
