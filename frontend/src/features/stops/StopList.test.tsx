import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StopList } from './StopList';
import type { Stop } from '../../api/stops';

const makeStop = (overrides: Partial<Stop> = {}): Stop => ({
  id: '00000000-0000-4000-8000-000000000002',
  trip_id: '00000000-0000-4000-8000-000000000001',
  name: 'Yellowstone Camp',
  location: 'Yellowstone, WY',
  arrived_at: '2025-06-02T10:00:00Z',
  departed_at: null,
  notes: null,
  created_at: '2025-06-01T00:00:00Z',
  updated_at: '2025-06-01T00:00:00Z',
  tags: [],
  ...overrides,
});

describe('StopList', () => {
  it('shows an empty-state message when there are no stops', () => {
    render(<StopList stops={[]} onDelete={vi.fn()} onEdit={vi.fn()} />);
    expect(screen.getByText(/no stops yet/i)).toBeInTheDocument();
  });

  it('renders each stop name', () => {
    const stops = [
      makeStop({ id: '00000000-0000-4000-8000-000000000002', name: 'Yellowstone Camp' }),
      makeStop({ id: '00000000-0000-4000-8000-000000000003', name: 'Grand Teton Site' }),
    ];
    render(<StopList stops={stops} onDelete={vi.fn()} onEdit={vi.fn()} />);
    expect(screen.getByText('Yellowstone Camp')).toBeInTheDocument();
    expect(screen.getByText('Grand Teton Site')).toBeInTheDocument();
  });

  it('renders the location when present', () => {
    const stops = [makeStop({ location: 'Yellowstone, WY' })];
    render(<StopList stops={stops} onDelete={vi.fn()} onEdit={vi.fn()} />);
    expect(screen.getByText('Yellowstone, WY')).toBeInTheDocument();
  });

  it('renders the arrived_at date for each stop', () => {
    const stops = [makeStop({ arrived_at: '2025-06-02T10:00:00Z' })];
    render(<StopList stops={stops} onDelete={vi.fn()} onEdit={vi.fn()} />);
    expect(screen.getByText(/2025-06-02/)).toBeInTheDocument();
  });

  it('calls onDelete with the stop id when the delete button is clicked', async () => {
    const onDelete = vi.fn();
    const stop = makeStop({
      id: '00000000-0000-4000-8000-000000000002',
      name: 'Yellowstone Camp',
    });
    render(<StopList stops={[stop]} onDelete={onDelete} onEdit={vi.fn()} />);

    await userEvent.click(screen.getByRole('button', { name: /delete yellowstone camp/i }));
    expect(onDelete).toHaveBeenCalledOnce();
    expect(onDelete).toHaveBeenCalledWith('00000000-0000-4000-8000-000000000002');
  });

  it('calls onEdit with the full stop object when the edit button is clicked', async () => {
    const onEdit = vi.fn();
    const stop = makeStop({
      id: '00000000-0000-4000-8000-000000000002',
      name: 'Yellowstone Camp',
    });
    render(<StopList stops={[stop]} onDelete={vi.fn()} onEdit={onEdit} />);

    await userEvent.click(screen.getByRole('button', { name: /edit yellowstone camp/i }));
    expect(onEdit).toHaveBeenCalledOnce();
    expect(onEdit).toHaveBeenCalledWith(stop);
  });
});
