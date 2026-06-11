package dto

// FilterCondition defines allowed filter operators.
type FilterCondition string

const (
	FilterConditionEQ      FilterCondition = "eq"
	FilterConditionIN      FilterCondition = "in"
	FilterConditionBetween FilterCondition = "between"
)
