import { apiFetch } from './client'

// ---------------------------------------------------------------------------
// Types — mirror the backend domain.Trip JSON shape exactly
// ---------------------------------------------------------------------------

/** Trip as returned by the API. All dates are ISO 8601 strings over the wire. */
export interface Trip {
  id: string
  name: string
  start_date: string
  end_date?: string
  notes?: string
  created_at: string
  updated_at: string
}

/** Paginated list envelope — matches the backend pagination shape from Phase 7. */
export interface TripListResponse {
  data: Trip[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

/** Fields sent when creating a new trip. */
export interface CreateTripInput {
  name: string
  start_date: string
  end_date?: string
  notes?: string
}

/** Fields sent when updating an existing trip. */
export interface UpdateTripInput {
  name: string
  start_date: string
  end_date?: string
  notes?: string
}

// ---------------------------------------------------------------------------
// API functions — one function per HTTP operation
// ---------------------------------------------------------------------------

/** Fetch a paginated list of trips. */
export function listTrips(page = 1, limit = 20): Promise<TripListResponse> {
  return apiFetch<TripListResponse>(`/trips?page=${page}&limit=${limit}`)
}

/** Fetch a single trip by ID. */
export function getTrip(id: string): Promise<Trip> {
  return apiFetch<Trip>(`/trips/${id}`)
}

/** Create a new trip. Returns the created trip with server-assigned fields. */
export function createTrip(input: CreateTripInput): Promise<Trip> {
  return apiFetch<Trip>('/trips', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

/** Replace all mutable fields of an existing trip. */
export function updateTrip(id: string, input: UpdateTripInput): Promise<Trip> {
  return apiFetch<Trip>(`/trips/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

/** Delete a trip by ID. */
export function deleteTrip(id: string): Promise<void> {
  return apiFetch<void>(`/trips/${id}`, { method: 'DELETE' })
}
