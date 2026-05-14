import type { ConditioningLevel, Person } from './types'

/**
 * Calculates the recommended maximum backpack weight based on age, gender, and conditioning.
 * Formula: Body Weight × 0.12 × F_age × F_gender × F_conditioning
 *
 * @param bodyWeightGrams - Person's body weight in grams
 * @param birthdate - Person's birthdate (ISO format YYYY-MM-DD)
 * @param gender - Person's gender
 * @param conditioningLevel - Person's conditioning level
 * @returns Recommended max weight in grams, or 0 if required data is missing
 */
export function calculateRecommendedMaxWeightGrams(
  bodyWeightGrams?: number | null,
  birthdate?: string | null,
  gender?: string | null,
  conditioningLevel?: ConditioningLevel | null,
): number {
  if (!bodyWeightGrams || bodyWeightGrams <= 0) {
    return 0
  }

  const bodyWeightKg = bodyWeightGrams / 1000

  const ageFactor = getAgeFactor(birthdate)
  const genderFactor = getGenderFactor(gender)
  const conditioningFactor = getConditioningFactor(conditioningLevel)

  // Formula: bodyWeightKg × 0.12 × ageFactor × genderFactor × conditioningFactor
  const recommendedWeightKg = bodyWeightKg * 0.12 * ageFactor * genderFactor * conditioningFactor

  return recommendedWeightKg * 1000
}

/**
 * Calculates the recommended max weight for a person from their complete profile
 */
export function getPersonRecommendedMaxWeightGrams(person: Person): number {
  return calculateRecommendedMaxWeightGrams(
    person.body_weight_grams,
    person.birthdate,
    person.gender,
    person.conditioning_level,
  )
}

function getAgeFactor(birthdate?: string | null): number {
  if (!birthdate) {
    return 1.1 // Default to adult peak (19-50)
  }

  const age = calculateAge(birthdate)

  if (age < 5) return 0.75 // Very young children
  if (age < 9) return 0.75 // Ages 5-8
  if (age < 13) return 0.85 // Ages 9-12
  if (age < 16) return 0.95 // Ages 13-15
  if (age < 19) return 1 // Ages 16-18
  if (age <= 50) return 1.1 // Ages 19-50 (peak)
  return 0.9 // Ages 50+
}

function getGenderFactor(gender?: string | null): number {
  switch (gender) {
    case 'female':
      return 0.95
    case 'male':
      return 1.05
    default:
      return 1 // Default neutral
  }
}

function getConditioningFactor(conditioningLevel?: ConditioningLevel | null): number {
  switch (conditioningLevel) {
    case 'sedentary':
      return 0.85
    case 'average':
      return 1
    case 'athletic':
      return 1.15
    case 'military':
      return 1.2
    default:
      return 1
  }
}

function calculateAge(birthdate: string): number {
  const today = new Date()
  const birth = new Date(birthdate)

  let age = today.getFullYear() - birth.getFullYear()

  // Adjust if birthday hasn't occurred this year
  const monthDayToday = today.getMonth() * 100 + today.getDate()
  const monthDayBirth = birth.getMonth() * 100 + birth.getDate()

  if (monthDayToday < monthDayBirth) {
    age--
  }

  return age
}
