import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { StopForm } from './StopForm';

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
});
