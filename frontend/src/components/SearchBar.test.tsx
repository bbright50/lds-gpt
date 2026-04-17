import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SearchBar } from './SearchBar.tsx'

describe('SearchBar', () => {
  it('renders an input field and submit button', () => {
    render(<SearchBar onSearch={vi.fn()} />)

    expect(screen.getByRole('textbox')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument()
  })

  it('submit button is disabled when input is empty', () => {
    render(<SearchBar onSearch={vi.fn()} />)

    expect(screen.getByRole('button', { name: /search/i })).toBeDisabled()
  })

  it('submit button is enabled when input has text', async () => {
    const user = userEvent.setup()
    render(<SearchBar onSearch={vi.fn()} />)

    await user.type(screen.getByRole('textbox'), 'What is faith?')

    expect(screen.getByRole('button', { name: /search/i })).toBeEnabled()
  })

  it('calls onSearch with query when form is submitted via button', async () => {
    const user = userEvent.setup()
    const onSearch = vi.fn()
    render(<SearchBar onSearch={onSearch} />)

    await user.type(screen.getByRole('textbox'), 'What is faith?')
    await user.click(screen.getByRole('button', { name: /search/i }))

    expect(onSearch).toHaveBeenCalledWith('What is faith?')
  })

  it('calls onSearch with query when Enter is pressed', async () => {
    const user = userEvent.setup()
    const onSearch = vi.fn()
    render(<SearchBar onSearch={onSearch} />)

    await user.type(screen.getByRole('textbox'), 'What is faith?{enter}')

    expect(onSearch).toHaveBeenCalledWith('What is faith?')
  })

  it('does not call onSearch when Enter is pressed with empty input', async () => {
    const user = userEvent.setup()
    const onSearch = vi.fn()
    render(<SearchBar onSearch={onSearch} />)

    await user.type(screen.getByRole('textbox'), '{enter}')

    expect(onSearch).not.toHaveBeenCalled()
  })

  it('shows loading text on button when loading is true', async () => {
    const user = userEvent.setup()
    render(<SearchBar onSearch={vi.fn()} loading={true} />)

    await user.type(screen.getByRole('textbox'), 'faith')

    const button = screen.getByRole('button', { name: /search/i })
    expect(button).toBeEnabled()
    expect(button).toHaveAttribute('aria-busy', 'true')
    expect(button).toHaveTextContent('Searching...')
  })
})
