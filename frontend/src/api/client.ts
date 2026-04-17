import type { SearchAPI, SearchRequest, SearchResponse } from '../types/search.ts'

export function createSearchClient(baseUrl: string): SearchAPI {
  return {
    async search(request: SearchRequest): Promise<SearchResponse> {
      const { signal, ...body } = request
      const response = await fetch(`${baseUrl}/api/search`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
        signal,
      })

      if (!response.ok) {
        throw new Error(`Search request failed: ${response.status} ${response.statusText}`)
      }

      const data: SearchResponse = await response.json()
      return data
    },
  }
}
