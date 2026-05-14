package domain

import (
	"math"
	"testing"
	"time"
)

func TestPersonCalculateRecommendedMaxWeightGramsReturnsZeroWithoutBodyWeight(t *testing.T) {
	t.Parallel()

	person := &Person{}

	if got := person.CalculateRecommendedMaxWeightGrams(); got != 0 {
		t.Fatalf("expected 0 without body weight, got %v", got)
	}
}

func TestPersonCalculateRecommendedMaxWeightGramsAdultMaleAthletic(t *testing.T) {
	t.Parallel()

	bodyWeightGrams := 70000.0
	gender := GenderMale
	conditioning := ConditioningLevelAthletic
	birthdate := time.Now().AddDate(-30, -1, 0)

	person := &Person{
		BodyWeightGrams:   &bodyWeightGrams,
		Gender:            &gender,
		Birthdate:         &birthdate,
		ConditioningLevel: &conditioning,
	}

	got := person.CalculateRecommendedMaxWeightGrams()
	want := 70000.0 * 0.12 * 1.10 * 1.05 * 1.15

	if math.Abs(got-want) > 0.01 {
		t.Fatalf("expected %.4f, got %.4f", want, got)
	}
}

func TestPersonCalculateRecommendedMaxWeightGramsChildFemaleSedentary(t *testing.T) {
	t.Parallel()

	bodyWeightGrams := 35000.0
	gender := GenderFemale
	conditioning := ConditioningLevelSedentary
	birthdate := time.Now().AddDate(-10, -1, 0)

	person := &Person{
		BodyWeightGrams:   &bodyWeightGrams,
		Gender:            &gender,
		Birthdate:         &birthdate,
		ConditioningLevel: &conditioning,
	}

	got := person.CalculateRecommendedMaxWeightGrams()
	want := 35000.0 * 0.12 * 0.85 * 0.95 * 0.85

	if math.Abs(got-want) > 0.01 {
		t.Fatalf("expected %.4f, got %.4f", want, got)
	}
}
