package dto

// Filter defines one validated filter clause for event listing queries.
type Filter struct {
	Key       FilterKey       `json:"key" validate:"required,oneof=category state acknowledged muted created_at"`
	Condition FilterCondition `json:"condition" validate:"required,oneof=eq in between"`
	Values    []string        `json:"values" validate:"required,min=1,dive,required,max=256"`
}
