import { LoadingSpinner } from '../components/LoadingSpinner'
import { TripList } from '../features/trips/TripList'
import { useTrips, useDeleteTrip } from '../features/trips/useTripQueries'

/**
 * TripsPage owns the /trips route.
 *
 * It is the "smart" component for the trips feature: it calls TanStack Query
 * hooks for data and mutations, then passes plain props down to the
 * presentational children (TripList, TripForm). No SQL, no fetch calls here.
 */
export function TripsPage() {
  const { data, isLoading, isError } = useTrips()
  const deleteTrip = useDeleteTrip()

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (isError) {
    return (
      <p className="text-red-600 py-4">
        Failed to load trips. Is the backend running?
      </p>
    )
  }

  return (
    <div className="max-w-2xl mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold mb-6">Trips</h1>

      {/* TripForm will be added here in step 10.5 */}

      <TripList
        trips={data?.data ?? []}
        onDelete={(id) => deleteTrip.mutate(id)}
      />
    </div>
  )
}
