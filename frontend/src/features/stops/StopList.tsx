import type { Stop } from '../../api/stops';
import { TagPill } from '../../components/TagPill';

/** Props for {@link StopList}. */
interface StopListProps {
  /** The stops to display, already fetched for a specific trip. */
  stops: Stop[];
  /** Called with the stop `id` when the user clicks Delete on a row. */
  onDelete: (id: string) => void;
  /** Called with the full stop object when the user clicks Edit on a row. */
  onEdit: (stop: Stop) => void;
}

/**
 * Presentational component that renders a list of stops for a trip.
 *
 * Owns no server state — data and callbacks come from the parent via props.
 * The arrived_at timestamp is displayed as the date portion only (YYYY-MM-DD)
 * — the time component is rarely useful in a logbook list view.
 */
export function StopList({ stops, onDelete, onEdit }: StopListProps) {
  if (stops.length === 0) {
    return (
      <p className="text-gray-500 italic py-4">
        No stops yet. Add your first stop below.
      </p>
    );
  }

  return (
    <ul className="divide-y divide-gray-200">
      {stops.map((stop) => (
        <li key={stop.id} className="flex items-center justify-between py-3">
          <div>
            <span className="font-medium text-gray-900">{stop.name}</span>
            {stop.location && (
              <span className="ml-2 text-sm text-gray-500">{stop.location}</span>
            )}
            <span className="ml-3 text-sm text-gray-400">
              {new Date(stop.arrived_at).toISOString().slice(0, 10)}
            </span>
            {stop.tags.length > 0 && (
              <div className="mt-1 flex flex-wrap gap-1">
                {stop.tags.map((tag) => (
                  <span key={tag.slug} data-testid={`stop-tag-${tag.slug}`}>
                    <TagPill name={tag.name} />
                  </span>
                ))}
              </div>
            )}
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              aria-label={`Edit ${stop.name}`}
              onClick={() => onEdit(stop)}
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              Edit
            </button>
            <button
              type="button"
              aria-label={`Delete ${stop.name}`}
              onClick={() => onDelete(stop.id)}
              className="text-sm text-red-600 hover:text-red-800"
            >
              Delete
            </button>
          </div>
        </li>
      ))}
    </ul>
  );
}
