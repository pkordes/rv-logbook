import { Skeleton } from '@/components/ui/skeleton'

/**
 * LoadingSpinner renders skeleton placeholder lines while data is fetching.
 *
 * Using Skeleton instead of a spinner gives the user a sense of the layout
 * that is about to appear — a pattern known as "skeleton loading" or
 * "content placeholder". It is preferred over spinners because it reduces
 * perceived wait time.
 *
 * The `label` prop is still accepted (for backward compatibility) but is now
 * used only as the accessible name on the status region.
 *
 * Usage:
 *   <LoadingSpinner />
 *   <LoadingSpinner label="Loading trips..." />
 */
export function LoadingSpinner({ label = 'Loading...' }: { label?: string }) {
  return (
    <div role="status" aria-label={label} className="space-y-3 py-4">
      <Skeleton className="h-5 w-3/4" />
      <Skeleton className="h-5 w-1/2" />
      <Skeleton className="h-5 w-5/6" />
      <Skeleton className="h-5 w-2/3" />
    </div>
  )
}
