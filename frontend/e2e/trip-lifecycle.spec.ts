import { test, expect } from '@playwright/test'

/**
 * Full trip lifecycle journey.
 *
 * Exercises the complete user workflow end-to-end against a real running stack:
 *
 *   CREATE TRIP → ADD STOP → VIEW TIMELINE → EXPORT CSV → DELETE
 *
 * The test is intentionally written as a single `test()` block rather than
 * multiple independent tests.  Each step depends on the state produced by
 * the previous step (the stop can't be deleted before it's created; the trip
 * can't be exported before the stop exists).  A single block makes the
 * dependency chain explicit and avoids the temptation to share state via
 * `beforeAll`, which hides the ordering assumption.
 *
 * Cleanup:
 *   The final step (delete trip) is the cleanup.  If the test fails mid-way,
 *   data is left in the database.  This is acceptable — CI always starts
 *   from a fresh, migrated database, so leftover rows from a failed run
 *   are gone on the next workflow execution.  Locally, re-running `make e2e`
 *   will create new uniquely-named rows alongside the leftover ones.
 *
 * Selector strategy — per CONTRIBUTING.md § "E2E Testability":
 *   Priority 1: getByRole()    — semantic elements (heading, link, button with aria-label)
 *   Priority 2: getByLabel()   — form fields tied to a <label> via id/htmlFor
 *   Priority 3: getByTestId()  — state-changing buttons and structural toggles
 *   Never:      CSS class names, positional selectors, or getByText() on interactive elements
 */
test.describe('Trip lifecycle', () => {
  // Timestamp suffix makes names unique across parallel runs (e.g. if two
  // developers are running against a shared staging DB) and avoids collisions
  // when a previous test run left data behind after a mid-test failure.
  const tripName = `E2E Trip ${Date.now()}`
  const stopName = `Yellowstone Camp`

  test('create trip → add stop with tag → view timeline → export CSV → delete', async ({
    page,
  }) => {
    // ── 1. Navigate to trips page ────────────────────────────────────────────
    await page.goto('/trips')
    await expect(page.getByRole('heading', { name: 'Trips' })).toBeVisible()

    // ── 2. Create a trip ─────────────────────────────────────────────────────
    await page.getByLabel('Trip Name').fill(tripName)
    await page.getByLabel('Start Date').fill('2025-06-01')
    // See CONTRIBUTING.md § "E2E Testability" for the selector strategy.
    await page.getByTestId('trip-form-submit').click()

    // TanStack Query invalidates and re-fetches after the mutation.
    // Wait for the newly created trip's link to appear in the list.
    const tripLink = page.getByRole('link', { name: tripName })
    await expect(tripLink).toBeVisible()

    // ── 3. Navigate to the trip detail page ──────────────────────────────────
    await tripLink.click()
    await expect(page.getByRole('heading', { name: tripName })).toBeVisible()

    // ── 4. Add a stop ────────────────────────────────────────────────────────
    await page.getByLabel('Stop Name').fill(stopName)
    await page.getByLabel('Arrived At').fill('2025-06-02')
    await page.getByTestId('stop-form-submit').click()

    // Wait for the stop to appear in the list — the Edit button is the most
    // reliable signal because it carries a per-row aria-label.
    await expect(
      page.getByRole('button', { name: `Edit ${stopName}` }),
    ).toBeVisible()

    // ── 5. Switch to timeline view and verify the stop is rendered ────────────
    await page.getByTestId('view-toggle-timeline').click()
    // timeline-stop-name is the only data-testid in the timeline component;
    // .first() handles the case where multiple stops exist on this trip.
    await expect(page.getByTestId('timeline-stop-name').first()).toHaveText(stopName)

    // ── 6. Export CSV ─────────────────────────────────────────────────────────
    // ExportButton creates a Blob, builds an object URL, and programmatically
    // clicks a hidden <a download="..."> element — Playwright intercepts this
    // as a download event even in headless mode.
    const downloadPromise = page.waitForEvent('download')
    await page.getByRole('button', { name: 'Export CSV' }).click()
    const download = await downloadPromise
    expect(download.suggestedFilename()).toBe('rv-logbook-export.csv')

    // ── 7. Switch back to list view and delete the stop ───────────────────────
    await page.getByTestId('view-toggle-list').click()
    await page.getByRole('button', { name: `Delete ${stopName}` }).click()
    await expect(
      page.getByRole('button', { name: `Delete ${stopName}` }),
    ).not.toBeVisible()

    // ── 8. Navigate back to trips list and delete the trip ───────────────────
    // Use the nav landmark rather than the breadcrumb link — the nav is always
    // present so this step works regardless of where on the page we are.
    await page
      .getByRole('navigation', { name: 'Main navigation' })
      .getByRole('link', { name: 'Trips' })
      .click()

    await page.getByRole('button', { name: `Delete ${tripName}` }).click()
    await expect(page.getByRole('link', { name: tripName })).not.toBeVisible()
  })
})
