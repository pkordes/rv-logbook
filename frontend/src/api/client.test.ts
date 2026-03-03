import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { apiFetch, ApiError } from './client'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function mockFetch(status: number, body: unknown) {
  return vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  )
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('apiFetch', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch(200, { id: '1', name: 'Test Trip' }))
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('prepends /api to the path', async () => {
    const fetchSpy = mockFetch(200, {})
    vi.stubGlobal('fetch', fetchSpy)

    await apiFetch('/trips')

    expect(fetchSpy).toHaveBeenCalledWith(
      '/api/trips',
      expect.objectContaining({}),
    )
  })

  it('returns parsed JSON on a 2xx response', async () => {
    const result = await apiFetch<{ id: string; name: string }>('/trips')
    expect(result).toEqual({ id: '1', name: 'Test Trip' })
  })

  it('sets Content-Type: application/json by default', async () => {
    const fetchSpy = mockFetch(200, {})
    vi.stubGlobal('fetch', fetchSpy)

    await apiFetch('/trips', { method: 'POST', body: JSON.stringify({}) })

    expect(fetchSpy).toHaveBeenCalledWith(
      '/api/trips',
      expect.objectContaining({
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
        }),
      }),
    )
  })

  it('throws ApiError with the response status on a 4xx response', async () => {
    vi.stubGlobal(
      'fetch',
      mockFetch(404, { error: { code: 'not_found', message: 'not found' } }),
    )

    await expect(apiFetch('/trips/missing')).rejects.toBeInstanceOf(ApiError)
  })

  it('includes the HTTP status code on the thrown ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      mockFetch(422, {
        error: { code: 'validation', message: 'name is required' },
      }),
    )

    try {
      await apiFetch('/trips')
      expect.fail('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).status).toBe(422)
    }
  })

  it('includes the server error message on the thrown ApiError', async () => {
    vi.stubGlobal(
      'fetch',
      mockFetch(422, {
        error: { code: 'validation', message: 'name is required' },
      }),
    )

    try {
      await apiFetch('/trips')
      expect.fail('should have thrown')
    } catch (err) {
      expect((err as ApiError).message).toBe('name is required')
    }
  })

  it('resolves without throwing on a 204 No Content response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(new Response(null, { status: 204 })),
    )

    await expect(apiFetch('/trips/123', { method: 'DELETE' })).resolves.toBeUndefined()
  })
})
