import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TripForm } from './TripForm';

describe('TripForm', () => {
  it('renders name, start date, and a submit button', () => {
    render(<TripForm onSubmit={vi.fn()} isSubmitting={false} />);
    expect(screen.getByLabelText(/trip name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/start date/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /add trip/i })).toBeInTheDocument();
  });

  it('shows a validation error when submitted with an empty name', async () => {
    render(<TripForm onSubmit={vi.fn()} isSubmitting={false} />);
    await userEvent.click(screen.getByRole('button', { name: /add trip/i }));
    expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
  });

  it('shows a validation error when submitted with an empty start date', async () => {
    render(<TripForm onSubmit={vi.fn()} isSubmitting={false} />);
    await userEvent.type(screen.getByLabelText(/trip name/i), 'Pacific Coast');
    await userEvent.click(screen.getByRole('button', { name: /add trip/i }));
    expect(await screen.findByText(/start date is required/i)).toBeInTheDocument();
  });

  it('calls onSubmit with trimmed name and start_date when form is valid', async () => {
    const onSubmit = vi.fn();
    render(<TripForm onSubmit={onSubmit} isSubmitting={false} />);

    await userEvent.type(screen.getByLabelText(/trip name/i), '  Pacific Coast  ');
    await userEvent.type(screen.getByLabelText(/start date/i), '2024-06-01');
    await userEvent.click(screen.getByRole('button', { name: /add trip/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
    expect(onSubmit).toHaveBeenCalledWith({
      name: 'Pacific Coast',
      start_date: '2024-06-01',
      end_date: undefined,
    });
  });

  it('shows a validation error when start date is not in YYYY-MM-DD format', async () => {
    render(<TripForm onSubmit={vi.fn()} isSubmitting={false} />);
    await userEvent.type(screen.getByLabelText(/trip name/i), 'Pacific Coast');
    await userEvent.type(screen.getByLabelText(/start date/i), '06012024');
    await userEvent.click(screen.getByRole('button', { name: /add trip/i }));
    expect(await screen.findByText(/yyyy-mm-dd/i)).toBeInTheDocument();
  });

  it('disables the submit button while isSubmitting is true', () => {
    render(<TripForm onSubmit={vi.fn()} isSubmitting={true} />);
    expect(screen.getByRole('button', { name: /saving/i })).toBeDisabled();
  });
});
