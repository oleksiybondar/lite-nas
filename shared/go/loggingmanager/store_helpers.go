package loggingmanager

import (
	"context"
	"database/sql"
	"errors"

	"lite-nas/shared/loggingmanager/query"
)

var errEventNotFound = errors.New("loggingmanager event not found")

func (core *Core) loadEventRecID(ctx context.Context, eventID string) (int64, error) {
	builtQuery := query.SelectEventRecIDByEventID(eventID)

	var recID int64
	err := core.db.QueryRowContext(ctx, builtQuery.SQL, builtQuery.Args...).Scan(&recID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errEventNotFound
	}
	if err != nil {
		return 0, err
	}
	return recID, nil
}
