import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import './index.css'
import { RootLayout } from './components/RootLayout'
import { HomePage } from './pages/HomePage'

/**
 * Application router.
 *
 * createBrowserRouter is the recommended API (over <BrowserRouter>) because
 * it supports data loaders, actions, and nested error boundaries — all of
 * which we will use in later phases. The route tree mirrors the URL tree:
 * '/' renders RootLayout; child routes render inside its <Outlet />.
 */
const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    children: [
      { index: true, element: <HomePage /> },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
