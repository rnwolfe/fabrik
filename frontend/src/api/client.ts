const BASE = '/api';

/** A reference to a related resource that caused an error. */
export interface ErrRef {
  kind: string;
  id: number;
  name: string;
}

/**
 * ApiError is thrown by the fetch client when the server returns a non-2xx
 * response. It carries a stable machine-readable `code`, a human-readable
 * `message`, optional `refs` to related resources, and the HTTP `status`.
 */
export class ApiError extends Error {
  readonly code: string;
  readonly refs: ErrRef[];
  readonly status: number;

  constructor(code: string, message: string, refs: ErrRef[], status: number) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.refs = refs;
    this.status = status;
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  });
  if (!res.ok) {
    // Attempt to parse a structured error body from the server.
    let code = 'unknown';
    let message = `${res.status} ${res.statusText}`;
    let refs: ErrRef[] = [];
    try {
      const body = await res.json();
      if (body && typeof body === 'object') {
        if (typeof body.code === 'string' && body.code) {
          code = body.code;
        }
        if (typeof body.message === 'string' && body.message) {
          message = body.message;
        } else if (typeof body.error === 'string' && body.error) {
          // Legacy fallback: servers that only have `error` field.
          message = body.error;
        }
        if (Array.isArray(body.refs)) {
          refs = body.refs as ErrRef[];
        }
      }
    } catch {
      // Body is not JSON — leave defaults as-is.
    }
    throw new ApiError(code, message, refs, res.status);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: <T = void>(path: string) => request<T>(path, { method: 'DELETE' }),
};
