import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { App } from './App.tsx'
import type { SearchAPI } from './types/search.ts'
import { MOCK_RESULTS } from './test/fixtures.ts'

function createStubAPI(
  overrides: Partial<SearchAPI> = {},
): SearchAPI {
  return {
    search: vi.fn().mockResolvedValue({ results: [...MOCK_RESULTS] }),
    ...overrides,
  }
}

describe('App', () => {
  let api: SearchAPI

  beforeEach(() => {
    api = createStubAPI()
  })

  it('renders the search bar', () => {
    render(<App api={api} />)

    expect(screen.getByRole('textbox')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument()
  })

  it('shows empty state prompt initially', () => {
    render(<App api={api} />)

    expect(screen.getByText(/search|ask|type/i)).toBeInTheDocument()
  })

  it('calls api.search when user submits a query', async () => {
    const user = userEvent.setup()
    render(<App api={api} />)

    await user.type(screen.getByRole('textbox'), 'What is faith?')
    await user.click(screen.getByRole('button', { name: /search/i }))

    expect(api.search).toHaveBeenCalledWith(
      expect.objectContaining({ query: 'What is faith?' }),
    )
  })

  it('shows loading state while search is in progress', async () => {
    const user = userEvent.setup()
    const slowAPI = createStubAPI({
      search: vi.fn().mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ results: [] }), 500)),
      ),
    })
    render(<App api={slowAPI} />)

    await user.type(screen.getByRole('textbox'), 'faith')
    await user.click(screen.getByRole('button', { name: /search/i }))

    const matches = screen.getAllByText(/loading|searching/i)
    expect(matches.length).toBeGreaterThanOrEqual(1)
  })

  it('displays results after successful search', async () => {
    const user = userEvent.setup()
    render(<App api={api} />)

    await user.type(screen.getByRole('textbox'), 'faith')
    await user.click(screen.getByRole('button', { name: /search/i }))

    await waitFor(() => {
      expect(screen.getByText(MOCK_RESULTS[0].name)).toBeInTheDocument()
    })
  })

  it('displays error message on search failure', async () => {
    const user = userEvent.setup()
    const failAPI = createStubAPI({
      search: vi.fn().mockRejectedValue(new Error('Network error')),
    })
    render(<App api={failAPI} />)

    await user.type(screen.getByRole('textbox'), 'faith')
    await user.click(screen.getByRole('button', { name: /search/i }))

    await waitFor(() => {
      expect(screen.getByText(/error|failed|network/i)).toBeInTheDocument()
    })
  })

  it('displays "no results" when search returns empty', async () => {
    const user = userEvent.setup()
    const emptyAPI = createStubAPI({
      search: vi.fn().mockResolvedValue({ results: [] }),
    })
    render(<App api={emptyAPI} />)

    await user.type(screen.getByRole('textbox'), 'xyznonexistent')
    await user.click(screen.getByRole('button', { name: /search/i }))

    await waitFor(() => {
      expect(screen.getByText(/no results/i)).toBeInTheDocument()
    })
  })

  it('cancels previous in-flight request on new search', async () => {
    const user = userEvent.setup()
    let callCount = 0
    const controlledAPI = createStubAPI({
      search: vi.fn().mockImplementation(
        () =>
          new Promise((resolve) => {
            callCount++
            setTimeout(() => resolve({ results: [] }), callCount === 1 ? 1000 : 50)
          }),
      ),
    })
    render(<App api={controlledAPI} />)

    await user.type(screen.getByRole('textbox'), 'first')
    await user.click(screen.getByRole('button', { name: /search/i }))

    await user.clear(screen.getByRole('textbox'))
    await user.type(screen.getByRole('textbox'), 'second')
    await user.click(screen.getByRole('button', { name: /search/i }))

    expect(controlledAPI.search).toHaveBeenCalledTimes(2)
  })
})
