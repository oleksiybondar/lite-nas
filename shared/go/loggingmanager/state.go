package loggingmanager

import (
	"context"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/query"
)

func (core *Core) setState(ctx context.Context, input dto.SetStateInput) error {
	recID, err := core.loadEventRecID(ctx, input.EventID)
	if err != nil {
		return err
	}

	message := ""
	if input.Message != nil {
		message = *input.Message
	}

	core.writerInputCh <- WriteRequest{
		Query: query.UpsertEventState(dto.EventStateRow{
			RecID:      recID,
			EventID:    input.EventID,
			EventRecID: recID,
			Status:     input.Status,
			Message:    message,
		}),
	}
	return nil
}
