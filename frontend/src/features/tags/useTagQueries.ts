import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { listAllTags, patchTag, deleteTag } from '../../api/tags'

// ---------------------------------------------------------------------------
// Query keys
//
// A flat key structure works here because tags are a global resource — they are
// not scoped under another entity the way stops are scoped under trips.
// ---------------------------------------------------------------------------

export const tagKeys = {
  all: ['tags'] as const,
  list: () => [...tagKeys.all, 'list'] as const,
}

// ---------------------------------------------------------------------------
// Queries (read)
// ---------------------------------------------------------------------------

/**
 * useTags fetches a paginated page of all tags ordered by slug.
 *
 * Uses default page 1 / limit 20 which is sufficient for the management page
 * in Phase 13. Pagination controls can be added as a future enhancement.
 */
export function useTags(page = 1, limit = 20) {
  return useQuery({
    queryKey: tagKeys.list(),
    queryFn: () => listAllTags(page, limit),
  })
}

// ---------------------------------------------------------------------------
// Mutations (write)
// ---------------------------------------------------------------------------

/**
 * useUpdateTag returns a mutation for renaming a tag.
 *
 * Accepts `{ slug, name }` so the call site reads like a named-parameter call,
 * which is clearer than positional arguments in an object spread.
 *
 * On success the tag list is invalidated so the page re-fetches the updated name.
 */
export function useUpdateTag() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ slug, name }: { slug: string; name: string }) =>
      patchTag(slug, name),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tagKeys.list() })
    },
  })
}

/**
 * useDeleteTag returns a mutation for permanently deleting a tag by slug.
 *
 * The server cascades the deletion to all stop_tags rows — the frontend does
 * not need to do extra cleanup. Both the tag list and all trip/stop caches
 * are invalidated on success so any open TripDetailPage reflects the removal.
 */
export function useDeleteTag() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (slug: string) => deleteTag(slug),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tagKeys.list() })
      // Invalidating ['trips'] cascades to all ['trips', tripId, 'stops', 'list']
      // caches, so TripDetailPage will refetch stops (without the deleted tag)
      // the next time it is viewed.
      void queryClient.invalidateQueries({ queryKey: ['trips'] })
    },
  })
}
