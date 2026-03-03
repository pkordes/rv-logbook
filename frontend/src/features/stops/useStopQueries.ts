import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createStop,
  deleteStop,
  listStops,
  updateStop,
  type CreateStopInput,
  type UpdateStopInput,
} from '../../api/stops'

// ---------------------------------------------------------------------------
// Query keys
//
// Stops are always scoped under a trip — the key includes the tripId so that
// caches for different trips never collide, and invalidating one trip's stops
// never affects another's.
//
// ['trips', tripId, 'stops', 'list'] means:
//   - invalidateQueries({ queryKey: ['trips', tripId] }) clears everything
//     for that trip (stops + the trip itself), useful when the trip is deleted.
//   - invalidateQueries({ queryKey: stopKeys.list(tripId) }) only clears stops.
// ---------------------------------------------------------------------------

export const stopKeys = {
  all: (tripId: string) => ['trips', tripId, 'stops'] as const,
  list: (tripId: string) => [...stopKeys.all(tripId), 'list'] as const,
}

// ---------------------------------------------------------------------------
// Queries (read)
// ---------------------------------------------------------------------------

/**
 * useStops fetches all stops for a given trip.
 *
 * The query is disabled when tripId is empty, which can happen
 * briefly during component mount before params are resolved.
 */
export function useStops(tripId: string) {
  return useQuery({
    queryKey: stopKeys.list(tripId),
    queryFn: () => listStops(tripId),
    enabled: Boolean(tripId),
  })
}

// ---------------------------------------------------------------------------
// Mutations (write)
// ---------------------------------------------------------------------------

/**
 * useCreateStop returns a mutation for adding a stop to a trip.
 * On success it invalidates the stop list so the UI reflects the new stop.
 */
export function useCreateStop(tripId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateStopInput) => createStop(tripId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: stopKeys.list(tripId) })
    },
  })
}

/**
 * useDeleteStop returns a mutation for removing a stop from a trip.
 * Accepts the stopId as the mutation argument.
 */
export function useDeleteStop(tripId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (stopId: string) => deleteStop(tripId, stopId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: stopKeys.list(tripId) })
    },
  })
}

/**
 * useUpdateStop returns a mutation for editing an existing stop.
 * The mutation argument is a { stopId, input } object so callers
 * can pass both the target ID and the updated fields together.
 */
export function useUpdateStop(tripId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ stopId, input }: { stopId: string; input: UpdateStopInput }) =>
      updateStop(tripId, stopId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: stopKeys.list(tripId) })
    },
  })
}
