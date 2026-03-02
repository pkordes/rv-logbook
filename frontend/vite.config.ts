import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
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
