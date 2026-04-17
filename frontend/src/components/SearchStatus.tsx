interface SearchStatusProps {
  loading: boolean
  error: string | null
  hasResults: boolean
  hasSearched: boolean
}

export function SearchStatus({ loading, error, hasResults, hasSearched }: SearchStatusProps): React.ReactElement | null {
  if (loading) {
    return <div>Searching...</div>
  }

  if (error) {
    return <div>{error}</div>
  }

  if (!hasSearched) {
    return <div>Type a question to search the scriptures</div>
  }

  if (!hasResults) {
    return <div>No results found</div>
  }

  return null
}
