import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { TripList } from './TripList';
import type { Trip } from '../../api/schemas';

// TripList uses <Link> internally, so every render needs a router context.
const renderWithRouter = (ui: React.ReactElement) =>
  render(<MemoryRouter>{ui}</MemoryRouter>);

const makeTrip = (overrides: Partial<Trip> = {}): Trip => ({
  id: 'aaaaaaaa-0000-0000-0000-000000000001',
  name: 'Test Trip',
  start_date: '2024-06-01',
  end_date: null,
  notes: '',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  ...overrides,
});

describe('TripList', () => {
  it('shows an empty-state message when there are no trips', () => {
    renderWithRouter(<TripList trips={[]} onDelete={vi.fn()} />);
    expect(screen.getByText(/no trips yet/i)).toBeInTheDocument();
  });

  it('renders each trip name', () => {
    const trips = [
      makeTrip({ id: 'aaaaaaaa-0000-0000-0000-000000000001', name: 'Pacific Coast' }),
      makeTrip({ id: 'aaaaaaaa-0000-0000-0000-000000000002', name: 'Desert Southwest' }),
    ];
    renderWithRouter(<TripList trips={trips} onDelete={vi.fn()} />);
    expect(screen.getByText('Pacific Coast')).toBeInTheDocument();
    expect(screen.getByText('Desert Southwest')).toBeInTheDocument();
  });

  it('renders the start date for each trip', () => {
    const trips = [makeTrip({ start_date: '2024-06-01' })];
    renderWithRouter(<TripList trips={trips} onDelete={vi.fn()} />);
    expect(screen.getByText(/2024-06-01/)).toBeInTheDocument();
  });

  it('calls onDelete with the trip id when the delete button is clicked', async () => {
    const onDelete = vi.fn();
    const trip = makeTrip({ id: 'aaaaaaaa-0000-0000-0000-000000000001', name: 'Road Trip' });
    renderWithRouter(<TripList trips={[trip]} onDelete={onDelete} />);

    await userEvent.click(screen.getByRole('button', { name: /delete road trip/i }));
    expect(onDelete).toHaveBeenCalledOnce();
    expect(onDelete).toHaveBeenCalledWith('aaaaaaaa-0000-0000-0000-000000000001');
  });
});
