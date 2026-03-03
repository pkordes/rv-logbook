import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TagPill } from './TagPill'

describe('TagPill', () => {
  it('renders the tag name', () => {
    render(<TagPill name="Yellowstone" />)

    expect(screen.getByText('Yellowstone')).toBeInTheDocument()
  })

  it('does not render a remove button when onRemove is not provided', () => {
    render(<TagPill name="Yellowstone" />)

    expect(
      screen.queryByRole('button', { name: /remove yellowstone/i }),
    ).not.toBeInTheDocument()
  })

  it('renders a remove button when onRemove is provided', () => {
    render(<TagPill name="Yellowstone" onRemove={() => {}} />)

    expect(
      screen.getByRole('button', { name: /remove yellowstone/i }),
    ).toBeInTheDocument()
  })

  it('calls onRemove when the remove button is clicked', async () => {
    const user = userEvent.setup()
    const onRemove = vi.fn()
    render(<TagPill name="Yellowstone" onRemove={onRemove} />)

    await user.click(screen.getByRole('button', { name: /remove yellowstone/i }))

    expect(onRemove).toHaveBeenCalledTimes(1)
  })
})
