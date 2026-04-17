import { useState } from 'react'
import type { SearchResult } from '../types/search.ts'
import { EntityIcon } from './EntityIcon.tsx'
import { DistanceBadge } from './DistanceBadge.tsx'
import { MetadataPanel } from './MetadataPanel.tsx'

interface ResultCardProps {
  result: SearchResult
}

const MAX_TEXT_LENGTH = 300

function formatEntityType(entityType: string): string {
  return entityType.replace(/_/g, ' ')
}

export function ResultCard({ result }: ResultCardProps): React.ReactElement {
  const [showMetadata, setShowMetadata] = useState(false)

  const truncatedText = result.text.length > MAX_TEXT_LENGTH
    ? result.text.slice(0, MAX_TEXT_LENGTH) + '...'
    : result.text

  const hasUrl = typeof result.metadata.url === 'string' && result.metadata.url.length > 0

  return (
    <div data-testid="result-card">
      <div>
        <EntityIcon entityType={result.entityType} />
        <span>{formatEntityType(result.entityType)}</span>
        <DistanceBadge distance={result.distance} />
      </div>
      <div data-testid="result-card-name">{result.name}</div>
      <div data-testid="result-card-text">{truncatedText}</div>
      {hasUrl && (
        <a href={result.metadata.url} target="_blank" rel="noopener noreferrer">
          Source
        </a>
      )}
      <button
        type="button"
        onClick={() => setShowMetadata((prev) => !prev)}
        aria-label="Show details"
      >
        {showMetadata ? 'Hide Details' : 'Details'}
      </button>
      {showMetadata && <MetadataPanel metadata={result.metadata} />}
    </div>
  )
}
