import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

/**
 * Scrolls to the element matching `window.location.hash` after the page renders.
 * Re-runs whenever the hash or the `contentReady` flag changes, so callers can
 * delay scrolling until async content (e.g. fetched markdown) has mounted.
 *
 * @param contentReady - Pass `true` once the content that contains the target
 *   heading is mounted. Defaults to `true` so callers that don't need it can
 *   omit it.
 */
export function useScrollToHash(contentReady: boolean = true) {
  const { hash } = useLocation();

  useEffect(() => {
    if (!hash || !contentReady) return;

    const id = hash.slice(1); // strip leading '#'
    if (!id) return;

    // Small timeout lets React flush the DOM before we query it.
    const timer = setTimeout(() => {
      const el = document.getElementById(id);
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
    }, 50);

    return () => clearTimeout(timer);
  }, [hash, contentReady]);
}
