package dto

// EventMetaRow models one row in the event_meta table.
type EventMetaRow struct {
	RecID     int64
	EventID   string
	MetaKey   string
	MetaValue string
}
