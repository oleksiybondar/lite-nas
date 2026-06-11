package dto

// LifecycleRow models one row in the lifecycle table.
type LifecycleRow struct {
	RecID          int64
	EventID        string
	EventRecID     int64
	Acknowledged   bool
	AcknowledgedBy string
	AcknowledgedAt string
	Muted          bool
	MutedBy        string
	MutedAt        string
}
