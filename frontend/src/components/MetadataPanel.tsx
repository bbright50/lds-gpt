import type { ResultMeta } from '../types/search.ts'

interface MetadataPanelProps {
  metadata: ResultMeta
}

const FIELD_LABELS: Record<string, string> = {
  startVerseNumber: 'Start Verse',
  endVerseNumber: 'End Verse',
  chapterID: 'Chapter ID',
  chapterNumber: 'Chapter',
  book: 'Book',
  chapter: 'Chapter',
  comprises: 'Comprises',
  compareRef: 'Compare',
  summary: 'Summary',
  verseNumber: 'Verse',
  reference: 'Reference',
}

function isDisplayable(value: unknown): boolean {
  if (value === undefined || value === null) return false
  if (typeof value === 'string' && value === '') return false
  return true
}

export function MetadataPanel({ metadata }: MetadataPanelProps): React.ReactElement | null {
  const entries = Object.entries(metadata).filter(
    ([key, value]) => key !== 'url' && isDisplayable(value),
  )

  if (entries.length === 0) {
    return null
  }

  return (
    <div data-testid="metadata-panel">
      {entries.map(([key, value]) => (
        <div key={key}>
          <span>{FIELD_LABELS[key] ?? key}: </span>
          <span>{String(value)}</span>
        </div>
      ))}
    </div>
  )
}
