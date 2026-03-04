import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright configuration for end-to-end tests.
 *
 * Stack expectations:
 *   - The Go backend must be running on :8080 before tests start.
 *     Locally: start it with `make backend/run` in a separate terminal.
 *     In CI: the e2e.yml workflow starts the binary before invoking Playwright.
 *   - Vite dev server is started automatically via the `webServer` entry below.
 *     Vite's proxy config forwards /api/* → localhost:8080, so the frontend
 *     and backend share one origin from the browser's perspective.
 *
 * workers: 1
 *   Journey tests mutate real database state (create trip, add stop, delete).
 *   Parallel workers would cause tests to interfere with each other's data.
 *   A single worker guarantees sequential, isolated execution.
 *
 * reuseExistingServer:
 *   Locally (`CI` unset): if `make frontend/dev` is already running, reuse it —
 *   no need to start a second Vite process.
 *   In CI (`CI=true`): always start fresh to avoid stale state from a prior run.
 */
export default defineConfig({
  testDir: './e2e',

  // Sequential execution — see note above.
  workers: 1,

  // Retry once on CI to absorb one-off network/timing flakes.
  retries: process.env.CI ? 1 : 0,

  // GitHub-native reporter annotates PR checks; human-readable list locally.
  reporter: process.env.CI ? 'github' : 'list',

  use: {
    baseURL: 'http://localhost:5173',
    // Capture a full execution trace on the first retry so failures can be
    // diagnosed by downloading the artifact from the CI run.
    trace: 'on-first-retry',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Playwright starts the Vite dev server before the first test and shuts it
  // down after the last one.  The dev server proxies /api → :8080, so the
  // backend is the only external dependency tests need to worry about.
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
    stdout: 'pipe',
    stderr: 'pipe',
  },
})
