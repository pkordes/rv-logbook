import { describe, it, expect, vi, afterEach } from 'vitest'
import { searchTags, listAllTags, patchTag } from './tags'

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
// Fixtures
// ---------------------------------------------------------------------------

const validTag = {
  id: '00000000-0000-4000-8000-000000000010',
  name: 'Yellowstone',
  slug: 'yellowstone',
  created_at: '2025-06-01T00:00:00Z',
}

const validTagList = {
  data: [validTag],
  pagination: { page: 1, limit: 20, total: 1 },
}

// ---------------------------------------------------------------------------
// searchTags
// ---------------------------------------------------------------------------

describe('searchTags', () => {
  afterEach(() => vi.restoreAllMocks())

  it('calls GET /api/tags?q=... with the provided prefix', async () => {
    const fetchSpy = mockFetch(200, validTagList)
    vi.stubGlobal('fetch', fetchSpy)

    await searchTags('yel')

    expect(fetchSpy).toHaveBeenCalledWith(
      '/api/tags?q=yel',
      expect.objectContaining({}),
    )
  })

  it('returns the data array (not the paginated wrapper)', async () => {
    vi.stubGlobal('fetch', mockFetch(200, validTagList))

    const result = await searchTags('yel')

    expect(Array.isArray(result)).toBe(true)
    expect(result).toHaveLength(1)
    expect(result[0].name).toBe('Yellowstone')
  })

  it('returns an empty array when q matches nothing', async () => {
    vi.stubGlobal(
      'fetch',
      mockFetch(200, { data: [], pagination: { page: 1, limit: 20, total: 0 } }),
    )

    const result = await searchTags('zzz')

    expect(result).toEqual([])
  })
})

// ---------------------------------------------------------------------------
// listAllTags
// ---------------------------------------------------------------------------

describe('listAllTags', () => {
  afterEach(() => vi.restoreAllMocks())

  it('calls GET /api/tags with page and limit params', async () => {
    const fetchSpy = mockFetch(200, validTagList)
    vi.stubGlobal('fetch', fetchSpy)

    await listAllTags()

    expect(fetchSpy).toHaveBeenCalledWith(
      '/api/tags?page=1&limit=20',
      expect.objectContaining({}),
    )
  })

  it('returns a validated TagListResponse', async () => {
    vi.stubGlobal('fetch', mockFetch(200, validTagList))

    const result = await listAllTags()

    expect(result.data).toHaveLength(1)
    expect(result.data[0].slug).toBe('yellowstone')
    expect(result.pagination.total).toBe(1)
  })

  it('forwards custom page and limit to the URL', async () => {
    const fetchSpy = mockFetch(200, validTagList)
    vi.stubGlobal('fetch', fetchSpy)

    await listAllTags(2, 50)

    expect(fetchSpy).toHaveBeenCalledWith(
      '/api/tags?page=2&limit=50',
      expect.objectContaining({}),
    )
  })
})

// ---------------------------------------------------------------------------
// patchTag
// ---------------------------------------------------------------------------

describe('patchTag', () => {
  afterEach(() => vi.restoreAllMocks())

  it('calls PATCH /api/tags/:slug with the new name', async () => {
    const fetchSpy = mockFetch(200, validTag)
    vi.stubGlobal('fetch', fetchSpy)

    await patchTag('yellowstone', 'Yellowstone NP')

    expect(fetchSpy).toHaveBeenCalledWith(
      '/api/tags/yellowstone',
      expect.objectContaining({
        method: 'PATCH',
        body: JSON.stringify({ name: 'Yellowstone NP' }),
      }),
    )
  })

  it('returns the updated Tag', async () => {
    vi.stubGlobal('fetch', mockFetch(200, { ...validTag, name: 'Yellowstone NP' }))

    const result = await patchTag('yellowstone', 'Yellowstone NP')

    expect(result.name).toBe('Yellowstone NP')
    expect(result.slug).toBe('yellowstone')
  })
})
