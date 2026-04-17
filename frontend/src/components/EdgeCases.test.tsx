import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DistanceBadge } from './DistanceBadge.tsx'
import { ResultCard } from './ResultCard.tsx'
import { PERFECT_MATCH_RESULT, LONG_TEXT_RESULT } from '../test/fixtures.ts'

describe('Edge Cases', () => {
  describe('distance=0 shows 100% relevance', () => {
    it('renders 100% for distance of exactly 0', () => {
      render(<DistanceBadge distance={0} />)

      expect(screen.getByTestId('distance-badge')).toHaveTextContent('100%')
    })
  })

  describe('long text truncation', () => {
    it('truncates very long text in ResultCard', () => {
      render(<ResultCard result={LONG_TEXT_RESULT} />)

      const textEl = screen.getByTestId('result-card-text')
      const displayedLength = textEl.textContent!.length
      const originalLength = LONG_TEXT_RESULT.text.length

      expect(displayedLength).toBeLessThan(originalLength)
    })
  })

  describe('perfect match result card', () => {
    it('renders a result with distance=0 correctly', () => {
      render(<ResultCard result={PERFECT_MATCH_RESULT} />)

      expect(screen.getByText(PERFECT_MATCH_RESULT.name)).toBeInTheDocument()
      expect(screen.getByTestId('distance-badge')).toHaveTextContent('100%')
    })
  })
})
