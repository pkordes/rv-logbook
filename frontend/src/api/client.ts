/**
 * ApiError is thrown by apiFetch for any non-2xx response.
 *
 * Consumers can catch it and inspect `status` to branch on 404 vs 422, etc.
 * This mirrors how Go service code returns typed sentinel errors (domain.ErrNotFound)
 * that callers switch on — same principle, different syntax.
 */
export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

/**
 * ServerErrorBody is the shape the Go API returns for non-2xx responses.
 * Matches the `{"error": {"code": "...", "message": "..."}}` envelope
 * established in Phase 7.
 */
interface ServerErrorBody {
  error?: { code?: string; message?: string }
}

/**
 * apiFetch is the single entry point for all API calls.
 *
 * It:
 * - Prefixes every path with `/api` (matched by the Vite dev proxy)
 * - Attaches `Content-Type: application/json` on every request
 * - Parses the response as JSON
 * - Throws `ApiError` for any non-2xx response
 *
 * Usage:
 *   const trip = await apiFetch<Trip>('/trips/123')
 *   await apiFetch('/trips', { method: 'POST', body: JSON.stringify(data) })
 */
export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init.headers,
    },
  })

  if (!response.ok) {
    const body: ServerErrorBody = await response.json().catch(() => ({}))
    const code = body.error?.code ?? 'unknown'
    const message = body.error?.message ?? response.statusText
    throw new ApiError(response.status, code, message)
  }

  return response.json() as Promise<T>
}
