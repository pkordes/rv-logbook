import { Link } from 'react-router-dom';
import type { Trip } from '../../api/schemas';

/** Props for {@link TripList}. */
interface TripListProps {
  /** The trips to display. */
  trips: Trip[];
  /** Called with the trip `id` when the user clicks Delete on a row. */
  onDelete: (id: string) => void;
}

/**
 * Presentational component that renders a list of trips.
 *
 * This component owns no server state — data and callbacks are passed in as
 * props. Keep business logic and data fetching in the parent (`TripsPage`) or
 * in TanStack Query hooks.
 */
export function TripList({ trips, onDelete }: TripListProps) {
  if (trips.length === 0) {
    return (
      <p className="text-gray-500 italic py-4">
        No trips yet. Add your first trip above.
      </p>
    );
  }

  return (
    <ul className="divide-y divide-gray-200">
      {trips.map((trip) => (
        <li key={trip.id} className="flex items-center justify-between py-3">
          <div>
            <Link
              to={`/trips/${trip.id}`}
              className="font-medium text-gray-900 hover:text-blue-600"
            >
              {trip.name}
            </Link>
            <span className="ml-3 text-sm text-gray-500">{trip.start_date}</span>
          </div>
          <button
            type="button"
            aria-label={`Delete ${trip.name}`}
            onClick={() => onDelete(trip.id)}
            className="text-sm text-red-600 hover:text-red-800"
          >
            Delete
          </button>
        </li>
      ))}
    </ul>
  );
}
