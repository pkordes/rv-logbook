import { apiFetch } from './client'
import { TagListSchema, TagSchema, type Tag, type TagListResponse } from './schemas'

// Re-export so feature code only needs to import from this file.
export type { Tag, TagListResponse }

// ---------------------------------------------------------------------------
// API functions
// ---------------------------------------------------------------------------

/**
 * Search for tags whose slug starts with `q`.
 *
 * Returns a plain array of matching tags — the pagination wrapper is stripped
 * because the primary consumer is an autocomplete dropdown that only needs the
 * list of candidates.
 */
export function searchTags(q: string): Promise<Tag[]> {
  return apiFetch<unknown>(`/tags?q=${encodeURIComponent(q)}`).then(
    (raw) => TagListSchema.parse(raw).data,
  )
}

/**
 * Fetch a paginated list of all tags, ordered by slug.
 *
 * Used by the Tags management page where the full paginated response is needed.
 */
export function listAllTags(page = 1, limit = 20): Promise<TagListResponse> {
  return apiFetch<unknown>(`/tags?page=${page}&limit=${limit}`).then((raw) =>
    TagListSchema.parse(raw),
  )
}

/**
 * Rename a tag. The slug is the stable identifier and is never changed.
 *
 * Returns the updated tag with its new display name.
 */
export function patchTag(slug: string, name: string): Promise<Tag> {
  return apiFetch<unknown>(`/tags/${encodeURIComponent(slug)}`, {
    method: 'PATCH',
    body: JSON.stringify({ name }),
  }).then((raw) => TagSchema.parse(raw))
}
