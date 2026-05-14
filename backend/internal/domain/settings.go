package domain

type WeightUnit string

const (
	WeightUnitGrams  WeightUnit = "g"
	WeightUnitOunces WeightUnit = "oz"
)

type DistanceUnit string

const (
	DistanceUnitKilometers DistanceUnit = "km"
	DistanceUnitMiles      DistanceUnit = "mi"
)

type TemperatureUnit string

const (
	TemperatureUnitCelsius    TemperatureUnit = "c"
	TemperatureUnitFahrenheit TemperatureUnit = "f"
)

type VolumeUnit string

const (
	VolumeUnitMilliliters VolumeUnit = "ml"
	VolumeUnitFluidOunces VolumeUnit = "fl_oz"
)

type Currency string

const (
	CurrencyUSD Currency = "usd"
	CurrencyEUR Currency = "eur"
)

type Settings struct {
	ID              int
	WeightUnit      WeightUnit
	DistanceUnit    DistanceUnit
	TemperatureUnit TemperatureUnit
	VolumeUnit      VolumeUnit
	Currency        Currency
}
