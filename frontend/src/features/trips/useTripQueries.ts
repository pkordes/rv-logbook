import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createTrip,
  deleteTrip,
  getTrip,
  listTrips,
  type CreateTripInput,
} from '../../api/trips'

// ---------------------------------------------------------------------------
// Query keys
//
// Centralising keys here means a rename is one change, not a search-and-replace.
// The array shape lets TanStack Query do partial invalidation:
// invalidating ['trips'] also invalidates ['trips', someId].
// ---------------------------------------------------------------------------

export const tripKeys = {
  all: ['trips'] as const,
  list: () => [...tripKeys.all, 'list'] as const,
  detail: (id: string) => [...tripKeys.all, 'detail', id] as const,
}

// ---------------------------------------------------------------------------
// Queries (read)
// ---------------------------------------------------------------------------

/**
 * useTrip fetches a single trip by ID.
 */
export function useTrip(id: string) {
  return useQuery({
    queryKey: tripKeys.detail(id),
    queryFn: () => getTrip(id),
    enabled: Boolean(id),
  })
}

/**
 * useTrips fetches the paginated trip list.
 *
 * Components use this instead of calling listTrips() directly so that
 * caching, loading state, and error state are handled consistently.
 * staleTime is inherited from the QueryClient default (30s).
 */
export function useTrips() {
  return useQuery({
    queryKey: tripKeys.list(),
    queryFn: () => listTrips(),
  })
}

// ---------------------------------------------------------------------------
// Mutations (write)
// ---------------------------------------------------------------------------

/**
 * useCreateTrip returns a mutation object for creating a new trip.
 *
 * On success it invalidates the trip list cache, which triggers a background
 * refetch — the list updates automatically without the component managing it.
 * This is the "invalidate and refetch" pattern — the simplest correct approach.
 */
export function useCreateTrip() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateTripInput) => createTrip(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tripKeys.list() })
    },
  })
}

/**
 * useDeleteTrip returns a mutation object for deleting a trip by ID.
 * Same invalidation pattern as useCreateTrip.
 */
export function useDeleteTrip() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteTrip(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: tripKeys.list() })
    },
  })
}
