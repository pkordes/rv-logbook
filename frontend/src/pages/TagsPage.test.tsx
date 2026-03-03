import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { TagsPage } from './TagsPage'
import * as tagQueries from '../features/tags/useTagQueries'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <TagsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// ---------------------------------------------------------------------------
// Shared mock return shapes
// ---------------------------------------------------------------------------

const validTag = {
  id: '00000000-0000-4000-8000-000000000010',
  name: 'Yellowstone',
  slug: 'yellowstone',
  created_at: '2025-06-01T00:00:00Z',
}

const validTagList = {
  data: [validTag],
  pagination: { page: 1, limit: 20, total: 1 },
}

const deleteMutate = vi.fn()
const updateMutate = vi.fn()

function mockHooks() {
  vi.spyOn(tagQueries, 'useTags').mockReturnValue({
    data: validTagList,
    isLoading: false,
    isError: false,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any)

  vi.spyOn(tagQueries, 'useDeleteTag').mockReturnValue({
    mutate: deleteMutate,
    isPending: false,
    isError: false,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any)

  vi.spyOn(tagQueries, 'useUpdateTag').mockReturnValue({
    mutate: updateMutate,
    isPending: false,
    isError: false,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } as any)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('TagsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockHooks()
  })

  it('renders the page heading', () => {
    renderPage()
    expect(screen.getByRole('heading', { name: /tags/i })).toBeInTheDocument()
  })

  it('renders a row for each tag showing name and slug', () => {
    renderPage()
    expect(screen.getByText('Yellowstone')).toBeInTheDocument()
    expect(screen.getByText('yellowstone')).toBeInTheDocument()
  })

  it('shows a loading spinner while data is loading', () => {
    vi.spyOn(tagQueries, 'useTags').mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any)
    renderPage()
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('shows an error message on fetch failure', () => {
    vi.spyOn(tagQueries, 'useTags').mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } as any)
    renderPage()
    expect(screen.getByText(/failed to load tags/i)).toBeInTheDocument()
  })

  it('calls deleteTag mutation when Delete button is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: /delete/i }))

    expect(deleteMutate).toHaveBeenCalledWith('yellowstone')
  })

  it('shows an inline rename form when Edit is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: /edit/i }))

    expect(screen.getByRole('textbox')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('calls updateTag mutation with the new name on Save', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: /edit/i }))

    const input = screen.getByRole('textbox')
    await user.clear(input)
    await user.type(input, 'Yellowstone NP')
    await user.click(screen.getByRole('button', { name: /save/i }))

    await waitFor(() =>
      expect(updateMutate).toHaveBeenCalledWith({ slug: 'yellowstone', name: 'Yellowstone NP' }),
    )
  })

  it('hides the rename form when Cancel is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: /edit/i }))
    expect(screen.getByRole('textbox')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
  })
})
