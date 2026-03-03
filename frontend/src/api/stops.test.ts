import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { listStops, createStop, deleteStop, removeTagFromStop } from './stops'

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

/** Returns a mock for responses with no body (204 No Content). */
function mockNoContent() {
  return vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
}

// Valid v4 UUIDs (Zod v4 enforces the RFC format strictly).
const TRIP_ID = '00000000-0000-4000-8000-000000000001'
const STOP_ID = '00000000-0000-4000-8000-000000000002'

const validStop = {
  id: '00000000-0000-4000-8000-000000000002',
  trip_id: '00000000-0000-4000-8000-000000000001',
  name: 'Yellowstone Camp',
  location: 'Yellowstone, WY',
  arrived_at: '2025-06-02T10:00:00Z',
  departed_at: null,
  notes: null,
  created_at: '2025-06-01T00:00:00Z',
  updated_at: '2025-06-01T00:00:00Z',
}

const validStopList = {
  data: [validStop],
  pagination: { page: 1, limit: 20, total: 1 },
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('listStops', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch(200, validStopList))
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('calls GET /api/trips/:id/stops', async () => {
    const fetchSpy = mockFetch(200, validStopList)
    vi.stubGlobal('fetch', fetchSpy)

    await listStops(TRIP_ID)

    expect(fetchSpy).toHaveBeenCalledWith(
      `/api/trips/${TRIP_ID}/stops?page=1&limit=20`,
      expect.objectContaining({}),
    )
  })

  it('returns a validated StopListResponse', async () => {
    const result = await listStops(TRIP_ID)
    expect(result.data).toHaveLength(1)
    expect(result.data[0].name).toBe('Yellowstone Camp')
    expect(result.pagination.total).toBe(1)
  })
})

describe('createStop', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('calls POST /api/trips/:id/stops', async () => {
    const fetchSpy = mockFetch(201, validStop)
    vi.stubGlobal('fetch', fetchSpy)

    await createStop(TRIP_ID, {
      name: 'Yellowstone Camp',
      arrived_at: '2025-06-02T10:00:00Z',
    })

    expect(fetchSpy).toHaveBeenCalledWith(
      `/api/trips/${TRIP_ID}/stops`,
      expect.objectContaining({ method: 'POST' }),
    )
  })

  it('returns the created stop', async () => {
    vi.stubGlobal('fetch', mockFetch(201, validStop))

    const result = await createStop(TRIP_ID, {
      name: 'Yellowstone Camp',
      arrived_at: '2025-06-02T10:00:00Z',
    })

    expect(result.id).toBe(STOP_ID)
    expect(result.trip_id).toBe(TRIP_ID)
  })
})

describe('deleteStop', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('calls DELETE /api/trips/:tripId/stops/:stopId', async () => {
    const fetchSpy = mockNoContent()
    vi.stubGlobal('fetch', fetchSpy)

    await deleteStop(TRIP_ID, STOP_ID)

    expect(fetchSpy).toHaveBeenCalledWith(
      `/api/trips/${TRIP_ID}/stops/${STOP_ID}`,
      expect.objectContaining({ method: 'DELETE' }),
    )
  })
})

describe('removeTagFromStop', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('calls DELETE /api/trips/:tripId/stops/:stopId/tags/:slug', async () => {
    const fetchSpy = mockNoContent()
    vi.stubGlobal('fetch', fetchSpy)

    await removeTagFromStop(TRIP_ID, STOP_ID, 'mountain')

    expect(fetchSpy).toHaveBeenCalledWith(
      `/api/trips/${TRIP_ID}/stops/${STOP_ID}/tags/mountain`,
      expect.objectContaining({ method: 'DELETE' }),
    )
  })
})
