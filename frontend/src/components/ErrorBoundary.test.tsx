import { render, screen } from '@testing-library/react'
import { ErrorBoundary } from './ErrorBoundary'

// A component that throws on render — used to trigger the boundary.
function Bomb(): never {
  throw new Error('test error')
}

// React logs the caught error to console.error — suppress that noise in test output.
beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {})
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('ErrorBoundary', () => {
  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <p>hello</p>
      </ErrorBoundary>,
    )

    expect(screen.getByText('hello')).toBeInTheDocument()
  })

  it('renders the fallback UI when a child throws', () => {
    render(
      <ErrorBoundary>
        <Bomb />
      </ErrorBoundary>,
    )

    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('displays the error message in the fallback UI', () => {
    render(
      <ErrorBoundary>
        <Bomb />
      </ErrorBoundary>,
    )

    expect(screen.getByText(/test error/i)).toBeInTheDocument()
  })
})
