import { TagPill } from '../../components/TagPill'
import type { Stop } from '../../api/stops'

/** Props for {@link TripTimeline}. */
interface TripTimelineProps {
  /** The stops to display, in any order. Will be sorted chronologically. */
  stops: Stop[]
}

/**
 * Formats an ISO datetime string as a short human-readable date.
 * Returns "Unknown date" when the value is null or unparseable.
 */
function formatDate(iso: string | null | undefined): string {
  if (!iso) return 'Unknown date'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return 'Unknown date'
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

/**
 * Sorts stops chronologically by arrived_at (ascending).
 * Stops with a null arrived_at are placed at the end.
 */
function sortedByDate(stops: Stop[]): Stop[] {
  return [...stops].sort((a, b) => {
    if (!a.arrived_at && !b.arrived_at) return 0
    if (!a.arrived_at) return 1
    if (!b.arrived_at) return -1
    return new Date(a.arrived_at).getTime() - new Date(b.arrived_at).getTime()
  })
}

/**
 * TripTimeline renders stops in chronological order as a vertical timeline.
 *
 * This is purely a derived view of the same stop data shown in StopList —
 * no new API calls are needed, just reordering and a different layout.
 */
export function TripTimeline({ stops }: TripTimelineProps) {
  if (stops.length === 0) {
    return (
      <p className="py-4 text-sm text-gray-500 italic">
        No stops yet. Add your first stop below.
      </p>
    )
  }

  const sorted = sortedByDate(stops)

  return (
    <ol className="relative border-l border-gray-200 ml-3">
      {sorted.map((stop) => (
        <li key={stop.id} className="mb-8 ml-6">
          {/* Timeline dot */}
          <span className="absolute -left-2.5 mt-1 flex h-5 w-5 items-center justify-center rounded-full bg-indigo-600 ring-4 ring-white" />

          {/* Date */}
          <time className="mb-1 block text-xs font-normal leading-none text-gray-400">
            {formatDate(stop.arrived_at)}
            {stop.departed_at && ` – ${formatDate(stop.departed_at)}`}
          </time>

          {/* Stop name */}
          <p
            data-testid="timeline-stop-name"
            className="text-sm font-semibold text-gray-900"
          >
            {stop.name}
          </p>

          {/* Location */}
          {stop.location && (
            <p className="mt-0.5 text-xs text-gray-500">{stop.location}</p>
          )}

          {/* Tags */}
          {stop.tags.length > 0 && (
            <div className="mt-1.5 flex flex-wrap gap-1">
              {stop.tags.map((tag) => (
                <TagPill key={tag.slug} name={tag.name} />
              ))}
            </div>
          )}
        </li>
      ))}
    </ol>
  )
}
