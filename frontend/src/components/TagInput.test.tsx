import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import { TagInput } from './TagInput'

// ---------------------------------------------------------------------------
// Mock the tags API so tests never hit the network
// ---------------------------------------------------------------------------

vi.mock('../api/tags', () => ({
  searchTags: vi.fn(),
}))

import { searchTags } from '../api/tags'
const mockSearchTags = vi.mocked(searchTags)

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const TAG_MOUNTAIN = {
  id: '00000000-0000-4000-8000-000000000010',
  name: 'Mountain',
  slug: 'mountain',
  created_at: '2025-06-01T00:00:00Z',
}
const TAG_NATIONAL_PARK = {
  id: '00000000-0000-4000-8000-000000000011',
  name: 'National Park',
  slug: 'national-park',
  created_at: '2025-06-01T00:00:00Z',
}

beforeEach(() => {
  // Default: return no suggestions
  mockSearchTags.mockResolvedValue([])
})

afterEach(() => {
  vi.clearAllMocks()
})

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('TagInput', () => {
  it('renders a text input', () => {
    render(<TagInput value={[]} onChange={vi.fn()} />)

    expect(screen.getByRole('textbox')).toBeInTheDocument()
  })

  it('renders existing value entries as removable pills', () => {
    render(<TagInput value={['Mountain', 'National Park']} onChange={vi.fn()} />)

    expect(screen.getByText('Mountain')).toBeInTheDocument()
    expect(screen.getByText('National Park')).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /remove/i })).toHaveLength(2)
  })

  it('calls onChange without the tag when a pill remove button is clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={['Mountain', 'National Park']} onChange={onChange} />)

    await user.click(screen.getByRole('button', { name: /remove mountain/i }))

    expect(onChange).toHaveBeenCalledWith(['National Park'])
  })

  it('calls onChange with new tag appended when Enter is pressed', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={['Mountain']} onChange={onChange} />)

    await user.type(screen.getByRole('textbox'), 'Camping{Enter}')

    expect(onChange).toHaveBeenCalledWith(['Mountain', 'Camping'])
  })

  it('trims the tag name when adding via Enter', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={[]} onChange={onChange} />)

    await user.type(screen.getByRole('textbox'), '  Hiking  {Enter}')

    expect(onChange).toHaveBeenCalledWith(['Hiking'])
  })

  it('does not call onChange when Enter is pressed with blank input', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={[]} onChange={onChange} />)

    await user.type(screen.getByRole('textbox'), '   {Enter}')

    expect(onChange).not.toHaveBeenCalled()
  })

  it('removes the last tag when Backspace is pressed on an empty input', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={['Mountain', 'National Park']} onChange={onChange} />)

    await user.click(screen.getByRole('textbox'))
    await user.keyboard('{Backspace}')

    expect(onChange).toHaveBeenCalledWith(['Mountain'])
  })

  it('does not call onChange on Backspace when value is empty', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={[]} onChange={onChange} />)

    await user.click(screen.getByRole('textbox'))
    await user.keyboard('{Backspace}')

    expect(onChange).not.toHaveBeenCalled()
  })

  it('calls searchTags when the user types 2 or more characters', async () => {
    const user = userEvent.setup()
    mockSearchTags.mockResolvedValue([TAG_MOUNTAIN])
    render(<TagInput value={[]} onChange={vi.fn()} />)

    await user.type(screen.getByRole('textbox'), 'mo')

    await waitFor(() => expect(mockSearchTags).toHaveBeenCalledWith('mo'))
  })

  it('does not call searchTags when fewer than 2 characters are typed', async () => {
    const user = userEvent.setup()
    render(<TagInput value={[]} onChange={vi.fn()} />)

    await user.type(screen.getByRole('textbox'), 'm')

    expect(mockSearchTags).not.toHaveBeenCalled()
  })

  it('renders suggestions in a dropdown when searchTags resolves', async () => {
    const user = userEvent.setup()
    mockSearchTags.mockResolvedValue([TAG_MOUNTAIN, TAG_NATIONAL_PARK])
    render(<TagInput value={[]} onChange={vi.fn()} />)

    await user.type(screen.getByRole('textbox'), 'mo')

    await waitFor(() =>
      expect(screen.getByRole('option', { name: 'Mountain' })).toBeInTheDocument(),
    )
    expect(screen.getByRole('option', { name: 'National Park' })).toBeInTheDocument()
  })

  it('adds a suggestion to value when clicked and clears the input', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    mockSearchTags.mockResolvedValue([TAG_MOUNTAIN])
    render(<TagInput value={[]} onChange={onChange} />)

    await user.type(screen.getByRole('textbox'), 'mo')
    await waitFor(() =>
      expect(screen.getByRole('option', { name: 'Mountain' })).toBeInTheDocument(),
    )

    await user.click(screen.getByRole('option', { name: 'Mountain' }))

    expect(onChange).toHaveBeenCalledWith(['Mountain'])
    expect(screen.getByRole('textbox')).toHaveValue('')
  })

  it('hides the dropdown after a suggestion is selected', async () => {
    const user = userEvent.setup()
    mockSearchTags.mockResolvedValue([TAG_MOUNTAIN])
    render(<TagInput value={[]} onChange={vi.fn()} />)

    await user.type(screen.getByRole('textbox'), 'mo')
    await waitFor(() =>
      expect(screen.getByRole('option', { name: 'Mountain' })).toBeInTheDocument(),
    )

    await user.click(screen.getByRole('option', { name: 'Mountain' }))

    expect(screen.queryByRole('option', { name: 'Mountain' })).not.toBeInTheDocument()
  })

  it('does not add a duplicate tag', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<TagInput value={['Mountain']} onChange={onChange} />)

    await user.type(screen.getByRole('textbox'), 'Mountain{Enter}')

    expect(onChange).not.toHaveBeenCalled()
  })
})
