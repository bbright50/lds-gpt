import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ResultList } from './ResultList.tsx'
import { MOCK_RESULTS, EMPTY_RESULTS } from '../test/fixtures.ts'

describe('ResultList', () => {
  it('renders a ResultCard for each search result', () => {
    render(<ResultList results={[...MOCK_RESULTS]} />)

    for (const result of MOCK_RESULTS) {
      const matches = screen.getAllByText(result.name)
      expect(matches.length).toBeGreaterThanOrEqual(1)
    }
  })

  it('renders nothing when results array is empty', () => {
    const { container } = render(<ResultList results={[...EMPTY_RESULTS]} />)

    expect(container.children).toHaveLength(0)
  })

  it('renders results in the order provided', () => {
    render(<ResultList results={[...MOCK_RESULTS]} />)

    const names = screen.getAllByTestId('result-card-name').map((el) => el.textContent)

    expect(names).toEqual(MOCK_RESULTS.map((r) => r.name))
  })
})
