/**
 * Unit conversion constants and utilities
 * All backend storage uses canonical units (grams, ml, etc.)
 * These functions convert between canonical units and display units
 */

export const GRAMS_PER_OUNCE = 28.349523125
export const GRAMS_PER_KILOGRAM = 1000
export const OUNCES_PER_POUND = 16
export const ML_PER_FL_OZ = 29.5735295625
export const LB_PER_KG = 2.2046226218

export type WeightUnit = 'g' | 'oz'
export type VolumeUnit = 'ml' | 'fl_oz'

// ─── Weight Conversions ────────────────────────────────────────────────────────

export function gramsToOunces(grams: number): number {
  return grams / GRAMS_PER_OUNCE
}

export function ouncesToGrams(ounces: number): number {
  return ounces * GRAMS_PER_OUNCE
}

export function gramsToKilograms(grams: number): number {
  return grams / GRAMS_PER_KILOGRAM
}

export function kilogramsToGrams(kg: number): number {
  return kg * GRAMS_PER_KILOGRAM
}

export function gramsToInput(grams: number, unit: WeightUnit): number {
  return unit === 'oz' ? gramsToOunces(grams) : grams
}

export function inputToGrams(value: number, unit: WeightUnit): number {
  return unit === 'oz' ? ouncesToGrams(value) : value
}

// ─── Volume Conversions ────────────────────────────────────────────────────────

export function mlToFluidOunces(ml: number): number {
  return ml / ML_PER_FL_OZ
}

export function fluidOuncesToMl(flOz: number): number {
  return flOz * ML_PER_FL_OZ
}

export function mlToInput(ml: number, unit: VolumeUnit): number {
  return unit === 'fl_oz' ? mlToFluidOunces(ml) : ml
}

export function inputToMl(value: number, unit: VolumeUnit): number {
  return unit === 'fl_oz' ? fluidOuncesToMl(value) : value
}

// ─── Display Formatting ─────────────────────────────────────────────────────────

export function toRoundedString(value: number): string {
  if (!Number.isFinite(value)) {
    return '0'
  }
  return Number.parseFloat(value.toFixed(2)).toString()
}

export function formatDisplayWeight(
  valueGrams: number,
  unit: WeightUnit
): string {
  if (unit === 'oz') {
    const ounces = gramsToOunces(valueGrams)
    if (Math.abs(ounces) >= OUNCES_PER_POUND) {
      const pounds = ounces / OUNCES_PER_POUND
      return `${toRoundedString(pounds)} lb`
    }
    return `${toRoundedString(ounces)} oz`
  }

  if (Math.abs(valueGrams) >= GRAMS_PER_KILOGRAM) {
    return `${toRoundedString(gramsToKilograms(valueGrams))} kg`
  }

  return `${toRoundedString(valueGrams)} g`
}
