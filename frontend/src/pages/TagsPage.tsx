import { useState } from 'react'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { useTags, useUpdateTag, useDeleteTag, useCreateTag } from '../features/tags/useTagQueries'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

/**
 * TagsPage owns the /tags route.
 *
 * It renders the global tag list with inline rename and delete actions.
 * Each row shows the tag's display name and slug. Clicking "Edit" expands
 * an inline text input so the user can rename without leaving the page.
 *
 * Design decisions:
 * - Inline editing (versus a modal) keeps the interaction lightweight and
 *   matches the stop-list edit pattern already established in TripDetailPage.
 * - Only one row can be in edit mode at a time — opening a new edit row
 *   automatically closes the previous one via the editingSlug state variable.
 */
export function TagsPage() {
  const { data, isLoading, isError } = useTags()
  const updateTag = useUpdateTag()
  const deleteTag = useDeleteTag()
  const createTag = useCreateTag()

  // slug of the row currently being renamed, or null when no row is editing.
  const [editingSlug, setEditingSlug] = useState<string | null>(null)
  // controlled value for the rename input.
  const [draftName, setDraftName] = useState('')
  // slug of the tag awaiting delete confirmation, or null.
  const [pendingDeleteSlug, setPendingDeleteSlug] = useState<string | null>(null)
  // controlled value for the new-tag input in the create form.
  const [newTagName, setNewTagName] = useState('')

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (isError) {
    return (
      <p className="text-destructive py-4">
        Failed to load tags. Is the backend running?
      </p>
    )
  }

  const tags = data?.data ?? []

  function startEdit(slug: string, currentName: string) {
    setEditingSlug(slug)
    setDraftName(currentName)
    setPendingDeleteSlug(null) // close any open confirmation
  }

  function cancelEdit() {
    setEditingSlug(null)
    setDraftName('')
  }

  function saveEdit(slug: string) {
    const trimmed = draftName.trim()
    if (trimmed === '') return
    updateTag.mutate({ slug, name: trimmed })
    setEditingSlug(null)
    setDraftName('')
  }

  function requestDelete(slug: string) {
    setPendingDeleteSlug(slug)
    setEditingSlug(null) // close any open rename form
    setDraftName('')
  }

  function confirmDelete(slug: string) {
    deleteTag.mutate(slug)
    setPendingDeleteSlug(null)
  }

  function cancelDelete() {
    setPendingDeleteSlug(null)
  }

  function handleCreateTag(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = newTagName.trim()
    if (!trimmed) return
    createTag.mutate(trimmed)
    setNewTagName('')
  }

  return (
    <div className="max-w-2xl mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold mb-6">Tags</h1>

      {/* New tag form */}
      <form onSubmit={handleCreateTag} className="flex gap-2 mb-6">
        <Input
          type="text"
          aria-label="New tag name"
          value={newTagName}
          onChange={(e) => setNewTagName(e.target.value)}
          placeholder="New tag name"
          className="flex-1"
        />
        <Button
          type="submit"
          data-testid="tag-form-submit"
          disabled={createTag.isPending || newTagName.trim() === ''}
        >
          Add Tag
        </Button>
      </form>

      {updateTag.isError && (
        <p role="alert" className="mb-4 text-sm text-destructive">
          Failed to rename tag: {updateTag.error?.message ?? 'Unknown error'}
        </p>
      )}
      {deleteTag.isError && (
        <p role="alert" className="mb-4 text-sm text-destructive">
          Failed to delete tag: {deleteTag.error?.message ?? 'Unknown error'}
        </p>
      )}

      {tags.length === 0 ? (
        <p className="text-gray-500">No tags yet. Add tags to stops to see them here.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left border-b border-gray-200">
              <th className="pb-2 font-semibold">Name</th>
              <th className="pb-2 font-semibold">Slug</th>
              <th className="pb-2" />
            </tr>
          </thead>
          <tbody>
            {tags.map((tag) => (
              <tr key={tag.slug} className="border-b border-gray-100 last:border-0">
                <td className="py-2 pr-4">
                  {editingSlug === tag.slug ? (
                    <Input
                      type="text"
                      aria-label="Rename tag"
                      value={draftName}
                      onChange={(e) => setDraftName(e.target.value)}
                      className="h-8 text-sm"
                      autoFocus
                    />
                  ) : (
                    tag.name
                  )}
                </td>
                <td className="py-2 pr-4 text-gray-500">{tag.slug}</td>
                <td className="py-2 whitespace-nowrap">
                  {editingSlug === tag.slug ? (
                    <span className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label="Save tag name"
                        onClick={() => saveEdit(tag.slug)}
                        disabled={updateTag.isPending}
                      >
                        Save
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label="Cancel renaming tag"
                        onClick={cancelEdit}
                      >
                        Cancel
                      </Button>
                    </span>
                  ) : pendingDeleteSlug === tag.slug ? (
                    <span className="flex items-center gap-2">
                      <span className="text-sm text-destructive">
                        This will remove it from all stops.
                      </span>
                      <Button
                        variant="destructive"
                        size="sm"
                        aria-label={`Confirm delete ${tag.name}`}
                        onClick={() => confirmDelete(tag.slug)}
                        disabled={deleteTag.isPending}
                      >
                        Confirm delete
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label={`Keep ${tag.name}`}
                        onClick={cancelDelete}
                      >
                        Keep
                      </Button>
                    </span>
                  ) : (
                    <span className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label={`Edit ${tag.name}`}
                        onClick={() => startEdit(tag.slug, tag.name)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        aria-label={`Delete ${tag.name}`}
                        onClick={() => requestDelete(tag.slug)}
                        disabled={deleteTag.isPending}
                        className="text-destructive hover:text-destructive"
                      >
                        Delete
                      </Button>
                    </span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
