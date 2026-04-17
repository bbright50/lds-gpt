import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EntityIcon } from './EntityIcon.tsx'
import { ENTITY_TYPES } from '../types/search.ts'
import type { EntityType } from '../types/search.ts'

describe('EntityIcon', () => {
  it.each(ENTITY_TYPES as unknown as EntityType[])(
    'renders an icon for entity type "%s"',
    (entityType) => {
      render(<EntityIcon entityType={entityType} />)

      const icon = screen.getByTestId('entity-icon')
      expect(icon).toBeInTheDocument()
      expect(icon.textContent!.length).toBeGreaterThan(0)
    },
  )

  it('renders different icons for different entity types', () => {
    const { rerender } = render(<EntityIcon entityType="verse" />)
    const verseIcon = screen.getByTestId('entity-icon').textContent

    rerender(<EntityIcon entityType="chapter" />)
    const chapterIcon = screen.getByTestId('entity-icon').textContent

    expect(verseIcon).not.toBe(chapterIcon)
  })
})
