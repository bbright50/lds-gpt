import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ResultCard } from './ResultCard.tsx'
import { SINGLE_RESULT, MOCK_RESULTS } from '../test/fixtures.ts'

describe('ResultCard', () => {
  it('displays the result name', () => {
    render(<ResultCard result={SINGLE_RESULT} />)

    expect(screen.getByText(SINGLE_RESULT.name)).toBeInTheDocument()
  })

  it('displays the entity type label', () => {
    render(<ResultCard result={SINGLE_RESULT} />)

    expect(screen.getByText(/verse.group/i)).toBeInTheDocument()
  })

  it('displays truncated text content', () => {
    render(<ResultCard result={SINGLE_RESULT} />)

    const textEl = screen.getByTestId('result-card-text')
    expect(textEl).toBeInTheDocument()
    expect(textEl.textContent).toContain('faith')
  })

  it('displays the entity icon', () => {
    render(<ResultCard result={SINGLE_RESULT} />)

    expect(screen.getByTestId('entity-icon')).toBeInTheDocument()
  })

  it('displays the distance badge', () => {
    render(<ResultCard result={SINGLE_RESULT} />)

    expect(screen.getByTestId('distance-badge')).toBeInTheDocument()
  })

  it('has a collapsible metadata section', async () => {
    const user = userEvent.setup()
    const chapterResult = MOCK_RESULTS.find((r) => r.entityType === 'chapter')!
    render(<ResultCard result={chapterResult} />)

    const toggle = screen.getByRole('button', { name: /metadata|details/i })
    expect(toggle).toBeInTheDocument()

    await user.click(toggle)

    expect(screen.getByTestId('metadata-panel')).toBeInTheDocument()
  })

  it('displays source link when URL is in metadata', () => {
    const chapterResult = MOCK_RESULTS.find((r) => r.entityType === 'chapter')!
    render(<ResultCard result={chapterResult} />)

    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', chapterResult.metadata.url)
    expect(link).toHaveAttribute('target', '_blank')
  })

  it('does not display source link when URL is absent', () => {
    const tgResult = MOCK_RESULTS.find(
      (r) => r.entityType === 'topical_guide',
    )!
    render(<ResultCard result={tgResult} />)

    expect(screen.queryByRole('link')).not.toBeInTheDocument()
  })
})
