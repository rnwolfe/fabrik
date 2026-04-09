import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

/**
 * Build the deep-link URL for a block. Exported as a standalone pure function so
 * it can be used outside of React hook contexts (e.g. tests, non-component code).
 *
 * URL format: /design?block=<blockId>
 */
export function buildBlockUrl(blockId: number): string {
  return `/design?block=${blockId}`;
}

/**
 * Returns a helper function that builds deep-link URLs for navigating to a specific
 * block on the design page, and a navigate function to go there immediately.
 *
 * URL format: /design?block=<blockId>
 */
export function useDeepLinkToBlock() {
  const navigate = useNavigate();

  /**
   * Build the deep-link URL for a block.
   */
  const buildBlockUrlMemo = useCallback((blockId: number): string => {
    return buildBlockUrl(blockId);
  }, []);

  /**
   * Navigate to the design page with the given block pre-selected.
   */
  const goToBlock = useCallback(
    (blockId: number) => {
      navigate(`/design?block=${blockId}`);
    },
    [navigate]
  );

  return { buildBlockUrl: buildBlockUrlMemo, goToBlock };
}
