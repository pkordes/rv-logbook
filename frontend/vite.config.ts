import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  test: {
    // Run tests in a browser-like environment so window, fetch, etc. exist.
    // 'jsdom' is the standard choice for React component tests.
    environment: 'jsdom',
    // Make describe/it/expect available globally without importing them —
    // mirrors the Jest API that most React tutorials use.
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
  server: {
    proxy: {
      // Forward /api/* to the Go backend during local development.
      // The browser sees everything on localhost:5173, so the Same-Origin
      // Policy never triggers — no CORS headers needed in dev.
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
