package metrics

import "time"

// SensorNumericReading is one measured value with a stable source key and unit.
type SensorNumericReading struct {
	Key   string  `json:"key" validate:"required,max=256"`
	Name  string  `json:"name" validate:"required,max=256"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit" validate:"required,max=16"`
}

// SensorStateReading is one current or latched sensor condition.
type SensorStateReading struct {
	Key    string `json:"key" validate:"required,max=256"`
	Name   string `json:"name" validate:"required,max=256"`
	Active bool   `json:"active"`
}

// SensorSnapshot contains all readings obtained during one polling cycle.
// An empty section means no readings in that category were available.
type SensorSnapshot struct {
	Timestamp    time.Time              `json:"timestamp" validate:"required"`
	Temperatures []SensorNumericReading `json:"temperatures" validate:"dive"`
	Voltages     []SensorNumericReading `json:"voltages" validate:"dive"`
	Fans         []SensorNumericReading `json:"fans" validate:"dive"`
	Clocks       []SensorNumericReading `json:"clocks" validate:"dive"`
	Throttling   []SensorStateReading   `json:"throttling" validate:"dive"`
}
