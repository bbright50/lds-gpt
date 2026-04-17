import type { SearchResult } from '../types/search.ts'
import { ResultCard } from './ResultCard.tsx'

interface ResultListProps {
  results: SearchResult[]
}

export function ResultList({ results }: ResultListProps): React.ReactElement | null {
  if (results.length === 0) {
    return null
  }

  return (
    <>
      {results.map((result) => (
        <ResultCard key={result.id} result={result} />
      ))}
    </>
  )
}
