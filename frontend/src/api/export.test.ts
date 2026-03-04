import { describe, it, expect, vi, afterEach } from 'vitest'
import { fetchExportBlob } from './export'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('fetchExportBlob', () => {
  it('calls GET /export with Accept: application/json by default', async () => {
    const blob = new Blob(['[{}]'], { type: 'application/json' })
    // Use a plain mock object — new Response(blob, ...) triggers blob.stream()
    // which is not implemented in jsdom.
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: true, blob: () => Promise.resolve(blob) }),
    )

    const result = await fetchExportBlob('json')

    const fetchMock = vi.mocked(fetch)
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toMatch(/\/api\/export/)
    expect((init.headers as Record<string, string>)['Accept']).toBe('application/json')
    expect(result).toBeInstanceOf(Blob)
  })

  it('calls GET /export?format=csv with Accept: text/csv when format is csv', async () => {
    const blob = new Blob(['trip,stop\n'], { type: 'text/csv' })
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: true, blob: () => Promise.resolve(blob) }),
    )

    await fetchExportBlob('csv')

    const fetchMock = vi.mocked(fetch)
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toMatch(/format=csv/)
    expect((init.headers as Record<string, string>)['Accept']).toBe('text/csv')
  })

  it('throws when the server responds with a non-2xx status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({ ok: false, status: 500 }),
    )

    await expect(fetchExportBlob('json')).rejects.toThrow()
  })
})
