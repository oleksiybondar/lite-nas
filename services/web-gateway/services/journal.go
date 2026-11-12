package services

import (
	"context"
	"errors"
	"time"

	contract "lite-nas/shared/contracts/journalmanager"
	"lite-nas/shared/messaging"
)

// JournalService forwards authenticated journal queries, unit discovery, and vacuum requests.
type JournalService interface {
	Query(context.Context, contract.QueryRequest) (contract.QueryResponse, error)
	Units(context.Context, contract.UnitsRequest) (contract.UnitsResponse, error)
	Vacuum(context.Context, contract.VacuumRequest) error
}

// journalService owns the NATS transport for journal operations.
type journalService struct{ client messaging.Client }

// NewJournalService binds journal requests to the shared messaging client.
func NewJournalService(client messaging.Client) JournalService { return journalService{client: client} }

// Query returns one bounded chunk of journal entries.
func (s journalService) Query(ctx context.Context, request contract.QueryRequest) (contract.QueryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var response contract.QueryResponse
	err := s.client.Request(ctx, contract.QuerySubject, request, &response)
	return response, err
}

// Units returns service units with entries matching the supplied filters.
func (s journalService) Units(ctx context.Context, request contract.UnitsRequest) (contract.UnitsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var response contract.UnitsResponse
	err := s.client.Request(ctx, contract.UnitsSubject, request, &response)
	return response, err
}

// Vacuum executes the confirmed retention action and checks completion.
func (s journalService) Vacuum(ctx context.Context, request contract.VacuumRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var response contract.VacuumResponse
	if err := s.client.Request(ctx, contract.VacuumSubject, request, &response); err != nil {
		return err
	}
	if !response.Completed {
		return errors.New("journal vacuum was not completed")
	}
	return nil
}
