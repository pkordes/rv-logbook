import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { TripTimeline } from './TripTimeline'
import type { Stop } from '../../api/stops'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeStop(id: string, name: string, arrivedAt: string): Stop {
  return {
    id,
    trip_id: 'trip-1',
    name,
    location: null,
    arrived_at: arrivedAt,
    departed_at: null,
    notes: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
    tags: [],
  }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('TripTimeline', () => {
  it('renders a timeline entry for each stop', () => {
    const stops = [makeStop('1', 'Grand Canyon', '2025-06-10T00:00:00Z')]
    render(<TripTimeline stops={stops} />)
    expect(screen.getByText('Grand Canyon')).toBeInTheDocument()
  })

  it('renders stops in chronological order regardless of input order', () => {
    const stops = [
      makeStop('2', 'Zion', '2025-06-15T00:00:00Z'),
      makeStop('1', 'Grand Canyon', '2025-06-10T00:00:00Z'),
      makeStop('3', 'Bryce Canyon', '2025-06-20T00:00:00Z'),
    ]
    render(<TripTimeline stops={stops} />)
    const entries = screen.getAllByTestId('timeline-stop-name').map((el) => el.textContent)
    expect(entries).toEqual(['Grand Canyon', 'Zion', 'Bryce Canyon'])
  })

  it('renders an empty state when there are no stops', () => {
    render(<TripTimeline stops={[]} />)
    expect(screen.getByText(/no stops yet/i)).toBeInTheDocument()
  })

  it('renders stop location when provided', () => {
    const stop = { ...makeStop('1', 'Moab', '2025-07-01T00:00:00Z'), location: 'Utah' }
    render(<TripTimeline stops={[stop]} />)
    expect(screen.getByText('Utah')).toBeInTheDocument()
  })

  it('renders stop tags as pills', () => {
    const stop = {
      ...makeStop('1', 'Red Rock', '2025-08-01T00:00:00Z'),
      tags: [
        { id: 'tag-1', name: 'Camping', slug: 'camping', created_at: '2025-01-01T00:00:00Z' },
      ],
    }
    render(<TripTimeline stops={[stop]} />)
    expect(screen.getByText('Camping')).toBeInTheDocument()
  })

  it('renders the arrived_at date in a human-readable form', () => {
    const stops = [makeStop('1', 'Yellowstone', '2025-06-10T00:00:00Z')]
    render(<TripTimeline stops={stops} />)
    // The word "Jun" should appear somewhere in the rendered date
    expect(screen.getByText(/Jun/)).toBeInTheDocument()
  })

  it('handles stops with null arrived_at by placing them at the end', () => {
    const stops = [
      { ...makeStop('2', 'Unknown Date', null as unknown as string) },
      makeStop('1', 'Dated Stop', '2025-06-10T00:00:00Z'),
    ]
    render(<TripTimeline stops={stops} />)
    const entries = screen.getAllByTestId('timeline-stop-name').map((el) => el.textContent)
    expect(entries).toEqual(['Dated Stop', 'Unknown Date'])
  })
})
