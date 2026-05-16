export type PackingList = {
  id: string
  name: string
  description?: string | null
  created_at: string
  updated_at: string
}

export type PackingListCreate = {
  name: string
  description?: string
}

export type PackingListUpdate = {
  name?: string
  description?: string
}

export type Label = {
  id: string
  name: string
  color?: string | null
  created_at: string
  updated_at: string
}

export type LabelCreate = {
  name: string
  color?: string
}

export type PackingListLabelAdd = {
  label_id: string
}
