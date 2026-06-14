import { z } from 'zod'
import type { PlannerDetails } from './plannerTypes'

// The numeric planner fields are bound to inputs as strings; validate the string
// form so bad entries surface to the user instead of being silently dropped when
// the trip payload is built (see tripPersistence.buildTripPayload). Empty is
// allowed — both fields are optional.
const durationDays = z
    .string()
    .trim()
    .refine((value) => value === '' || (/^\d+$/.test(value) && Number.parseInt(value, 10) > 0), {
        message: 'Enter a whole number of days greater than 0.',
    })

const distanceKm = z
    .string()
    .trim()
    .refine((value) => value === '' || (Number.isFinite(Number(value)) && Number(value) >= 0), {
        message: 'Enter a distance of 0 or more.',
    })

export const plannerDetailsSchema = z.object({
    durationDays,
    distanceKm,
})

export type PlannerDetailsFieldErrorKey = 'durationDays' | 'distanceKm'

/**
 * Validates the numeric planner detail fields and returns a `field → message`
 * map of the fields that fail (empty object when everything is valid).
 */
export function validatePlannerDetails(
    details: Pick<PlannerDetails, PlannerDetailsFieldErrorKey>,
): Partial<Record<PlannerDetailsFieldErrorKey, string>> {
    const result = plannerDetailsSchema.safeParse({
        durationDays: details.durationDays,
        distanceKm: details.distanceKm,
    })
    if (result.success) {
        return {}
    }

    const errors: Partial<Record<PlannerDetailsFieldErrorKey, string>> = {}
    for (const issue of result.error.issues) {
        const key = issue.path[0]
        if ((key === 'durationDays' || key === 'distanceKm') && !errors[key]) {
            errors[key] = issue.message
        }
    }
    return errors
}
