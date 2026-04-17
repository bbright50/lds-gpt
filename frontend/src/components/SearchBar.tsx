import { useState } from 'react'

interface SearchBarProps {
  onSearch: (query: string) => void
  loading?: boolean
}

export function SearchBar({ onSearch, loading = false }: SearchBarProps): React.ReactElement {
  const [query, setQuery] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = query.trim()
    if (trimmed.length > 0) {
      onSearch(trimmed)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search scriptures..."
      />
      <button
        type="submit"
        disabled={query.trim().length === 0}
        aria-label="Search"
        aria-busy={loading}
      >
        {loading ? 'Searching...' : 'Go'}
      </button>
    </form>
  )
}
