package query

import (
	"errors"
	"fmt"
)

const (
	// EventIDMaxLength defines the maximum supported event_id length.
	EventIDMaxLength = 20
	// EventIDPrefixMaxLength defines the maximum supported ID prefix length.
	EventIDPrefixMaxLength = 10
	// EventIDMaxSequence defines the inclusive upper bound for generated IDs.
	EventIDMaxSequence = 99999999
)

var (
	errEmptyEventIDPrefix      = errors.New("event id prefix is required")
	errEventIDPrefixTooLong    = errors.New("event id prefix exceeds maximum length")
	errEventIDSequenceOverflow = errors.New("event id sequence exceeds maximum value")
	errGeneratedEventIDTooLong = errors.New("generated event id exceeds maximum length")
)

// BuildEventID builds a generated event identifier in "<prefix>_<seq>" format.
//
// Contract:
//   - prefix must be non-empty and at most 10 characters.
//   - seq must be in range [0, 99999999].
//   - resulting identifier length must not exceed 20 characters.
//
// Side effects:
//   - None. This function performs pure in-memory formatting.
func BuildEventID(prefix string, seq uint32) (string, error) {
	if prefix == "" {
		return "", errEmptyEventIDPrefix
	}
	if len(prefix) > EventIDPrefixMaxLength {
		return "", errEventIDPrefixTooLong
	}
	if seq > EventIDMaxSequence {
		return "", errEventIDSequenceOverflow
	}

	eventID := fmt.Sprintf("%s_%d", prefix, seq)
	if len(eventID) > EventIDMaxLength {
		return "", errGeneratedEventIDTooLong
	}

	return eventID, nil
}
