package loggingmanager

import (
	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/query"
)

// WriteRequest envelopes one queued query with optional deferred write-tail
// intents that the writer applies once per flushed transaction.
//
// Tail intents:
//   - RuntimeStateUpdates are de-duplicated by key; only the latest value per
//     key is persisted at flush tail.
//   - TouchesOccurrences indicates the batch touched occurrence inserts and
//     should append max-occurrence cleanup once for the transaction.
type WriteRequest struct {
	Query               query.Query
	RuntimeStateUpdates []dto.RuntimeStateRow
	TouchesOccurrences  bool
}
