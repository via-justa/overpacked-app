import { apiClient } from '../../../lib/api/client'
import { unwrapApiResponse } from '../../../lib/api/request'
import type { components } from '../../../lib/api/schema'

// Server types are sourced from the generated OpenAPI schema (single source of truth).
export type SearchEntityType = components['schemas']['SearchEntityType']
export type SearchResult = components['schemas']['SearchResult']

const SEARCH_ERROR_FALLBACK = 'Unable to perform search'

export async function globalSearch(
    q: string,
    types?: SearchEntityType[],
    limit?: number,
): Promise<SearchResult[]> {
    const query: { q: string; types?: SearchEntityType[]; limit?: number } = { q }
    if (types && types.length > 0) {
        query.types = types
    }
    if (typeof limit === 'number') {
        query.limit = limit
    }

    return unwrapApiResponse(apiClient.GET('/api/v1/search', { params: { query } }), SEARCH_ERROR_FALLBACK)
}
