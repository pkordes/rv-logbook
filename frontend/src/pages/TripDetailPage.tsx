import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { StopList } from '../features/stops/StopList'
import { StopForm, type StopFormValues } from '../features/stops/StopForm'
import { TripTimeline } from '../features/trips/TripTimeline'
import { ExportButton } from '../features/trips/ExportButton'
import { useTrip } from '../features/trips/useTripQueries'
import { useStops, useDeleteStop, stopKeys } from '../features/stops/useStopQueries'
import { createStop, updateStop, addTagToStop, removeTagFromStop } from '../api/stops'
import type { Stop } from '../api/stops'
import { ApiError } from '../api/client'
import { Button } from '@/components/ui/button'

/**
 * TripDetailPage owns the /trips/:id route.
 *
 * It reading the trip ID from the URL via useParams — this is the key insight
 * for nested routing: the URL is the single source of truth for "which trip
 * am I looking at". No prop drilling, no global state.
 *
 * The create-stop flow is multi-step:
 *   1. POST /trips/:id/stops  → get the new stop's ID
 *   2. For each tagName: POST /trips/:id/stops/:stopId/tags
 *   3. Invalidate the stop list so the UI refreshes
 * The edit-stop flow:
 *   1. PUT /trips/:id/stops/:stopId  → update core fields
 *   2. Reconcile tags against the stop's current tags:
 *      - Tags newly added in the form → POST .../tags
 *      - Tags removed in the form (present before, absent now) → DELETE .../tags/:slug
 *      - Tags unchanged → no API calls
 *   3. Invalidate the stop list so the UI refreshes
 * We manage both with direct API calls + useState rather than a useMutation
 * chain, which keeps the async logic explicit and easy to follow.
 */
export function TripDetailPage() {
  const { id: tripId = '' } = useParams<{ id: string }>()
  const queryClient = useQueryClient()

  const trip = useTrip(tripId)
  const stops = useStops(tripId)
  const deleteStop = useDeleteStop(tripId)

  const [isAdding, setIsAdding] = useState(false)
  const [addError, setAddError] = useState<string | null>(null)
  const [editingStop, setEditingStop] = useState<Stop | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [editError, setEditError] = useState<string | null>(null)
  const [view, setView] = useState<'list' | 'timeline'>('list')

  async function handleAddStop(values: StopFormValues) {
    setIsAdding(true)
    setAddError(null)
    try {
      const stop = await createStop(tripId, values)
      for (const name of values.tagNames) {
        await addTagToStop(tripId, stop.id, { name })
      }
      await queryClient.invalidateQueries({ queryKey: stopKeys.list(tripId) })
    } catch (e) {
      if (e instanceof ApiError) {
        if (e.status >= 400 && e.status < 500) {
          setAddError('Could not save stop. Please check your entries and try again.')
        } else {
          setAddError('Server error. Please try again in a moment.')
        }
      } else {
        setAddError('Could not reach the server. Is the backend running?')
      }
    } finally {
      setIsAdding(false)
    }
  }

  async function handleEditStop(values: StopFormValues) {
    if (!editingStop) return
    setIsEditing(true)
    setEditError(null)
    try {
      await updateStop(tripId, editingStop.id, values)

      // Reconcile tags: add newly introduced names, remove deleted ones.
      const originalNames = new Set(editingStop.tags.map((t) => t.name))
      const tagsToAdd = values.tagNames.filter((n) => !originalNames.has(n))
      const tagsToRemove = editingStop.tags.filter((t) => !values.tagNames.includes(t.name))

      for (const name of tagsToAdd) {
        await addTagToStop(tripId, editingStop.id, { name })
      }
      for (const tag of tagsToRemove) {
        await removeTagFromStop(tripId, editingStop.id, tag.slug)
      }

      await queryClient.invalidateQueries({ queryKey: stopKeys.list(tripId) })
      setEditingStop(null)
    } catch (e) {
      if (e instanceof ApiError) {
        if (e.status >= 400 && e.status < 500) {
          setEditError('Could not save changes. Please check your entries and try again.')
        } else {
          setEditError('Server error. Please try again in a moment.')
        }
      } else {
        setEditError('Could not reach the server. Is the backend running?')
      }
    } finally {
      setIsEditing(false)
    }
  }

  // ── Trip loading states ──────────────────────────────────────────────────

  if (trip.isLoading) {
    return <LoadingSpinner label="Loading trip..." />
  }

  if (trip.isError || !trip.data) {
    return (
      <p className="text-destructive py-4">
        Failed to load trip. Is the backend running?
      </p>
    )
  }

  const { data: tripData } = trip

  // ── Page ─────────────────────────────────────────────────────────────────

  return (
    <div className="max-w-2xl mx-auto py-8 px-4">
      {/* Breadcrumb */}
      <p className="text-sm text-gray-500 mb-4">
        <Link to="/trips" className="hover:underline">
          ← All Trips
        </Link>
      </p>

      {/* Trip header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold mb-1">{tripData.name}</h1>
          <p className="text-sm text-gray-500">
            Started: {tripData.start_date}
            {tripData.end_date && ` · Ended: ${tripData.end_date}`}
          </p>
        </div>
        <ExportButton />
      </div>

      {/* Stops header + view toggle */}
      <div className="flex items-center justify-between mb-2">
        <h2 className="text-lg font-semibold">Stops</h2>
        <div className="flex gap-1">
          <Button
            type="button"
            variant={view === 'list' ? 'default' : 'outline'}
            size="sm"
            data-testid="view-toggle-list"
            onClick={() => setView('list')}
          >
            List
          </Button>
          <Button
            type="button"
            variant={view === 'timeline' ? 'default' : 'outline'}
            size="sm"
            data-testid="view-toggle-timeline"
            onClick={() => setView('timeline')}
          >
            Timeline
          </Button>
        </div>
      </div>

      {deleteStop.isError && (
        <p role="alert" className="mb-3 text-sm text-destructive">
          Failed to delete stop: {deleteStop.error?.message ?? 'Unknown error'}
        </p>
      )}
      {addError && (
        <p role="alert" className="mb-3 text-sm text-destructive">
          Failed to add stop: {addError}
        </p>
      )}

      {editError && (
        <p role="alert" className="mb-3 text-sm text-destructive">
          Failed to save changes: {editError}
        </p>
      )}

      {stops.isLoading && <LoadingSpinner label="Loading stops..." />}
      {stops.isError && (
        <p className="text-red-600 py-2">Failed to load stops.</p>
      )}
      {!stops.isLoading && !stops.isError && view === 'list' && (
        <StopList
          stops={stops.data?.data ?? []}
          onDelete={(id) => deleteStop.mutate(id)}
          onEdit={(stop) => setEditingStop(stop)}
        />
      )}
      {!stops.isLoading && !stops.isError && view === 'timeline' && (
        <TripTimeline stops={stops.data?.data ?? []} />
      )}

      {editingStop ? (
        <>
          <h2 className="text-lg font-semibold mt-6 mb-2">Edit Stop</h2>
          <StopForm
            key={editingStop.id}
            onSubmit={handleEditStop}
            isSubmitting={isEditing}
            initialValues={editingStop}
            onCancel={() => setEditingStop(null)}
          />
        </>
      ) : (
        <>
          <h2 className="text-lg font-semibold mt-6 mb-2">Add Stop</h2>
          <StopForm key="new" onSubmit={handleAddStop} isSubmitting={isAdding} />
        </>
      )}
    </div>
  )
}
