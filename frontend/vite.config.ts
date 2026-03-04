import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    // Run tests in a browser-like environment so window, fetch, etc. exist.
    // 'jsdom' is the standard choice for React component tests.
    environment: 'jsdom',
    // Make describe/it/expect available globally without importing them —
    // mirrors the Jest API that most React tutorials use.
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    // Exclude Playwright E2E specs — they import from @playwright/test, not
    // vitest, and must run via `make e2e` / `npx playwright test` instead.
    exclude: ['e2e/**', '**/node_modules/**'],
  },
  server: {
    proxy: {
      // Forward /api/* to the Go backend during local development.
      // The browser sees everything on localhost:5173, so the Same-Origin
      // Policy never triggers — no CORS headers needed in dev.
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // The backend has no /api prefix — strip it before forwarding.
        // /api/trips → http://localhost:8080/trips
        rewrite: (path: string) => path.replace(/^\/api/, ''),
      },
    },
  },
})
