import { apiClient } from '../../../lib/api/client'

export type SearchEntityType = 'item' | 'set' | 'packing_list' | 'person' | 'manufacturer' | 'trip'

export interface SearchResult {
    entity_type: SearchEntityType
    id: string
    title: string
    subtitle?: string | null
    score: number
}

const SEARCH_ERROR_FALLBACK = 'Unable to perform search'

function getErrorMessage(error: unknown, fallback: string): string {
    if (error && typeof error === 'object' && 'error' in error) {
        const message = (error as { error?: unknown }).error
        if (typeof message === 'string' && message.trim().length > 0) {
            return message
        }
    }

    return fallback
}

export async function globalSearch(
    q: string,
    types?: SearchEntityType[],
    limit?: number,
): Promise<SearchResult[]> {
    const query: { q: string; types?: string[]; limit?: number } = { q }
    if (types && types.length > 0) {
        query.types = types
    }
    if (typeof limit === 'number') {
        query.limit = limit
    }

    const { data, error, response } = await apiClient.GET('/api/v1/search', {
        params: { query },
    })

    if (!response.ok || !data) {
        throw new Error(getErrorMessage(error, SEARCH_ERROR_FALLBACK))
    }

    return data as unknown as SearchResult[]
}
