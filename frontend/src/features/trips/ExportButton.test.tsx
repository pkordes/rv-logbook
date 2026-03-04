import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ExportButton } from './ExportButton'
import * as exportApi from '../../api/export'

beforeEach(() => {
  vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:fake-url')
  vi.spyOn(URL, 'revokeObjectURL').mockReturnValue(undefined)
  vi.spyOn(HTMLAnchorElement.prototype, 'click').mockReturnValue(undefined)
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ExportButton', () => {
  it('renders an Export CSV button', () => {
    render(<ExportButton />)
    expect(screen.getByRole('button', { name: /export csv/i })).toBeInTheDocument()
  })

  it('calls fetchExportBlob with csv format when clicked', async () => {
    vi.spyOn(exportApi, 'fetchExportBlob').mockResolvedValue(
      new Blob(['trip,stop\n'], { type: 'text/csv' }),
    )

    render(<ExportButton />)
    await userEvent.click(screen.getByRole('button', { name: /export csv/i }))

    await waitFor(() => expect(exportApi.fetchExportBlob).toHaveBeenCalledWith('csv'))
  })

  it('creates an object URL from the downloaded blob', async () => {
    const blob = new Blob(['trip,stop\n'], { type: 'text/csv' })
    vi.spyOn(exportApi, 'fetchExportBlob').mockResolvedValue(blob)

    render(<ExportButton />)
    await userEvent.click(screen.getByRole('button', { name: /export csv/i }))

    await waitFor(() => expect(URL.createObjectURL).toHaveBeenCalledWith(blob))
  })

  it('revokes the object URL after triggering the download', async () => {
    vi.spyOn(exportApi, 'fetchExportBlob').mockResolvedValue(
      new Blob(['trip,stop\n'], { type: 'text/csv' }),
    )

    render(<ExportButton />)
    await userEvent.click(screen.getByRole('button', { name: /export csv/i }))

    await waitFor(() => expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:fake-url'))
  })

  it('disables the button while the download is in flight', async () => {
    let resolve!: (b: Blob) => void
    vi.spyOn(exportApi, 'fetchExportBlob').mockReturnValue(
      new Promise<Blob>((res) => { resolve = res }),
    )

    render(<ExportButton />)
    await userEvent.click(screen.getByRole('button', { name: /export csv/i }))

    expect(screen.getByRole('button', { name: /export csv/i })).toBeDisabled()

    resolve(new Blob(['trip,stop\n'], { type: 'text/csv' }))
  })

  it('shows an error message when the download fails', async () => {
    vi.spyOn(exportApi, 'fetchExportBlob').mockRejectedValue(new Error('Network error'))

    render(<ExportButton />)
    await userEvent.click(screen.getByRole('button', { name: /export csv/i }))

    await waitFor(() =>
      expect(screen.getByText(/export failed/i)).toBeInTheDocument(),
    )
  })
})
