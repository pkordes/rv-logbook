import { useState } from 'react'
import { LoadingSpinner } from '../components/LoadingSpinner'
import { useTags, useUpdateTag, useDeleteTag } from '../features/tags/useTagQueries'

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

  // slug of the row currently being renamed, or null when no row is editing.
  const [editingSlug, setEditingSlug] = useState<string | null>(null)
  // controlled value for the rename input.
  const [draftName, setDraftName] = useState('')

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (isError) {
    return (
      <p className="text-red-600 py-4">
        Failed to load tags. Is the backend running?
      </p>
    )
  }

  const tags = data?.data ?? []

  function startEdit(slug: string, currentName: string) {
    setEditingSlug(slug)
    setDraftName(currentName)
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

  return (
    <div className="max-w-2xl mx-auto py-8 px-4">
      <h1 className="text-2xl font-bold mb-6">Tags</h1>

      {updateTag.isError && (
        <p className="mb-4 text-sm text-red-600">
          Failed to rename tag: {updateTag.error?.message ?? 'Unknown error'}
        </p>
      )}
      {deleteTag.isError && (
        <p className="mb-4 text-sm text-red-600">
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
                    <input
                      type="text"
                      value={draftName}
                      onChange={(e) => setDraftName(e.target.value)}
                      className="border border-gray-300 rounded px-2 py-1 text-sm w-full"
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
                      <button
                        onClick={() => saveEdit(tag.slug)}
                        disabled={updateTag.isPending}
                        className="text-sm text-blue-600 hover:underline"
                      >
                        Save
                      </button>
                      <button
                        onClick={cancelEdit}
                        className="text-sm text-gray-500 hover:underline"
                      >
                        Cancel
                      </button>
                    </span>
                  ) : (
                    <span className="flex gap-2">
                      <button
                        onClick={() => startEdit(tag.slug, tag.name)}
                        className="text-sm text-blue-600 hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => deleteTag.mutate(tag.slug)}
                        disabled={deleteTag.isPending}
                        className="text-sm text-red-600 hover:underline"
                      >
                        Delete
                      </button>
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
