import { apiFetch } from './client'
import type { components } from './openapi.d'
import {
  StopListSchema,
  StopSchema,
  type Stop,
  type StopListResponse,
} from './schemas'

// ---------------------------------------------------------------------------
// Input types — taken directly from the generated spec types.
// Inputs go *to* the server so there's no need for runtime parsing;
// React Hook Form + Zod validates them before they leave the client.
// ---------------------------------------------------------------------------

/** Fields sent when creating a new stop. */
export type CreateStopInput = components['schemas']['CreateStopRequest']

/** Fields sent when replacing an existing stop. */
export type UpdateStopInput = components['schemas']['UpdateStopRequest']

// Re-export response types so feature code only imports from this file.
export type { Stop, StopListResponse }

// ---------------------------------------------------------------------------
// API functions
// ---------------------------------------------------------------------------

/** Fetch a paginated list of stops for a trip. */
export function listStops(
  tripId: string,
  page = 1,
  limit = 20,
): Promise<StopListResponse> {
  return apiFetch<unknown>(
    `/trips/${tripId}/stops?page=${page}&limit=${limit}`,
  ).then((raw) => StopListSchema.parse(raw))
}

/** Fetch a single stop by ID. */
export function getStop(tripId: string, stopId: string): Promise<Stop> {
  return apiFetch<unknown>(`/trips/${tripId}/stops/${stopId}`).then((raw) =>
    StopSchema.parse(raw),
  )
}

/** Create a new stop on a trip. Returns the created stop with server-assigned fields. */
export function createStop(
  tripId: string,
  input: CreateStopInput,
): Promise<Stop> {
  return apiFetch<unknown>(`/trips/${tripId}/stops`, {
    method: 'POST',
    body: JSON.stringify(input),
  }).then((raw) => StopSchema.parse(raw))
}

/** Replace all mutable fields of an existing stop. */
export function updateStop(
  tripId: string,
  stopId: string,
  input: UpdateStopInput,
): Promise<Stop> {
  return apiFetch<unknown>(`/trips/${tripId}/stops/${stopId}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }).then((raw) => StopSchema.parse(raw))
}

/** Delete a stop by ID. The server returns 204 No Content. */
export function deleteStop(tripId: string, stopId: string): Promise<void> {
  return apiFetch<void>(`/trips/${tripId}/stops/${stopId}`, {
    method: 'DELETE',
  })
}

/** Input for adding a tag to a stop. */
export type AddTagInput = components['schemas']['AddTagRequest']

/**
 * Add a tag to a stop by name. The server creates the tag if it does not
 * exist, or reuses an existing one with the same slug.
 * Returns void — callers invalidate the stop list themselves.
 */
export function addTagToStop(
  tripId: string,
  stopId: string,
  input: AddTagInput,
): Promise<void> {
  return apiFetch<void>(`/trips/${tripId}/stops/${stopId}/tags`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

/**
 * Remove a tag from a stop by slug. The tag record itself is not deleted —
 * only the association between this stop and the tag is removed.
 * Returns void — callers invalidate the stop list themselves.
 */
export function removeTagFromStop(
  tripId: string,
  stopId: string,
  slug: string,
): Promise<void> {
  return apiFetch<void>(`/trips/${tripId}/stops/${stopId}/tags/${encodeURIComponent(slug)}`, {
    method: 'DELETE',
  })
}
