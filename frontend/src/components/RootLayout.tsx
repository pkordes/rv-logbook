import { NavLink, Outlet } from 'react-router-dom'

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
    <div>
      <nav aria-label="Main navigation">
        <NavLink to="/">Home</NavLink>
        {' | '}
        <NavLink to="/trips">Trips</NavLink>
        {' | '}
        <NavLink to="/tags">Tags</NavLink>
      </nav>
      <main>
        <Outlet />
      </main>
    </div>
  )
}
