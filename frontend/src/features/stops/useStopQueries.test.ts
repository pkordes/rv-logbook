import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { createElement, type ReactNode } from 'react'
import { useStops, useCreateStop, useDeleteStop, useUpdateStop, stopKeys } from './useStopQueries'
import * as stopsApi from '../../api/stops'
import type { Stop } from '../../api/stops'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Creates a fresh QueryClient for each test so caches never bleed across tests. */
function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const TRIP_ID = '00000000-0000-4000-8000-000000000001'
const STOP_ID = '00000000-0000-4000-8000-000000000002'

const validStop: Stop = {
  id: STOP_ID,
  trip_id: TRIP_ID,
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

describe('stopKeys', () => {
  it('scopes list key under the trip', () => {
    expect(stopKeys.list(TRIP_ID)).toEqual(['trips', TRIP_ID, 'stops', 'list'])
  })
})

describe('useStops', () => {
  beforeEach(() => {
    vi.spyOn(stopsApi, 'listStops').mockResolvedValue(validStopList)
  })

  it('calls listStops with the tripId', async () => {
    const { result } = renderHook(() => useStops(TRIP_ID), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(stopsApi.listStops).toHaveBeenCalledWith(TRIP_ID)
  })

  it('exposes the stops in the data field', async () => {
    const { result } = renderHook(() => useStops(TRIP_ID), {
      wrapper: makeWrapper(),
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.data).toHaveLength(1)
    expect(result.current.data?.data[0].name).toBe('Yellowstone Camp')
  })
})

describe('useCreateStop', () => {
  it('calls createStop and invalidates the stop list', async () => {
    vi.spyOn(stopsApi, 'createStop').mockResolvedValue(validStop)
    vi.spyOn(stopsApi, 'listStops').mockResolvedValue(validStopList)

    const wrapper = makeWrapper()
    const { result } = renderHook(() => useCreateStop(TRIP_ID), { wrapper })

    result.current.mutate({ name: 'New Stop', arrived_at: '2025-07-01T10:00:00Z' })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(stopsApi.createStop).toHaveBeenCalledWith(TRIP_ID, {
      name: 'New Stop',
      arrived_at: '2025-07-01T10:00:00Z',
    })
  })
})

describe('useDeleteStop', () => {
  it('calls deleteStop with both ids', async () => {
    vi.spyOn(stopsApi, 'deleteStop').mockResolvedValue()
    vi.spyOn(stopsApi, 'listStops').mockResolvedValue(validStopList)

    const wrapper = makeWrapper()
    const { result } = renderHook(() => useDeleteStop(TRIP_ID), { wrapper })

    result.current.mutate(STOP_ID)

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(stopsApi.deleteStop).toHaveBeenCalledWith(TRIP_ID, STOP_ID)
  })
})

describe('useUpdateStop', () => {
  it('calls updateStop with tripId, stopId, and input', async () => {
    const updated = { ...validStop, name: 'Updated Camp' }
    vi.spyOn(stopsApi, 'updateStop').mockResolvedValue(updated)
    vi.spyOn(stopsApi, 'listStops').mockResolvedValue(validStopList)

    const wrapper = makeWrapper()
    const { result } = renderHook(() => useUpdateStop(TRIP_ID), { wrapper })

    result.current.mutate({
      stopId: STOP_ID,
      input: { name: 'Updated Camp', arrived_at: '2025-06-02T00:00:00Z' },
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(stopsApi.updateStop).toHaveBeenCalledWith(TRIP_ID, STOP_ID, {
      name: 'Updated Camp',
      arrived_at: '2025-06-02T00:00:00Z',
    })
  })
})
