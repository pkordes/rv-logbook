import { useState } from 'react'
import { fetchExportBlob } from '../../api/export'
import { Button } from '@/components/ui/button'

/**
 * ExportButton triggers a browser file download of the full trip/stop data.
 *
 * The download mechanism:
 *   1. Fetch the export endpoint → get a Blob
 *   2. Create a temporary object URL from the Blob
 *   3. Programmatically click a hidden <a download="…"> element
 *   4. Revoke the object URL to free memory
 *
 * This is the standard browser pattern for fetch-to-download without a
 * server-side redirect — equivalent to saving a file from a stream in backend code.
 */
export function ExportButton() {
  const [isPending, setIsPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleExport() {
    setIsPending(true)
    setError(null)
    try {
      const blob = await fetchExportBlob('csv')
      const objectUrl = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = objectUrl
      a.download = 'rv-logbook-export.csv'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(objectUrl)
    } catch {
      setError('Export failed. Please try again.')
    } finally {
      setIsPending(false)
    }
  }

  return (
    <div>
      <Button
        type="button"
        variant="outline"
        size="sm"
        aria-label="Export CSV"
        onClick={() => void handleExport()}
        disabled={isPending}
        aria-busy={isPending}
      >
        Export CSV
      </Button>
      {error !== null && (
        <p className="mt-1 text-xs text-destructive" role="alert">
          {error}
        </p>
      )}
    </div>
  )
}
