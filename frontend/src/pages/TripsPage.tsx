import { LoadingSpinner } from '../components/LoadingSpinner'
import { TripList } from '../features/trips/TripList'
import { TripForm } from '../features/trips/TripForm'
import { useTrips, useCreateTrip, useDeleteTrip } from '../features/trips/useTripQueries'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'sonner'

/**
 * TripsPage owns the /trips route.
 *
 * It is the "smart" component for the trips feature: it calls TanStack Query
 * hooks for data and mutations, then passes plain props down to the
 * presentational children (TripList, TripForm). No SQL, no fetch calls here.
 */

export const TripsPage = () => {
  const { data, isLoading, isError } = useTrips()
  const createTrip = useCreateTrip()
  const deleteTrip = useDeleteTrip()

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (isError) {
    return (
      <p className="text-destructive py-4">
        Failed to load trips. Is the backend running?
      </p>
    )
  }

  return (
    <div className="max-w-xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Trips</h1>

      <Card>
        <CardHeader>
          <CardTitle>New Trip</CardTitle>
        </CardHeader>
        <CardContent>
          <TripForm
            onSubmit={(values) =>
              createTrip.mutate(values, {
                onError: (e) => toast.error(`Failed to create trip: ${e.message ?? 'Unknown error'}`),
              })
            }
            isSubmitting={createTrip.isPending}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Your Trips</CardTitle>
        </CardHeader>
        <CardContent>
          <TripList
            trips={data?.data ?? []}
            onDelete={(id) =>
              deleteTrip.mutate(id, {
                onError: (e) => toast.error(`Failed to delete trip: ${e.message ?? 'Unknown error'}`),
              })
            }
          />
        </CardContent>
      </Card>
    </div>
  )
}
