/** Entity types matching Go backend's EntityType constants. */
export type EntityType =
  | 'verse_group'
  | 'chapter'
  | 'topical_guide'
  | 'bible_dict'
  | 'index'
  | 'jst_passage'
  | 'verse'

/** All supported entity types as an array for iteration. */
export const ENTITY_TYPES: readonly EntityType[] = [
  'verse_group',
  'chapter',
  'topical_guide',
  'bible_dict',
  'index',
  'jst_passage',
  'verse',
] as const

/** Entity-specific metadata fields matching Go backend's ResultMeta struct. */
export interface ResultMeta {
  startVerseNumber?: number
  endVerseNumber?: number
  chapterID?: number
  chapterNumber?: number
  url?: string
  book?: string
  chapter?: string
  comprises?: string
  compareRef?: string
  summary?: string
  verseNumber?: number
  reference?: string
}

/** A single search result matching Go backend's SearchResult struct. */
export interface SearchResult {
  entityType: EntityType
  id: number
  name: string
  text: string
  distance: number
  metadata: ResultMeta
}

/** Request payload for the search endpoint. */
export interface SearchRequest {
  query: string
  knn?: number
  signal?: AbortSignal
}

/** Response payload from the search endpoint. */
export interface SearchResponse {
  results: SearchResult[]
}

/** API client contract for search operations. */
export interface SearchAPI {
  search(request: SearchRequest): Promise<SearchResponse>
}
