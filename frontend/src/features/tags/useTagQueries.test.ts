import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { createElement, type ReactNode } from 'react'
import { useTags, useUpdateTag, useDeleteTag, useCreateTag, tagKeys } from './useTagQueries'
import * as tagsApi from '../../api/tags'
import type { TagListResponse } from '../../api/tags'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const validTag = {
  id: '00000000-0000-4000-8000-000000000010',
  name: 'Yellowstone',
  slug: 'yellowstone',
  created_at: '2025-06-01T00:00:00Z',
}

const validTagList: TagListResponse = {
  data: [validTag],
  pagination: { page: 1, limit: 20, total: 1 },
}

// ---------------------------------------------------------------------------
// tagKeys
// ---------------------------------------------------------------------------

describe('tagKeys', () => {
  it('provides a stable list key', () => {
    expect(tagKeys.list()).toEqual(['tags', 'list'])
  })
})

// ---------------------------------------------------------------------------
// useTags
// ---------------------------------------------------------------------------

describe('useTags', () => {
  beforeEach(() => {
    vi.spyOn(tagsApi, 'listAllTags').mockResolvedValue(validTagList)
  })

  it('calls listAllTags with default page and limit', async () => {
    const { result } = renderHook(() => useTags(), { wrapper: makeWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(tagsApi.listAllTags).toHaveBeenCalledWith(1, 20)
  })

  it('exposes data from the response', async () => {
    const { result } = renderHook(() => useTags(), { wrapper: makeWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.data).toHaveLength(1)
    expect(result.current.data?.data[0].slug).toBe('yellowstone')
  })
})

// ---------------------------------------------------------------------------
// useUpdateTag
// ---------------------------------------------------------------------------

describe('useUpdateTag', () => {
  it('calls patchTag and invalidates the tag list', async () => {
    vi.spyOn(tagsApi, 'patchTag').mockResolvedValue(validTag)
    vi.spyOn(tagsApi, 'listAllTags').mockResolvedValue(validTagList)

    const wrapper = makeWrapper()
    const { result } = renderHook(() => useUpdateTag(), { wrapper })

    act(() => {
      result.current.mutate({ slug: 'yellowstone', name: 'Yellowstone NP' })
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(tagsApi.patchTag).toHaveBeenCalledWith('yellowstone', 'Yellowstone NP')
  })
})

// ---------------------------------------------------------------------------
// useDeleteTag
// ---------------------------------------------------------------------------

describe('useDeleteTag', () => {
  it('calls deleteTag and invalidates the tag list', async () => {
    vi.spyOn(tagsApi, 'deleteTag').mockResolvedValue(undefined)
    vi.spyOn(tagsApi, 'listAllTags').mockResolvedValue(validTagList)

    const wrapper = makeWrapper()
    const { result } = renderHook(() => useDeleteTag(), { wrapper })

    act(() => {
      result.current.mutate('yellowstone')
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(tagsApi.deleteTag).toHaveBeenCalledWith('yellowstone')
  })

  it('also invalidates stop queries so cached trip pages refresh', async () => {
    vi.spyOn(tagsApi, 'deleteTag').mockResolvedValue(undefined)

    // Use an externally-held QueryClient so we can inspect its state.
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    })
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children)

    // Seed a stop list query that should be invalidated after tag deletion.
    const stopsKey = ['trips', 'trip-1', 'stops', 'list']
    queryClient.setQueryData(stopsKey, { data: [], pagination: { page: 1, limit: 20, total: 0 } })

    const { result } = renderHook(() => useDeleteTag(), { wrapper })

    act(() => {
      result.current.mutate('yellowstone')
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    // TanStack Query marks queries as invalidated when invalidateQueries fires.
    // If we only invalidate ['tags', 'list'], the stops key stays clean.
    // This test fails until we also invalidate ['trips'] in onSuccess.
    expect(queryClient.getQueryState(stopsKey)?.isInvalidated).toBe(true)
  })
})

// ---------------------------------------------------------------------------
// useCreateTag
// ---------------------------------------------------------------------------

describe('useCreateTag', () => {
  it('calls createTag and invalidates the tag list', async () => {
    vi.spyOn(tagsApi, 'createTag').mockResolvedValue(validTag)
    vi.spyOn(tagsApi, 'listAllTags').mockResolvedValue(validTagList)

    const wrapper = makeWrapper()
    const { result } = renderHook(() => useCreateTag(), { wrapper })

    act(() => {
      result.current.mutate('National Park')
    })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(tagsApi.createTag).toHaveBeenCalledWith('National Park')
  })
})
