import { useEffect, useState } from 'react'
import { searchTags } from '../api/tags'
import type { Tag } from '../api/tags'
import { TagPill } from './TagPill'

/** Props for {@link TagInput}. */
interface TagInputProps {
  /**
   * Current list of pending tag names.
   * Controlled by the parent — typically React Hook Form via `<Controller>`.
   */
  value: string[]
  /** Called whenever the tag list changes. */
  onChange: (tags: string[]) => void
}

/**
 * Controlled tag input with pill display and autocomplete.
 *
 * Behaviour:
 * - Existing tags are shown as removable {@link TagPill} elements.
 * - Typing 2+ characters queries the API for matching tags.
 * - Clicking a suggestion, or pressing Enter, adds the tag.
 * - Backspace on an empty input removes the last tag.
 * - Duplicate tag names (case-sensitive) are silently ignored.
 *
 * Designed to be used inside a React Hook Form `<Controller>` wrapper so the
 * form owns the canonical state.
 */
export function TagInput({ value, onChange }: TagInputProps) {
  const [inputValue, setInputValue] = useState('')
  const [suggestions, setSuggestions] = useState<Tag[]>([])

  // Fetch autocomplete suggestions whenever the input changes.
  // Clearing suggestions when input is too short is handled in the onChange
  // handler to avoid calling setState synchronously inside the effect body.
  useEffect(() => {
    const trimmed = inputValue.trim()
    if (trimmed.length < 2) return

    let cancelled = false
    searchTags(trimmed).then((tags) => {
      if (!cancelled) setSuggestions(tags)
    })
    return () => {
      cancelled = true
    }
  }, [inputValue])

  function addTag(name: string) {
    const trimmed = name.trim()
    if (!trimmed) return
    if (value.includes(trimmed)) return
    onChange([...value, trimmed])
    setInputValue('')
    setSuggestions([])
  }

  function removeTag(name: string) {
    onChange(value.filter((t) => t !== name))
  }

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value
    setInputValue(val)
    // Clear suggestions immediately when the query becomes too short — avoids
    // stale dropdown showing while the user is deleting characters.
    if (val.trim().length < 2) {
      setSuggestions([])
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      // preventDefault stops the browser from submitting the containing form.
      // stopPropagation prevents the event reaching parent keydown handlers —
      // needed because React 17+ delegates events at the root, which can allow
      // the browser's "Enter submits form" behaviour to fire first.
      e.preventDefault()
      e.stopPropagation()
      addTag(inputValue)
    } else if (e.key === 'Tab' && inputValue.trim() !== '') {
      // Commit the typed text as a tag and let Tab move focus normally.
      // Do NOT call preventDefault so the browser still advances focus.
      addTag(inputValue)
    } else if (e.key === 'Backspace' && inputValue === '') {
      if (value.length > 0) {
        onChange(value.slice(0, -1))
      }
    }
  }

  return (
    <div className="flex flex-wrap gap-1 rounded-md border border-gray-300 p-1.5 focus-within:ring-2 focus-within:ring-indigo-500">
      {value.map((tag) => (
        <TagPill key={tag} name={tag} onRemove={() => removeTag(tag)} />
      ))}
      <div className="relative flex-1">
        <input
          type="text"
          aria-label="Add tag"
          value={inputValue}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          className="w-full min-w-[6rem] border-none bg-transparent p-0 text-sm focus:outline-none"
          placeholder={value.length === 0 ? 'Add tags…' : ''}
        />
        {suggestions.length > 0 && (
          <ul
            role="listbox"
            aria-label="Tag suggestions"
            className="absolute left-0 top-full z-10 mt-1 max-h-40 w-48 overflow-auto rounded-md border border-gray-200 bg-white shadow-lg"
          >
            {suggestions.map((tag) => (
              <li
                key={tag.slug}
                role="option"
                aria-selected={false}
                onClick={() => addTag(tag.name)}
                className="cursor-pointer px-3 py-1.5 text-sm hover:bg-indigo-50"
              >
                {tag.name}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
