import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { TripDetailPage } from './TripDetailPage';
import * as tripQueries from '../features/trips/useTripQueries';
import * as stopQueries from '../features/stops/useStopQueries';
import * as stopsApi from '../api/stops';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const TRIP_ID = '00000000-0000-4000-8000-000000000001';

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/trips/${TRIP_ID}`]}>
        <Routes>
          <Route path="/trips/:id" element={<TripDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

const mockTrip = {
  id: TRIP_ID,
  name: 'Summer Tour 2025',
  start_date: '2025-06-01',
  end_date: null,
  notes: '',
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-01T00:00:00Z',
};

const mockStop = {
  id: '00000000-0000-4000-8000-000000000002',
  trip_id: TRIP_ID,
  name: 'Yellowstone Camp',
  location: null,
  arrived_at: '2025-06-02T00:00:00Z',
  departed_at: null,
  notes: null,
  created_at: '2025-06-01T00:00:00Z',
  updated_at: '2025-06-01T00:00:00Z',
};

const mockStopList = {
  data: [],
  pagination: { page: 1, limit: 20, total: 0 },
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('TripDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(tripQueries, 'useTrip').mockReturnValue({
      data: mockTrip,
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof tripQueries.useTrip>);

    vi.spyOn(stopQueries, 'useStops').mockReturnValue({
      data: mockStopList,
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof stopQueries.useStops>);

    vi.spyOn(stopQueries, 'useDeleteStop').mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
      isError: false,
      error: null,
    } as unknown as ReturnType<typeof stopQueries.useDeleteStop>);
  });

  it('renders the trip name as a heading', () => {
    renderPage();
    expect(screen.getByRole('heading', { name: 'Summer Tour 2025' })).toBeInTheDocument();
  });

  it('renders the trip start date', () => {
    renderPage();
    expect(screen.getByText(/2025-06-01/)).toBeInTheDocument();
  });

  it('renders the empty-state stop message when there are no stops', () => {
    renderPage();
    expect(screen.getByText(/no stops yet/i)).toBeInTheDocument();
  });

  it('renders the add stop form', () => {
    renderPage();
    expect(screen.getByLabelText(/stop name/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add stop/i })).toBeInTheDocument();
  });

  it('shows a loading spinner while the trip is loading', () => {
    vi.spyOn(tripQueries, 'useTrip').mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
    } as unknown as ReturnType<typeof tripQueries.useTrip>);

    renderPage();
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('shows an error message when the trip fails to load', () => {
    vi.spyOn(tripQueries, 'useTrip').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as unknown as ReturnType<typeof tripQueries.useTrip>);

    renderPage();
    expect(screen.getByText(/failed to load trip/i)).toBeInTheDocument();
  });

  it('switches to the edit form when the edit button is clicked', async () => {
    vi.spyOn(stopQueries, 'useStops').mockReturnValue({
      data: { data: [mockStop], pagination: { page: 1, limit: 20, total: 1 } },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof stopQueries.useStops>);

    renderPage();
    await userEvent.click(screen.getByRole('button', { name: /edit yellowstone camp/i }));

    expect(screen.getByRole('button', { name: /save changes/i })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByLabelText(/stop name/i)).toHaveValue('Yellowstone Camp');
    });
  });

  it('returns to the add stop form when Cancel is clicked in edit mode', async () => {
    vi.spyOn(stopQueries, 'useStops').mockReturnValue({
      data: { data: [mockStop], pagination: { page: 1, limit: 20, total: 1 } },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof stopQueries.useStops>);

    renderPage();
    await userEvent.click(screen.getByRole('button', { name: /edit yellowstone camp/i }));
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }));

    expect(screen.getByRole('button', { name: /add stop/i })).toBeInTheDocument();
  });

  // ---------------------------------------------------------------------------
  // API orchestration regression tests
  //
  // These were added after a bug was found via manual testing where addTagToStop
  // was never called on the edit path. Spying on the raw API functions (not just
  // the query hooks) is the right level for testing page-level orchestration.
  // ---------------------------------------------------------------------------

  it('calls createStop then addTagToStop for each tag when the add form is submitted', async () => {
    const createdStop = { ...mockStop, id: '00000000-0000-4000-8000-000000000099' };
    vi.spyOn(stopsApi, 'createStop').mockResolvedValue(createdStop);
    vi.spyOn(stopsApi, 'addTagToStop').mockResolvedValue();

    renderPage();
    await userEvent.type(screen.getByLabelText(/stop name/i), 'Firehole Camp');
    await userEvent.type(screen.getByLabelText(/arrived at/i), '2025-07-01');
    await userEvent.type(screen.getByLabelText(/tags/i), 'camping, hiking');
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));

    await waitFor(() => expect(stopsApi.createStop).toHaveBeenCalledOnce());
    expect(stopsApi.createStop).toHaveBeenCalledWith(
      TRIP_ID,
      expect.objectContaining({ name: 'Firehole Camp', arrived_at: '2025-07-01T00:00:00Z' }),
    );
    expect(stopsApi.addTagToStop).toHaveBeenCalledTimes(2);
    expect(stopsApi.addTagToStop).toHaveBeenCalledWith(TRIP_ID, createdStop.id, { name: 'camping' });
    expect(stopsApi.addTagToStop).toHaveBeenCalledWith(TRIP_ID, createdStop.id, { name: 'hiking' });
  });

  it('calls updateStop then addTagToStop for each tag when the edit form is submitted', async () => {
    vi.spyOn(stopQueries, 'useStops').mockReturnValue({
      data: { data: [mockStop], pagination: { page: 1, limit: 20, total: 1 } },
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof stopQueries.useStops>);
    vi.spyOn(stopsApi, 'updateStop').mockResolvedValue(mockStop);
    vi.spyOn(stopsApi, 'addTagToStop').mockResolvedValue();

    renderPage();
    await userEvent.click(screen.getByRole('button', { name: /edit yellowstone camp/i }));
    await waitFor(() => {
      expect(screen.getByLabelText(/stop name/i)).toHaveValue('Yellowstone Camp');
    });

    await userEvent.type(screen.getByLabelText(/tags/i), 'wildlife');
    await userEvent.click(screen.getByRole('button', { name: /save changes/i }));

    await waitFor(() => expect(stopsApi.updateStop).toHaveBeenCalledOnce());
    expect(stopsApi.updateStop).toHaveBeenCalledWith(
      TRIP_ID,
      mockStop.id,
      expect.objectContaining({ name: 'Yellowstone Camp', arrived_at: '2025-06-02T00:00:00Z' }),
    );
    expect(stopsApi.addTagToStop).toHaveBeenCalledOnce();
    expect(stopsApi.addTagToStop).toHaveBeenCalledWith(TRIP_ID, mockStop.id, { name: 'wildlife' });
  });
});
