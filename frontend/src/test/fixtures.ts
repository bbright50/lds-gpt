import type { SearchResult } from '../types/search.ts'

/** Canned search results covering all 7 entity types. */
export const MOCK_RESULTS: readonly SearchResult[] = [
  {
    entityType: 'verse_group',
    id: 1,
    name: 'Alma 32:21-23',
    text: 'And now as I said concerning faith—faith is not to have a perfect knowledge of things; therefore if ye have faith ye hope for things which are not seen, which are true.',
    distance: 0.15,
    metadata: {
      startVerseNumber: 21,
      endVerseNumber: 23,
      chapterID: 100,
    },
  },
  {
    entityType: 'chapter',
    id: 2,
    name: 'Hebrews 11',
    text: 'Now faith is the substance of things hoped for, the evidence of things not seen.',
    distance: 0.22,
    metadata: {
      chapterNumber: 11,
      url: 'https://www.churchofjesuschrist.org/study/scriptures/nt/heb/11',
      summary: 'By faith we understand that the worlds were framed by the word of God.',
    },
  },
  {
    entityType: 'topical_guide',
    id: 3,
    name: 'Faith',
    text: 'See also Belief; Trust in God; Faithful.',
    distance: 0.30,
    metadata: {},
  },
  {
    entityType: 'bible_dict',
    id: 4,
    name: 'Faith',
    text: 'A principle of action and power that motivates day-to-day activities.',
    distance: 0.35,
    metadata: {},
  },
  {
    entityType: 'index',
    id: 5,
    name: 'Faith, Faithful, Faithfulness',
    text: 'References to faith throughout the triple combination.',
    distance: 0.40,
    metadata: {},
  },
  {
    entityType: 'jst_passage',
    id: 6,
    name: 'JST Hebrews 11:1',
    text: 'Now faith is the assurance of things hoped for, the evidence of things not seen.',
    distance: 0.50,
    metadata: {
      book: 'Hebrews',
      chapter: '11',
      comprises: 'Hebrews 11:1 (KJV)',
      compareRef: 'Hebrews 11:1',
      summary: 'JST changes "substance" to "assurance".',
    },
  },
  {
    entityType: 'verse',
    id: 7,
    name: 'Moroni 7:33',
    text: 'And Christ hath said: If ye will have faith in me ye shall have power to do whatsoever thing is expedient in me.',
    distance: 0.60,
    metadata: {
      verseNumber: 33,
      reference: 'Moroni 7:33',
    },
  },
] as const

/** A single verse_group result for focused testing. */
export const SINGLE_RESULT: SearchResult = MOCK_RESULTS[0]

/** Empty results for testing no-results state. */
export const EMPTY_RESULTS: readonly SearchResult[] = [] as const

/** Result with distance=0 for edge case testing. */
export const PERFECT_MATCH_RESULT: SearchResult = {
  entityType: 'verse',
  id: 99,
  name: 'Exact Match',
  text: 'This is an exact match result.',
  distance: 0,
  metadata: {
    verseNumber: 1,
    reference: 'Test 1:1',
  },
}

/** Result with very long text for truncation testing. */
export const LONG_TEXT_RESULT: SearchResult = {
  entityType: 'chapter',
  id: 100,
  name: 'Long Chapter',
  text: 'Lorem ipsum '.repeat(200).trim(),
  distance: 0.25,
  metadata: {
    chapterNumber: 1,
    url: 'https://example.com/long',
  },
}
