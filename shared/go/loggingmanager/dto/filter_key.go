package dto

// FilterKey defines the supported filter keys for event listing queries.
type FilterKey string

const (
	FilterKeyCategory     FilterKey = "category"
	FilterKeyState        FilterKey = "state"
	FilterKeyAcknowledged FilterKey = "acknowledged"
	FilterKeyMuted        FilterKey = "muted"
	FilterKeyCreatedAt    FilterKey = "created_at"
)
