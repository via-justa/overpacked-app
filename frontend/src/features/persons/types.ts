export type Gender = 'male' | 'female' | 'other'

export type ConditioningLevel = 'sedentary' | 'average' | 'athletic' | 'military'

export type Person = {
  id: string
  name: string
  gender?: Gender | null
  birthdate?: string | null
  body_weight_grams?: number | null
  conditioning_level?: ConditioningLevel | null
  created_at: string
  updated_at: string
}

export type PersonCreate = {
  name: string
  gender?: Gender
  birthdate?: string
  body_weight_grams?: number
  conditioning_level?: ConditioningLevel
}

export type PersonUpdate = {
  name?: string
  gender?: Gender
  birthdate?: string
  body_weight_grams?: number
  conditioning_level?: ConditioningLevel
}

export type PersonFormValues = {
  name: string
  gender: Gender
  birthdate: string
  body_weight_value: string
  conditioning_level: ConditioningLevel
}
