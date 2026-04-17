import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SearchStatus } from './SearchStatus.tsx'

describe('SearchStatus', () => {
  it('renders loading indicator when loading is true', () => {
    render(
      <SearchStatus
        loading={true}
        error={null}
        hasResults={false}
        hasSearched={false}
      />,
    )

    expect(screen.getByText(/loading|searching/i)).toBeInTheDocument()
  })

  it('renders error message when error is provided', () => {
    render(
      <SearchStatus
        loading={false}
        error="Something went wrong"
        hasResults={false}
        hasSearched={true}
      />,
    )

    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
  })

  it('renders empty state prompt when no search has been performed', () => {
    render(
      <SearchStatus
        loading={false}
        error={null}
        hasResults={false}
        hasSearched={false}
      />,
    )

    expect(screen.getByText(/search|ask|type/i)).toBeInTheDocument()
  })

  it('renders "no results found" when search returned empty', () => {
    render(
      <SearchStatus
        loading={false}
        error={null}
        hasResults={false}
        hasSearched={true}
      />,
    )

    expect(screen.getByText(/no results/i)).toBeInTheDocument()
  })

  it('renders nothing when results are present', () => {
    const { container } = render(
      <SearchStatus
        loading={false}
        error={null}
        hasResults={true}
        hasSearched={true}
      />,
    )

    expect(container.textContent).toBe('')
  })
})
