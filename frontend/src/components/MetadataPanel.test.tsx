import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MetadataPanel } from './MetadataPanel.tsx'
import type { ResultMeta } from '../types/search.ts'

describe('MetadataPanel', () => {
  it('renders non-empty metadata fields', () => {
    const metadata: ResultMeta = {
      chapterNumber: 11,
      url: 'https://example.com',
      summary: 'A chapter summary.',
    }

    render(<MetadataPanel metadata={metadata} />)

    expect(screen.getByTestId('metadata-panel')).toBeInTheDocument()
    expect(screen.getByText('11')).toBeInTheDocument()
    expect(screen.getByText(/chapter summary/i)).toBeInTheDocument()
  })

  it('does not render fields that are undefined', () => {
    const metadata: ResultMeta = {
      chapterNumber: 5,
    }

    render(<MetadataPanel metadata={metadata} />)

    expect(screen.queryByText(/url/i)).not.toBeInTheDocument()
    expect(screen.queryByText(/summary/i)).not.toBeInTheDocument()
    expect(screen.queryByText(/book/i)).not.toBeInTheDocument()
  })

  it('renders fields that are 0', () => {
    const metadata: ResultMeta = {
      startVerseNumber: 0,
      chapterNumber: 5,
    }

    render(<MetadataPanel metadata={metadata} />)

    expect(screen.getByText('0')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('does not render fields that are empty string', () => {
    const metadata: ResultMeta = {
      url: '',
      summary: 'Present',
    }

    render(<MetadataPanel metadata={metadata} />)

    expect(screen.queryByText(/url/i)).not.toBeInTheDocument()
    expect(screen.getByText(/present/i)).toBeInTheDocument()
  })

  it('renders nothing when metadata has only url', () => {
    const metadata: ResultMeta = {
      url: 'https://example.com/chapter',
    }

    const { container } = render(<MetadataPanel metadata={metadata} />)

    expect(container.textContent).toBe('')
  })

  it('renders nothing when metadata is completely empty', () => {
    const { container } = render(<MetadataPanel metadata={{}} />)

    expect(container.textContent).toBe('')
  })
})
