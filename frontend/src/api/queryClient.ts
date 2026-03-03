import { QueryClient } from '@tanstack/react-query'

/**
 * queryClient is the single shared QueryClient instance for the application.
 *
 * Configuration decisions:
 *
 * - staleTime: 30s — data fetched within the last 30 seconds is considered fresh
 *   and won't trigger a background refetch when a component mounts. Without this,
 *   TanStack Query refetches on every mount, which is too aggressive for an API
 *   that doesn't change frequently.
 *
 * - retry: 1 — on network error, retry once before showing an error to the user.
 *   The default (3 retries) adds too much latency before the user sees feedback.
 *
 * - refetchOnWindowFocus: false — by default TanStack Query refetches when the
 *   browser tab regains focus. Useful for real-time apps, but for a logbook it
 *   creates unnecessary noise. Can be re-enabled per-query where freshness matters.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30 * 1000, // 30 seconds
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})
