package loggingmanager

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/query"
)

func (core *Core) acknowledgeEvent(ctx context.Context, input dto.AcknowledgeEventInput) error {
	lifecycle, err := core.loadLifecycle(ctx, input.EventID)
	if err != nil {
		return err
	}

	ackAt := input.AcknowledgedAt
	if ackAt == "" {
		ackAt = time.Now().UTC().Format(time.RFC3339)
	}

	lifecycle.Acknowledged = true
	lifecycle.AcknowledgedBy = input.AcknowledgedBy
	lifecycle.AcknowledgedAt = ackAt

	core.writerInputCh <- WriteRequest{Query: query.UpsertLifecycle(lifecycle)}
	return nil
}

func (core *Core) muteEvent(ctx context.Context, input dto.MuteEventInput) error {
	lifecycle, err := core.loadLifecycle(ctx, input.EventID)
	if err != nil {
		return err
	}

	mutedAt := input.MutedAt
	if mutedAt == "" {
		mutedAt = time.Now().UTC().Format(time.RFC3339)
	}

	lifecycle.Muted = true
	lifecycle.MutedBy = input.MutedBy
	lifecycle.MutedAt = mutedAt

	core.writerInputCh <- WriteRequest{Query: query.UpsertLifecycle(lifecycle)}
	return nil
}

func (core *Core) loadLifecycle(ctx context.Context, eventID string) (dto.LifecycleRow, error) {
	var (
		row             dto.LifecycleRow
		acknowledgedInt int
		mutedInt        int
	)

	builtQuery := query.SelectLifecycleByEventID(eventID)
	err := core.db.QueryRowContext(ctx, builtQuery.SQL, builtQuery.Args...).Scan(
		&row.RecID,
		&row.EventID,
		&row.EventRecID,
		&acknowledgedInt,
		&row.AcknowledgedBy,
		&row.AcknowledgedAt,
		&mutedInt,
		&row.MutedBy,
		&row.MutedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return dto.LifecycleRow{}, errEventNotFound
	}
	if err != nil {
		return dto.LifecycleRow{}, err
	}

	row.Acknowledged = acknowledgedInt == 1
	row.Muted = mutedInt == 1
	return row, nil
}
