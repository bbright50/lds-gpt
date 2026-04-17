import { describe, it, expect } from 'vitest'
import { createMockSearchAPI } from './mockSearch.ts'
import { ENTITY_TYPES } from '../types/search.ts'
import type { SearchAPI, EntityType } from '../types/search.ts'

describe('mockSearch', () => {
  function getAPI(): SearchAPI {
    return createMockSearchAPI()
  }

  it('returns a valid SearchResponse', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'What is faith?' })

    expect(response).toBeDefined()
    expect(response.results).toBeDefined()
    expect(Array.isArray(response.results)).toBe(true)
  })

  it('returns results covering all 7 entity types', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith' })
    const returnedTypes = new Set(response.results.map((r) => r.entityType))

    for (const entityType of ENTITY_TYPES) {
      expect(
        returnedTypes.has(entityType),
        `missing entity type: ${entityType}`,
      ).toBe(true)
    }
  })

  it('each result has required fields', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith' })

    for (const result of response.results) {
      expect(typeof result.id).toBe('number')
      expect(typeof result.name).toBe('string')
      expect(result.name.length).toBeGreaterThan(0)
      expect(typeof result.text).toBe('string')
      expect(result.text.length).toBeGreaterThan(0)
      expect(typeof result.distance).toBe('number')
      expect(result.distance).toBeGreaterThanOrEqual(0)
      expect(result.distance).toBeLessThanOrEqual(1)
      expect(result.metadata).toBeDefined()
    }
  })

  it('respects knn limit', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith', knn: 3 })

    expect(response.results.length).toBeLessThanOrEqual(3)
  })

  it('uses default knn of 20 when not specified', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith' })

    expect(response.results.length).toBeLessThanOrEqual(20)
  })

  it('returns empty results for knn of 0', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith', knn: 0 })

    expect(response.results).toEqual([])
  })

  it('returns empty results for negative knn', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith', knn: -5 })

    expect(response.results).toEqual([])
  })

  it('distances span a realistic range (0.1 to 0.9)', async () => {
    const api = getAPI()
    const response = await api.search({ query: 'faith' })
    const distances = response.results.map((r) => r.distance)
    const minDistance = Math.min(...distances)
    const maxDistance = Math.max(...distances)

    expect(minDistance).toBeLessThan(0.5)
    expect(maxDistance).toBeGreaterThan(0.3)
  })

  it('simulates async delay', async () => {
    const api = getAPI()
    const start = performance.now()
    await api.search({ query: 'faith' })
    const elapsed = performance.now() - start

    expect(elapsed).toBeGreaterThanOrEqual(100)
  })

  describe('entity-specific metadata', () => {
    it('verse_group results include startVerseNumber, endVerseNumber, chapterID', async () => {
      const api = getAPI()
      const response = await api.search({ query: 'faith' })
      const vg = response.results.find(
        (r) => r.entityType === 'verse_group' as EntityType,
      )

      expect(vg).toBeDefined()
      expect(vg!.metadata.startVerseNumber).toBeDefined()
      expect(vg!.metadata.endVerseNumber).toBeDefined()
      expect(vg!.metadata.chapterID).toBeDefined()
    })

    it('chapter results include chapterNumber, url, summary', async () => {
      const api = getAPI()
      const response = await api.search({ query: 'faith' })
      const ch = response.results.find(
        (r) => r.entityType === 'chapter' as EntityType,
      )

      expect(ch).toBeDefined()
      expect(ch!.metadata.chapterNumber).toBeDefined()
      expect(ch!.metadata.url).toBeDefined()
      expect(ch!.metadata.summary).toBeDefined()
    })

    it('jst_passage results include book, chapter, comprises, compareRef', async () => {
      const api = getAPI()
      const response = await api.search({ query: 'faith' })
      const jst = response.results.find(
        (r) => r.entityType === 'jst_passage' as EntityType,
      )

      expect(jst).toBeDefined()
      expect(jst!.metadata.book).toBeDefined()
      expect(jst!.metadata.chapter).toBeDefined()
      expect(jst!.metadata.comprises).toBeDefined()
      expect(jst!.metadata.compareRef).toBeDefined()
    })

    it('verse results include verseNumber, reference', async () => {
      const api = getAPI()
      const response = await api.search({ query: 'faith' })
      const v = response.results.find(
        (r) => r.entityType === 'verse' as EntityType,
      )

      expect(v).toBeDefined()
      expect(v!.metadata.verseNumber).toBeDefined()
      expect(v!.metadata.reference).toBeDefined()
    })
  })
})
