import type { SearchAPI, SearchRequest, SearchResponse } from '../types/search.ts'
import { MOCK_RESULTS } from '../test/fixtures.ts'

const DEFAULT_KNN = 20
const SIMULATED_DELAY_MS = 150

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

export function createMockSearchAPI(): SearchAPI {
  return {
    async search(request: SearchRequest): Promise<SearchResponse> {
      await delay(SIMULATED_DELAY_MS)

      const knn = request.knn ?? DEFAULT_KNN

      if (knn <= 0) {
        return { results: [] }
      }

      const results = MOCK_RESULTS.slice(0, knn)
      return { results: [...results] }
    },
  }
}
