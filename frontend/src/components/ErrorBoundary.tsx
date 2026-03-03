import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

/**
 * ErrorBoundary catches uncaught errors thrown during the render of any
 * child component and displays a fallback UI instead of crashing the page.
 *
 * This must be a class component — React's error boundary API (getDerivedStateFromError,
 * componentDidCatch) is only available on class components. There is no hook equivalent.
 *
 * Wrap it around any subtree where you want isolated error containment:
 *
 *   <ErrorBoundary>
 *     <TripList />
 *   </ErrorBoundary>
 *
 * If TripList throws, the error boundary catches it and renders the fallback.
 * The rest of the page is unaffected.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    // Called when a child throws. Return the new state to trigger a re-render
    // with the fallback UI. This is a static method — no access to `this`.
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // Called after the fallback UI is rendered. Use this for error reporting
    // (e.g. Sentry). For now, log to console.
    console.error('ErrorBoundary caught:', error, info)
  }

  render() {
    if (this.state.error) {
      return (
        <div role="alert">
          <p>Something went wrong.</p>
          <p>{this.state.error.message}</p>
        </div>
      )
    }

    return this.props.children
  }
}
