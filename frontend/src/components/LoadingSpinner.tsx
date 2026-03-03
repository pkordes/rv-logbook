/**
 * LoadingSpinner is a simple presentational component shown while data is
 * fetching. It accepts an optional label for screen reader accessibility.
 *
 * Usage:
 *   <LoadingSpinner />
 *   <LoadingSpinner label="Loading trips..." />
 */
export function LoadingSpinner({ label = 'Loading...' }: { label?: string }) {
  return (
    <div role="status" aria-label={label}>
      <span aria-hidden="true">⏳</span>
    </div>
  )
}
