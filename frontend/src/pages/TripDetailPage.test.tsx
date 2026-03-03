import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { TripDetailPage } from './TripDetailPage';
import * as tripQueries from '../features/trips/useTripQueries';
import * as stopQueries from '../features/stops/useStopQueries';

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

const mockStopList = {
  data: [],
  pagination: { page: 1, limit: 20, total: 0 },
};

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('TripDetailPage', () => {
  beforeEach(() => {
    vi.spyOn(tripQueries, 'useTrip').mockReturnValue({
      data: mockTrip,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof tripQueries.useTrip>);

    vi.spyOn(stopQueries, 'useStops').mockReturnValue({
      data: mockStopList,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof stopQueries.useStops>);

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
    } as ReturnType<typeof tripQueries.useTrip>);

    renderPage();
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('shows an error message when the trip fails to load', () => {
    vi.spyOn(tripQueries, 'useTrip').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as ReturnType<typeof tripQueries.useTrip>);

    renderPage();
    expect(screen.getByText(/failed to load trip/i)).toBeInTheDocument();
  });
});
