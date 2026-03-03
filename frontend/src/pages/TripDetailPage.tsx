import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { StopList } from '../features/stops/StopList'
import { StopForm, type StopFormValues } from '../features/stops/StopForm'
import { useTrip } from '../features/trips/useTripQueries'
import { useStops, useDeleteStop, stopKeys } from '../features/stops/useStopQueries'
import { createStop, updateStop, addTagToStop } from '../api/stops'
import type { Stop } from '../api/stops'
import { ApiError } from '../api/client'

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
 * The edit-stop flow is the same shape:
 *   1. PUT /trips/:id/stops/:stopId  → update core fields
 *   2. For each tagName typed in the edit form: POST .../tags  (additive only,
 *      existing tags are not removed — full tag management is Phase 12)
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
      // Tags in the edit form are additive: existing tags are not removed.
      // Full tag management (remove, replace) is Phase 12 work.
      for (const name of values.tagNames) {
        await addTagToStop(tripId, editingStop.id, { name })
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
      <p className="text-red-600 py-4">
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
      <h1 className="text-2xl font-bold mb-1">{tripData.name}</h1>
      <p className="text-sm text-gray-500 mb-6">
        Started: {tripData.start_date}
        {tripData.end_date && ` · Ended: ${tripData.end_date}`}
      </p>

      {/* Stops */}
      <h2 className="text-lg font-semibold mb-2">Stops</h2>

      {deleteStop.isError && (
        <p className="mb-3 text-sm text-red-600">
          Failed to delete stop: {deleteStop.error?.message ?? 'Unknown error'}
        </p>
      )}
      {addError && (
        <p className="mb-3 text-sm text-red-600">
          Failed to add stop: {addError}
        </p>
      )}

      {editError && (
        <p className="mb-3 text-sm text-red-600">
          Failed to save changes: {editError}
        </p>
      )}

      {stops.isLoading && <LoadingSpinner label="Loading stops..." />}
      {stops.isError && (
        <p className="text-red-600 py-2">Failed to load stops.</p>
      )}
      {!stops.isLoading && !stops.isError && (
        <StopList
          stops={stops.data?.data ?? []}
          onDelete={(id) => deleteStop.mutate(id)}
          onEdit={(stop) => setEditingStop(stop)}
        />
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
