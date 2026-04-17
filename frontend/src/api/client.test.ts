import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createSearchClient } from './client.ts'
import type { SearchAPI } from '../types/search.ts'

describe('searchClient', () => {
  const mockFetch = vi.fn()

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    mockFetch.mockReset()
  })

  it('implements SearchAPI interface', () => {
    const client: SearchAPI = createSearchClient('http://localhost:8080')

    expect(client.search).toBeDefined()
    expect(typeof client.search).toBe('function')
  })

  it('sends POST request to /api/search', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ results: [] }),
    })

    const client = createSearchClient('http://localhost:8080')
    await client.search({ query: 'faith' })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/search',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
        }),
        body: JSON.stringify({ query: 'faith' }),
      }),
    )
  })

  it('includes knn in request body when provided', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ results: [] }),
    })

    const client = createSearchClient('http://localhost:8080')
    await client.search({ query: 'faith', knn: 5 })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/search',
      expect.objectContaining({
        body: JSON.stringify({ query: 'faith', knn: 5 }),
      }),
    )
  })

  it('returns parsed SearchResponse', async () => {
    const mockResponse = {
      results: [
        {
          entityType: 'verse',
          id: 1,
          name: 'Test',
          text: 'Test text',
          distance: 0.5,
          metadata: {},
        },
      ],
    }

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(mockResponse),
    })

    const client = createSearchClient('http://localhost:8080')
    const response = await client.search({ query: 'test' })

    expect(response.results).toHaveLength(1)
    expect(response.results[0].entityType).toBe('verse')
  })

  it('throws on non-ok HTTP response', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
    })

    const client = createSearchClient('http://localhost:8080')

    await expect(client.search({ query: 'faith' })).rejects.toThrow()
  })

  it('throws on network error', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network error'))

    const client = createSearchClient('http://localhost:8080')

    await expect(client.search({ query: 'faith' })).rejects.toThrow(
      'Network error',
    )
  })

  it('passes AbortSignal to fetch when provided', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ results: [] }),
    })

    const controller = new AbortController()
    const client = createSearchClient('http://localhost:8080')
    await client.search({ query: 'faith', signal: controller.signal })

    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/search',
      expect.objectContaining({
        signal: controller.signal,
        body: JSON.stringify({ query: 'faith' }),
      }),
    )
  })

  it('does not include signal in request body', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ results: [] }),
    })

    const controller = new AbortController()
    const client = createSearchClient('http://localhost:8080')
    await client.search({ query: 'faith', signal: controller.signal })

    const callArgs = mockFetch.mock.calls[0]
    const body = JSON.parse(callArgs[1].body)
    expect(body).toEqual({ query: 'faith' })
    expect(body.signal).toBeUndefined()
  })
})
