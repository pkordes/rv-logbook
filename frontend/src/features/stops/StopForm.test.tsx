import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StopForm } from './StopForm';
import type { Stop } from '../../api/stops';

const makeInitialStop = (overrides: Partial<Stop> = {}): Stop => ({
  id: '00000000-0000-4000-8000-000000000002',
  trip_id: '00000000-0000-4000-8000-000000000001',
  name: 'Yellowstone Camp',
  location: null,
  arrived_at: '2025-06-02T00:00:00Z',
  departed_at: null,
  notes: null,
  created_at: '2025-06-01T00:00:00Z',
  updated_at: '2025-06-01T00:00:00Z',
  ...overrides,
});

describe('StopForm', () => {
  it('renders name, arrived_at, and a submit button', () => {
    render(<StopForm onSubmit={vi.fn()} isSubmitting={false} />);
    expect(screen.getByLabelText(/stop name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/arrived at/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add stop/i })).toBeInTheDocument();
  });

  it('shows a validation error when submitted with an empty name', async () => {
    render(<StopForm onSubmit={vi.fn()} isSubmitting={false} />);
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));
    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
  });

  it('shows a validation error when submitted with an empty arrived_at', async () => {
    render(<StopForm onSubmit={vi.fn()} isSubmitting={false} />);
    await userEvent.type(screen.getByLabelText(/stop name/i), 'Yellowstone Camp');
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));
    expect(await screen.findByText(/arrived at is required/i)).toBeInTheDocument();
  });

  it('shows a validation error when arrived_at is not in YYYY-MM-DD format', async () => {
    // This is the regression test for the Bad Request bug: the form must reject
    // invalid dates itself so garbage never reaches the API.
    render(<StopForm onSubmit={vi.fn()} isSubmitting={false} />);
    await userEvent.type(screen.getByLabelText(/stop name/i), 'Yellowstone Camp');
    await userEvent.type(screen.getByLabelText(/arrived at/i), '06/02/2025'); // wrong format
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));
    expect(await screen.findByText(/yyyy-mm-dd/i)).toBeInTheDocument();
  });

  it('calls onSubmit with trimmed values when form is valid', async () => {
    const onSubmit = vi.fn();
    render(<StopForm onSubmit={onSubmit} isSubmitting={false} />);

    await userEvent.type(screen.getByLabelText(/stop name/i), '  Yellowstone Camp  ');
    await userEvent.type(screen.getByLabelText(/arrived at/i), '2025-06-02');
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Yellowstone Camp',
        arrived_at: '2025-06-02T00:00:00Z',
      }),
    );
  });

  it('parses comma-separated tags into an array', async () => {
    const onSubmit = vi.fn();
    render(<StopForm onSubmit={onSubmit} isSubmitting={false} />);

    await userEvent.type(screen.getByLabelText(/stop name/i), 'Test Stop');
    await userEvent.type(screen.getByLabelText(/arrived at/i), '2025-06-01');
    await userEvent.type(screen.getByLabelText(/tags/i), 'camping, national park , hiking');
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        tagNames: ['camping', 'national park', 'hiking'],
      }),
    );
  });

  it('calls onSubmit with an empty tagNames array when tags field is blank', async () => {
    const onSubmit = vi.fn();
    render(<StopForm onSubmit={onSubmit} isSubmitting={false} />);

    await userEvent.type(screen.getByLabelText(/stop name/i), 'Plain Stop');
    await userEvent.type(screen.getByLabelText(/arrived at/i), '2025-06-01');
    await userEvent.click(screen.getByRole('button', { name: /add stop/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ tagNames: [] }),
    );
  });

  it('disables the submit button while isSubmitting is true', () => {
    render(<StopForm onSubmit={vi.fn()} isSubmitting={true} />);
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled();
  });

  it('pre-fills the form with initialValues when provided', () => {
    const stop = makeInitialStop();
    render(<StopForm onSubmit={vi.fn()} isSubmitting={false} initialValues={stop} />);
    expect(screen.getByLabelText(/stop name/i)).toHaveValue('Yellowstone Camp');
    // arrived_at is stored as RFC 3339; form should display just the date part
    expect(screen.getByLabelText(/arrived at/i)).toHaveValue('2025-06-02');
  });

  it('shows a Save Changes button when initialValues is provided', () => {
    const stop = makeInitialStop();
    render(<StopForm onSubmit={vi.fn()} isSubmitting={false} initialValues={stop} />);
    expect(screen.getByRole('button', { name: /save changes/i })).toBeInTheDocument();
  });

  it('calls onCancel when the cancel button is clicked', async () => {
    const onCancel = vi.fn();
    const stop = makeInitialStop();
    render(
      <StopForm onSubmit={vi.fn()} isSubmitting={false} initialValues={stop} onCancel={onCancel} />,
    );
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onCancel).toHaveBeenCalledOnce();
  });
});
