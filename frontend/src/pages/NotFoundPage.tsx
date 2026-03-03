import { Link } from 'react-router-dom'

/**
 * NotFoundPage is rendered for any URL that doesn't match a defined route.
 * Wired as the catch-all "*" route in main.tsx.
 */
export function NotFoundPage() {
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold text-gray-900">Page not found</h1>
      <p className="mt-2 text-gray-600">
        The page you're looking for doesn't exist.
      </p>
      <Link to="/" className="mt-4 inline-block text-blue-600 hover:underline">
        Go home
      </Link>
    </div>
  )
}
