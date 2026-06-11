package dto

// ListEventsInput defines paginated list input with optional filters.
type ListEventsInput struct {
	Page     int      `json:"page" validate:"required,gte=1"`
	PageSize int      `json:"page_size" validate:"omitempty,gte=1,lte=500"`
	Filters  []Filter `json:"filters,omitempty" validate:"omitempty,dive"`
}
