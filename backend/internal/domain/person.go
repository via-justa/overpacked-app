package domain

import (
	"time"

	"github.com/google/uuid"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type ConditioningLevel string

const (
	ConditioningLevelSedentary ConditioningLevel = "sedentary"
	ConditioningLevelAverage   ConditioningLevel = "average"
	ConditioningLevelAthletic  ConditioningLevel = "athletic"
	ConditioningLevelMilitary  ConditioningLevel = "military"
)

type Person struct {
	ID                uuid.UUID
	Name              string
	Gender            *Gender
	Birthdate         *time.Time
	BodyWeightGrams   *float64
	ConditioningLevel *ConditioningLevel
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CalculateRecommendedMaxWeightGrams calculates the recommended maximum backpack
// weight based on age, gender, and conditioning level using a scientific formula:
// Recommended Load = Body Weight × 0.12 × F_age × F_gender × F_conditioning
//
// Returns the recommended weight in grams, or 0 if required fields are missing.
func (p *Person) CalculateRecommendedMaxWeightGrams() float64 {
	if p.BodyWeightGrams == nil || *p.BodyWeightGrams <= 0 {
		return 0
	}

	bodyWeightKg := *p.BodyWeightGrams / 1000.0
	baselinePercent := 0.12

	// Calculate age factor based on birthdate
	ageFactor := p.getAgeFactor()

	// Calculate gender factor
	genderFactor := p.getGenderFactor()

	// Calculate conditioning factor
	conditioningFactor := p.getConditioningFactor()

	// Formula: bodyWeightKg × 0.12 × ageFactor × genderFactor × conditioningFactor
	recommendedWeightKg := bodyWeightKg * baselinePercent * ageFactor * genderFactor * conditioningFactor

	// Convert back to grams
	return recommendedWeightKg * 1000
}

func (p *Person) getAgeFactor() float64 {
	if p.Birthdate == nil {
		return 1.10 // Default to adult peak (19-50)
	}

	age := calculateAge(*p.Birthdate)

	switch {
	case age < 5:
		return 0.75 // Very young children
	case age < 9:
		return 0.75 // Ages 5-8
	case age < 13:
		return 0.85 // Ages 9-12
	case age < 16:
		return 0.95 // Ages 13-15
	case age < 19:
		return 1.00 // Ages 16-18
	case age <= 50:
		return 1.10 // Ages 19-50 (peak)
	default:
		return 0.90 // Ages 50+
	}
}

func (p *Person) getGenderFactor() float64 {
	if p.Gender == nil {
		return 1.00 // Default neutral
	}

	switch *p.Gender {
	case GenderFemale:
		return 0.95
	case GenderMale:
		return 1.05
	default:
		return 1.00
	}
}

func (p *Person) getConditioningFactor() float64 {
	if p.ConditioningLevel == nil {
		return 1.00 // Default to average
	}

	switch *p.ConditioningLevel {
	case ConditioningLevelSedentary:
		return 0.85
	case ConditioningLevelAverage:
		return 1.00
	case ConditioningLevelAthletic:
		return 1.15
	case ConditioningLevelMilitary:
		return 1.20
	default:
		return 1.00
	}
}

func calculateAge(birthdate time.Time) int {
	today := time.Now()
	age := today.Year() - birthdate.Year()
	// Adjust if birthday hasn't occurred this year
	if today.Month() < birthdate.Month() || (today.Month() == birthdate.Month() && today.Day() < birthdate.Day()) {
		age--
	}
	return age
}
