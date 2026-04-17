import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { createSearchClient } from './api/client.ts'
import { createMockSearchAPI } from './api/mockSearch.ts'
import type { SearchAPI } from './types/search.ts'

const baseUrl = import.meta.env.VITE_API_BASE_URL as string | undefined

const api: SearchAPI = baseUrl
  ? createSearchClient(baseUrl)
  : createMockSearchAPI()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App api={api} />
  </StrictMode>,
)
