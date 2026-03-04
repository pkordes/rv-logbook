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

    // ── 4. Add a stop with a tag ────────────────────────────────────────────
    await page.getByLabel('Stop Name').fill(stopName)
    await page.getByLabel('Arrived At').fill('2025-06-02')
    // Type into TagInput (aria-label="Add tag") and press Enter to commit the
    // tag as a pill.  The pill's remove button is the most stable assertion
    // handle — it carries aria-label="Remove camping".
    await page.getByLabel('Add tag').fill('camping')
    await page.getByLabel('Add tag').press('Enter')
    await expect(page.getByRole('button', { name: 'Remove camping' })).toBeVisible()
    await page.getByTestId('stop-form-submit').click()

    // Wait for the stop to appear in the list — the Edit button is the most
    // reliable signal because it carries a per-row aria-label.
    await expect(
      page.getByRole('button', { name: `Edit ${stopName}` }),
    ).toBeVisible()
    // The TagPill is also rendered in the stop row — confirms the tag was
    // persisted and re-fetched correctly by TanStack Query.
    // data-testid="stop-tag-{slug}" is the stable handle for read-only pills
    // in the stop list (no onRemove handler, so no button to query).
    await expect(page.getByTestId('stop-tag-camping')).toBeVisible()

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

/**
 * Tags management journey.
 *
 * Exercises the full CRUD lifecycle on the /tags page:
 *
 *   CREATE TAG → RENAME TAG → DELETE TAG
 *
 * Runs independently of the trip lifecycle test — no shared state between
 * the two describe blocks.  Uses a timestamped name so re-runs against a
 * non-empty database don't collide with tags left from prior runs.
 *
 * This test exercises every per-row identifier added in fix/e2e-testability:
 *   aria-label="Edit {name}"          → open rename form
 *   aria-label="Rename tag"           → rename input (getByLabel)
 *   aria-label="Save tag name"        → confirm rename
 *   aria-label="Delete {newName}"     → open delete confirmation
 *   aria-label="Confirm delete {…}"   → confirm deletion
 *   data-testid="tag-form-submit"     → create form submit
 */
test.describe('Tags management', () => {
  const tagName = `e2e-tag-${Date.now()}`
  const renamedName = `${tagName}-renamed`

  test('create tag → rename tag → delete tag', async ({ page }) => {
    // ── 1. Navigate to tags page ──────────────────────────────────────────────
    await page.goto('/tags')
    await expect(page.getByRole('heading', { name: 'Tags' })).toBeVisible()

    // ── 2. Create a tag ───────────────────────────────────────────────────────
    // See CONTRIBUTING.md § "E2E Testability" for the selector strategy.
    await page.getByLabel('New tag name').fill(tagName)
    await page.getByTestId('tag-form-submit').click()

    // Wait for the new row — the Edit button carries the tag name so its
    // presence proves both that the row rendered AND the name is correct.
    // exact: true prevents partial matching when tagName is a prefix of renamedName.
    await expect(page.getByRole('button', { name: `Edit ${tagName}`, exact: true })).toBeVisible()

    // ── 3. Rename the tag ─────────────────────────────────────────────────────
    await page.getByRole('button', { name: `Edit ${tagName}`, exact: true }).click()
    // Inline rename input replaces the tag name cell.
    const renameInput = page.getByLabel('Rename tag')
    await renameInput.clear()
    await renameInput.fill(renamedName)
    await page.getByRole('button', { name: 'Save tag name' }).click()

    // The row should now show the new name; the old name should be gone.
    // exact: true on both — tagName is a prefix of renamedName so partial
    // matching would cause the old-name assertion to falsely pass.
    await expect(page.getByRole('button', { name: `Edit ${renamedName}`, exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: `Edit ${tagName}`, exact: true })).not.toBeVisible()

    // ── 4. Delete the tag ─────────────────────────────────────────────────────
    await page.getByRole('button', { name: `Delete ${renamedName}` }).click()
    // Inline confirmation row — assert it appeared before clicking to confirm.
    await expect(
      page.getByRole('button', { name: `Confirm delete ${renamedName}` }),
    ).toBeVisible()
    await page.getByRole('button', { name: `Confirm delete ${renamedName}` }).click()

    // The row should be gone.
    await expect(
      page.getByRole('button', { name: `Delete ${renamedName}` }),
    ).not.toBeVisible()
  })
})
