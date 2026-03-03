/**
 * schemas.ts — Zod runtime validation schemas for all API response types.
 *
 * These schemas serve two purposes at once:
 *
 * 1. Runtime validation — Zod parses the raw API response and throws a
 *    descriptive error if the shape doesn't match (e.g. a required field is
 *    missing, or a field has the wrong type). This catches API contract
 *    violations at the boundary rather than deep inside a component.
 *
 * 2. TypeScript types — `z.infer<typeof TripSchema>` produces a TypeScript
 *    type from the schema. Zod is the single source of truth: one definition
 *    gives you both runtime checking and compile-time safety.
 *
 * Relationship to openapi.d.ts:
 *
 * `openapi.d.ts` is generated from `openapi.yaml` by openapi-typescript and
 * contains the authoritative spec types (e.g. `components['schemas']['Trip']`).
 * The Zod schemas here must stay compatible with those generated types.
 * If the spec changes, run `make frontend/generate`, then update these schemas
 * to match — the TypeScript compiler will flag any mismatch.
 */

import { z } from 'zod'
import type { components } from './openapi.d'

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

export const PaginationSchema = z.object({
  page: z.number().int().positive(),
  limit: z.number().int().positive(),
  total: z.number().int().nonnegative(),
})

// ---------------------------------------------------------------------------
// Tag
// ---------------------------------------------------------------------------

export const TagSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  slug: z.string(),
  created_at: z.string(),
})

/** Tag as returned by the API — type is derived from the Zod schema. */
export type Tag = z.infer<typeof TagSchema>

export const TagListSchema = z.object({
  data: z.array(TagSchema),
  pagination: PaginationSchema,
})

export type TagListResponse = z.infer<typeof TagListSchema>

type _TagCompatCheck = Tag extends components['schemas']['Tag'] ? true : false
const _tagIsCompatible: _TagCompatCheck = true
void _tagIsCompatible

// ---------------------------------------------------------------------------
// Trip
// ---------------------------------------------------------------------------

export const TripSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  start_date: z.string(),
  end_date: z.string().nullable().optional(),
  notes: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
})

/** Trip as returned by the API — type is derived from the Zod schema. */
export type Trip = z.infer<typeof TripSchema>

export const TripListSchema = z.object({
  data: z.array(TripSchema),
  pagination: PaginationSchema,
})

export type TripListResponse = z.infer<typeof TripListSchema>

// ---------------------------------------------------------------------------
// Compile-time compatibility checks
//
// These lines verify that the Zod-inferred types are assignable to the types
// generated from openapi.yaml. If the spec changes and the Zod schemas fall
// out of sync, TypeScript will error here — not silently somewhere downstream.
//
// Note: we check assignability in one direction (Zod → spec). The spec types
// allow `string | null` for nullable fields; Zod infers `string | null | undefined`
// for `.nullable().optional()`, which is a superset. This is intentional —
// the API may omit optional fields entirely.
// ---------------------------------------------------------------------------

type _TripCompatCheck = Trip extends components['schemas']['Trip'] ? true : false
// If this line errors, the TripSchema is missing a required field from the spec.
const _tripIsCompatible: _TripCompatCheck = true
void _tripIsCompatible

// ---------------------------------------------------------------------------
// Stop
// ---------------------------------------------------------------------------

export const StopSchema = z.object({
  id: z.string().uuid(),
  trip_id: z.string().uuid(),
  name: z.string(),
  location: z.string().nullable().optional(),
  arrived_at: z.string(),
  departed_at: z.string().nullable().optional(),
  notes: z.string().nullable().optional(),
  created_at: z.string(),
  updated_at: z.string(),
  // Tags are always present on responses from list/get endpoints.
  // The backend repo guarantees a non-nil slice; the spec declares it optional
  // for write-operation responses (create/update) that don't join tags.
  tags: z.array(TagSchema).optional().default([]),
})

/** Stop as returned by the API — type is derived from the Zod schema. */
export type Stop = z.infer<typeof StopSchema>

export const StopListSchema = z.object({
  data: z.array(StopSchema),
  pagination: PaginationSchema,
})

export type StopListResponse = z.infer<typeof StopListSchema>

type _StopCompatCheck = Stop extends components['schemas']['Stop'] ? true : false
// If this line errors, the StopSchema is missing a required field from the spec.
const _stopIsCompatible: _StopCompatCheck = true
void _stopIsCompatible
