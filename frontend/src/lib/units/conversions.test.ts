import {
  GRAMS_PER_OUNCE,
  bodyWeightGramsToInput,
  bodyWeightInputToGrams,
  formatDisplayWeight,
  fluidOuncesToMl,
  gramsToInput,
  gramsToKilograms,
  gramsToOunces,
  inputToGrams,
  inputToMl,
  kilogramsToGrams,
  mlToFluidOunces,
  mlToInput,
  ouncesToGrams,
  toRoundedString,
} from './conversions'

describe('weight conversions', () => {
  it('converts grams <-> ounces', () => {
    expect(ouncesToGrams(1)).toBeCloseTo(GRAMS_PER_OUNCE, 6)
    expect(gramsToOunces(GRAMS_PER_OUNCE)).toBeCloseTo(1, 6)
  })

  it('converts grams <-> kilograms', () => {
    expect(gramsToKilograms(1500)).toBe(1.5)
    expect(kilogramsToGrams(1.5)).toBe(1500)
  })

  it('gramsToInput / inputToGrams respect the unit', () => {
    expect(gramsToInput(100, 'g')).toBe(100)
    expect(gramsToInput(GRAMS_PER_OUNCE, 'oz')).toBeCloseTo(1, 6)
    expect(inputToGrams(100, 'g')).toBe(100)
    expect(inputToGrams(1, 'oz')).toBeCloseTo(GRAMS_PER_OUNCE, 6)
  })
})

describe('body weight conversions (kg/lb, rounded to grams)', () => {
  it('round-trips kg', () => {
    expect(bodyWeightInputToGrams(70, 'kg')).toBe(70000)
    expect(bodyWeightGramsToInput(70000, 'kg')).toBe(70)
  })

  it('round-trips lb', () => {
    const grams = bodyWeightInputToGrams(154, 'lb')
    expect(grams).toBeGreaterThan(69000)
    expect(bodyWeightGramsToInput(70000, 'lb')).toBeCloseTo(154.32, 1)
  })
})

describe('volume conversions', () => {
  it('converts ml <-> fluid ounces', () => {
    expect(mlToFluidOunces(fluidOuncesToMl(2))).toBeCloseTo(2, 6)
  })

  it('mlToInput / inputToMl respect the unit', () => {
    expect(mlToInput(100, 'ml')).toBe(100)
    expect(mlToInput(fluidOuncesToMl(3), 'fl_oz')).toBeCloseTo(3, 6)
    expect(inputToMl(100, 'ml')).toBe(100)
    expect(inputToMl(3, 'fl_oz')).toBeCloseTo(fluidOuncesToMl(3), 6)
  })
})

describe('toRoundedString', () => {
  it('trims trailing zeros to two decimals', () => {
    expect(toRoundedString(1.5)).toBe('1.5')
    expect(toRoundedString(2)).toBe('2')
    expect(toRoundedString(3.14159)).toBe('3.14')
  })

  it('returns "0" for non-finite values', () => {
    expect(toRoundedString(Number.POSITIVE_INFINITY)).toBe('0')
    expect(toRoundedString(Number.NaN)).toBe('0')
  })
})

describe('formatDisplayWeight', () => {
  it('keeps grams below 1 kg, promotes to kg at/above 1000 g', () => {
    expect(formatDisplayWeight(500, 'g')).toBe('500 g')
    expect(formatDisplayWeight(1500, 'g')).toBe('1.5 kg')
    expect(formatDisplayWeight(1000, 'g')).toBe('1 kg')
  })

  it('keeps ounces below 16 oz, promotes to lb at/above 16 oz', () => {
    // 100 g ≈ 3.53 oz
    expect(formatDisplayWeight(100, 'oz')).toBe('3.53 oz')
    // 500 g ≈ 17.6 oz → 1.1 lb
    expect(formatDisplayWeight(500, 'oz')).toBe('1.1 lb')
  })
})
