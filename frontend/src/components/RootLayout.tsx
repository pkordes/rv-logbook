import { NavLink, Outlet } from 'react-router-dom'
import { cn } from '@/lib/utils'

/**
 * RootLayout is the persistent shell rendered for every route.
 *
 * It contains the top navigation bar and an <Outlet /> where the matched
 * child page renders. Think of it like a Go chi sub-router that wraps every
 * handler with common middleware — except here we get shared HTML chrome
 * instead of middleware logic.
 */
export function RootLayout() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b bg-primary text-primary-foreground shadow-sm">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
          {/* App identity — links to home */}
          <NavLink
            to="/"
            className="text-lg font-semibold tracking-tight hover:opacity-80 transition-opacity"
          >
            🚐 RV Logbook
          </NavLink>

          {/* Primary navigation links */}
          <nav aria-label="Main navigation" className="flex items-center gap-1">
            {[
              { to: '/trips', label: 'Trips' },
              { to: '/tags', label: 'Tags' },
            ].map(({ to, label }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  cn(
                    'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                    'hover:bg-primary-foreground/10',
                    isActive
                      ? 'bg-primary-foreground/20 underline underline-offset-4'
                      : 'opacity-80',
                  )
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>

      <main className="mx-auto max-w-5xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
