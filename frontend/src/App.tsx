import { useState, useRef, useCallback } from 'react'
import type { SearchAPI, SearchResult } from './types/search.ts'
import { SearchBar } from './components/SearchBar.tsx'
import { SearchStatus } from './components/SearchStatus.tsx'
import { ResultList } from './components/ResultList.tsx'

interface AppProps {
  api: SearchAPI
}

export function App({ api }: AppProps): React.ReactElement {
  const [results, setResults] = useState<SearchResult[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasSearched, setHasSearched] = useState(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  const handleSearch = useCallback(async (query: string) => {
    abortControllerRef.current?.abort()
    const controller = new AbortController()
    abortControllerRef.current = controller

    setLoading(true)
    setError(null)

    try {
      const response = await api.search({ query, signal: controller.signal })

      if (controller.signal.aborted) {
        return
      }

      setResults(response.results)
      setHasSearched(true)
    } catch (err) {
      if (controller.signal.aborted) {
        return
      }

      const message = err instanceof Error ? err.message : 'Search failed'
      setError(message)
      setResults([])
      setHasSearched(true)
    } finally {
      if (!controller.signal.aborted) {
        setLoading(false)
      }
    }
  }, [api])

  return (
    <div>
      <SearchBar onSearch={handleSearch} loading={loading} />
      <SearchStatus
        loading={loading}
        error={error}
        hasResults={results.length > 0}
        hasSearched={hasSearched}
      />
      <ResultList results={results} />
    </div>
  )
}

export default App
