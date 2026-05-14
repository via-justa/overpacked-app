import type { Item } from '../items/types'

export type ItemSet = {
  id: string
  name: string
  set_category: string
  created_at: string
  updated_at: string
}

export type ItemSetCreate = {
  name: string
  set_category: string
}

export type ItemSetUpdate = {
  name?: string
  set_category: string
}

export type SetItemCreate = {
  item_id: string
  quantity: number
  notes?: string
  sort_order?: number
}

export type SetItemUpdate = {
  quantity?: number
  notes?: string
  sort_order?: number
}

export type SetItemWithDetails = {
  item_id: string
  quantity: number
  notes?: string | null
  sort_order?: number | null
  item: Item
}
