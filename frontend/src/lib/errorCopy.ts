import { ApiError } from '@/api/client';

export interface ErrorCopy {
  title: string;
  description: string;
}

/** Generic fallback used when no specific copy is available. */
const GENERIC: ErrorCopy = {
  title: 'Something went wrong',
  description: 'An unexpected error occurred. Please try again.',
};

/** Map of known domain error codes to user-facing copy. */
const CODE_COPY: Record<string, ErrorCopy | ((err: unknown) => ErrorCopy)> = {
  not_found: {
    title: 'Not found',
    description: 'The requested item no longer exists.',
  },
  constraint_violation: (err: unknown) => ({
    title: 'Constraint violation',
    description:
      err instanceof ApiError
        ? err.message
        : 'The request violates a data constraint.',
  }),
  duplicate: {
    title: 'Already exists',
    description: 'An item with this name already exists.',
  },
  seed_read_only: {
    title: 'Read-only',
    description: 'Seed catalog entries cannot be modified.',
  },
  ru_overflow: {
    title: 'Rack full',
    description: "Adding this device would exceed the rack's available rack units.",
  },
  position_overlap: {
    title: 'Position taken',
    description: 'Another device already occupies that position.',
  },
  conflict: {
    title: 'Conflict',
    description: 'This change conflicts with existing data.',
  },
  agg_ports_full: {
    title: 'No available ports',
    description: 'All aggregation ports on this device are allocated.',
  },
  agg_model_downsize: {
    title: 'Model downsize blocked',
    description: 'The new model has fewer ports than are currently in use.',
  },
};

/**
 * Map any thrown value to user-facing `{title, description}` copy.
 *
 * - Known `ApiError` code → specific copy.
 * - Unknown `ApiError` code → server message as description.
 * - Non-`ApiError` (network failure, plain `Error`, etc.) → generic message.
 * - `null` / `undefined` → generic message.
 *
 * This function never throws.
 */
export function errorToCopy(err: unknown): ErrorCopy {
  try {
    if (err instanceof ApiError) {
      const known = CODE_COPY[err.code];
      if (known) {
        return typeof known === 'function' ? known(err) : known;
      }
      // Unknown code — use server message verbatim so it's still useful.
      return {
        title: 'Error',
        description: err.message || GENERIC.description,
      };
    }
    if (err instanceof Error) {
      return GENERIC;
    }
    return GENERIC;
  } catch {
    return GENERIC;
  }
}
