import type { Stop } from '../../api/stops';

/** Props for {@link StopList}. */
interface StopListProps {
  /** The stops to display, already fetched for a specific trip. */
  stops: Stop[];
  /** Called with the stop `id` when the user clicks Delete on a row. */
  onDelete: (id: string) => void;
}

/**
 * Presentational component that renders a list of stops for a trip.
 *
 * Owns no server state — data and callbacks come from the parent via props.
 * The arrived_at timestamp is displayed as the date portion only (YYYY-MM-DD)
 * — the time component is rarely useful in a logbook list view.
 */
export function StopList({ stops, onDelete }: StopListProps) {
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
              {stop.arrived_at.slice(0, 10)}
            </span>
          </div>
          <button
            type="button"
            aria-label={`Delete ${stop.name}`}
            onClick={() => onDelete(stop.id)}
            className="text-sm text-red-600 hover:text-red-800"
          >
            Delete
          </button>
        </li>
      ))}
    </ul>
  );
}
