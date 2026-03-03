/** Props for {@link TagPill}. */
interface TagPillProps {
  /** The display name of the tag. */
  name: string
  /**
   * When provided, renders an × button inside the pill.
   * Called with no arguments when the user clicks remove.
   */
  onRemove?: () => void
}

/**
 * A small badge that displays a tag name.
 *
 * Renders in read-only mode by default. Pass `onRemove` to add an interactive
 * × button — used inside {@link TagInput} where the user can delete pending tags.
 */
export function TagPill({ name, onRemove }: TagPillProps) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-indigo-100 px-2.5 py-0.5 text-xs font-medium text-indigo-800">
      {name}
      {onRemove !== undefined && (
        <button
          type="button"
          aria-label={`Remove ${name}`}
          onClick={onRemove}
          className="ml-0.5 inline-flex h-3.5 w-3.5 flex-shrink-0 items-center justify-center rounded-full text-indigo-600 hover:bg-indigo-200 hover:text-indigo-900 focus:outline-none"
        >
          {/* visually an × symbol, sized to fit inside the pill */}
          <span aria-hidden="true">&times;</span>
        </button>
      )}
    </span>
  )
}
