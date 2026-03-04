/** Format options for the data export. */
export type ExportFormat = 'json' | 'csv'

/**
 * Fetches the full data export from the backend and returns it as a Blob.
 *
 * Uses raw `fetch` (not the typed `apiFetch` wrapper) because the response
 * is binary/text data that should be downloaded, not parsed as JSON.
 *
 * @param format - 'json' returns a JSON array; 'csv' returns a flat CSV table
 */
export async function fetchExportBlob(format: ExportFormat): Promise<Blob> {
  const accept = format === 'csv' ? 'text/csv' : 'application/json'
  const query = format === 'csv' ? '?format=csv' : ''
  const url = `/api/export${query}`

  const res = await fetch(url, {
    headers: { Accept: accept },
  })

  if (!res.ok) {
    throw new Error(`Export request failed with status ${res.status}`)
  }

  return res.blob()
}
