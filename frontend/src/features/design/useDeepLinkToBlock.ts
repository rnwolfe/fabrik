import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

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
  const buildBlockUrl = useCallback((blockId: number): string => {
    return `/design?block=${blockId}`;
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

  return { buildBlockUrl, goToBlock };
}
