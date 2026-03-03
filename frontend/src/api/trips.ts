import { apiFetch } from './client'
import type { components } from './openapi.d'
import {
  TripListSchema,
  TripSchema,
  type Trip,
  type TripListResponse,
} from './schemas'

// ---------------------------------------------------------------------------
// Input types — taken directly from the generated spec types.
// Inputs go *to* the server so there's no need for runtime parsing;
// React Hook Form + Zod validates them before they leave the client.
// ---------------------------------------------------------------------------

/** Fields sent when creating a new trip. */
export type CreateTripInput = components['schemas']['CreateTripRequest']

/** Fields sent when updating an existing trip. */
export type UpdateTripInput = components['schemas']['UpdateTripRequest']

// Re-export response types so feature code only imports from this file.
export type { Trip, TripListResponse }

// ---------------------------------------------------------------------------
// API functions
//
// Each function calls apiFetch<unknown> (we don't know the shape yet) and
// then parses the response through the Zod schema. If the response doesn't
// match the schema, Zod throws a ZodError with a precise field-level message
// rather than allowing bad data to propagate into components silently.
// ---------------------------------------------------------------------------

/** Fetch a paginated list of trips. */
export function listTrips(page = 1, limit = 20): Promise<TripListResponse> {
  return apiFetch<unknown>(`/trips?page=${page}&limit=${limit}`).then(
    (raw) => TripListSchema.parse(raw),
  )
}

/** Fetch a single trip by ID. */
export function getTrip(id: string): Promise<Trip> {
  return apiFetch<unknown>(`/trips/${id}`).then((raw) => TripSchema.parse(raw))
}

/** Create a new trip. Returns the created trip with server-assigned fields. */
export function createTrip(input: CreateTripInput): Promise<Trip> {
  return apiFetch<unknown>('/trips', {
    method: 'POST',
    body: JSON.stringify(input),
  }).then((raw) => TripSchema.parse(raw))
}

/** Replace all mutable fields of an existing trip. */
export function updateTrip(id: string, input: UpdateTripInput): Promise<Trip> {
  return apiFetch<unknown>(`/trips/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }).then((raw) => TripSchema.parse(raw))
}

/** Delete a trip by ID. The server returns 204 No Content. */
export function deleteTrip(id: string): Promise<void> {
  return apiFetch<void>(`/trips/${id}`, { method: 'DELETE' })
}
