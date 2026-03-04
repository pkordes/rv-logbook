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
const createMutate = vi.fn()

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

  it('shows inline confirmation when Delete is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: 'Delete Yellowstone' }))

    expect(screen.getByText(/this will remove it from all stops/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /confirm delete/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /keep/i })).toBeInTheDocument()
    // Original Delete button should no longer be visible for that row
    expect(screen.queryByRole('button', { name: 'Delete Yellowstone' })).not.toBeInTheDocument()
  })

  it('calls deleteTag mutation when Confirm delete is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: 'Delete Yellowstone' }))
    await user.click(screen.getByRole('button', { name: /confirm delete/i }))

    expect(deleteMutate).toHaveBeenCalledWith('yellowstone')
  })

  it('hides confirmation and does not delete when Keep is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: 'Delete Yellowstone' }))
    expect(screen.getByRole('button', { name: /confirm delete/i })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /keep/i }))

    expect(deleteMutate).not.toHaveBeenCalled()
    // Restore the row to normal
    expect(screen.getByRole('button', { name: 'Delete Yellowstone' })).toBeInTheDocument()
  })

  it('shows an inline rename form when Edit is clicked', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: /edit/i }))

    expect(screen.getByRole('textbox', { name: /rename tag/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('calls updateTag mutation with the new name on Save', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: /edit/i }))

    const input = screen.getByRole('textbox', { name: /rename tag/i })
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
    expect(screen.getByRole('textbox', { name: /rename tag/i })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(screen.queryByRole('textbox', { name: /rename tag/i })).not.toBeInTheDocument()
  })

  it('renders the new tag name input', () => {
    renderPage()
    expect(screen.getByRole('textbox', { name: /new tag name/i })).toBeInTheDocument()
  })

  it('calls createTag mutation when the new-tag form is submitted', async () => {
    const user = userEvent.setup()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.spyOn(tagQueries, 'useCreateTag').mockReturnValue({ mutate: createMutate, isPending: false, isError: false } as any)
    renderPage()

    await user.type(screen.getByRole('textbox', { name: /new tag name/i }), 'Waterfall')
    await user.click(screen.getByRole('button', { name: /add tag/i }))

    expect(createMutate).toHaveBeenCalledWith('Waterfall')
  })

  it('clears the new-tag input after submission', async () => {
    const user = userEvent.setup()
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.spyOn(tagQueries, 'useCreateTag').mockReturnValue({ mutate: createMutate, isPending: false, isError: false } as any)
    renderPage()

    await user.type(screen.getByRole('textbox', { name: /new tag name/i }), 'Waterfall')
    await user.click(screen.getByRole('button', { name: /add tag/i }))

    expect(screen.getByRole('textbox', { name: /new tag name/i })).toHaveValue('')
  })
})
