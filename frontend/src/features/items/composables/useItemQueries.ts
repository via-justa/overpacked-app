import { useQuery } from '@tanstack/vue-query'
import { listItems, listItemTypes, listManufacturers } from '../api/itemsApi'

/**
 * Reusable query for fetching all items.
 * Uses ['items'] as the query key.
 */
export function useItemsQuery() {
  return useQuery({
    queryKey: ['items'],
    queryFn: listItems,
  })
}

/**
 * Reusable query for fetching all item types (categories).
 * Uses ['item-types'] as the query key.
 */
export function useItemTypesQuery() {
  return useQuery({
    queryKey: ['item-types'],
    queryFn: listItemTypes,
  })
}

/**
 * Reusable query for fetching all manufacturers.
 * Uses ['manufacturers'] as the query key.
 */
export function useManufacturersQuery() {
  return useQuery({
    queryKey: ['manufacturers'],
    queryFn: listManufacturers,
  })
}
