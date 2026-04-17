import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DistanceBadge } from './DistanceBadge.tsx'

describe('DistanceBadge', () => {
  it('converts distance to relevance percentage', () => {
    render(<DistanceBadge distance={0.15} />)

    expect(screen.getByTestId('distance-badge')).toHaveTextContent('85%')
  })

  it('shows 100% for distance of 0', () => {
    render(<DistanceBadge distance={0} />)

    expect(screen.getByTestId('distance-badge')).toHaveTextContent('100%')
  })

  it('shows 0% for distance of 1', () => {
    render(<DistanceBadge distance={1} />)

    expect(screen.getByTestId('distance-badge')).toHaveTextContent('0%')
  })

  it('caps at 100% for negative distance', () => {
    render(<DistanceBadge distance={-0.1} />)

    expect(screen.getByTestId('distance-badge')).toHaveTextContent('100%')
  })

  it('caps at 0% for distance greater than 1', () => {
    render(<DistanceBadge distance={1.5} />)

    expect(screen.getByTestId('distance-badge')).toHaveTextContent('0%')
  })

  it('rounds to nearest integer', () => {
    render(<DistanceBadge distance={0.333} />)

    expect(screen.getByTestId('distance-badge')).toHaveTextContent('67%')
  })
})
