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

func TestPersonGetAgeFactor(t *testing.T) {
	t.Parallel()

	withAge := func(years int) *Person {
		d := time.Now().AddDate(-years, 0, -1)
		return &Person{Birthdate: &d}
	}
	cases := []struct {
		years int
		want  float64
	}{{3, 0.75}, {7, 0.75}, {11, 0.85}, {14, 0.95}, {17, 1.00}, {30, 1.10}, {60, 0.90}}
	for _, c := range cases {
		if got := withAge(c.years).getAgeFactor(); got != c.want {
			t.Errorf("age %d: got %v, want %v", c.years, got, c.want)
		}
	}
	if got := (&Person{}).getAgeFactor(); got != 1.10 {
		t.Errorf("nil birthdate: got %v, want 1.10", got)
	}
}

func TestPersonGetGenderAndConditioningFactors(t *testing.T) {
	t.Parallel()

	male, female, other := GenderMale, GenderFemale, Gender("other")
	if (&Person{Gender: &male}).getGenderFactor() != 1.05 ||
		(&Person{Gender: &female}).getGenderFactor() != 0.95 ||
		(&Person{Gender: &other}).getGenderFactor() != 1.00 ||
		(&Person{}).getGenderFactor() != 1.00 {
		t.Error("gender factor mismatch")
	}

	sed, avg := ConditioningLevelSedentary, ConditioningLevelAverage
	ath, mil, unknown := ConditioningLevelAthletic, ConditioningLevelMilitary, ConditioningLevel("unknown")
	if (&Person{ConditioningLevel: &sed}).getConditioningFactor() != 0.85 ||
		(&Person{ConditioningLevel: &avg}).getConditioningFactor() != 1.00 ||
		(&Person{ConditioningLevel: &ath}).getConditioningFactor() != 1.15 ||
		(&Person{ConditioningLevel: &mil}).getConditioningFactor() != 1.20 ||
		(&Person{ConditioningLevel: &unknown}).getConditioningFactor() != 1.00 ||
		(&Person{}).getConditioningFactor() != 1.00 {
		t.Error("conditioning factor mismatch")
	}
}

func TestValidationErrorMessage(t *testing.T) {
	t.Parallel()
	if got := (ValidationError{Message: "bad"}).Error(); got != "validation error: bad" {
		t.Errorf("unexpected message: %q", got)
	}
}
