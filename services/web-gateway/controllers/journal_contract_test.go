package controllers

import (
	"context"
	"testing"

	"lite-nas/services/web-gateway/dto/journal"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/shared/authtoken"
	contract "lite-nas/shared/contracts/journalmanager"
)

type journalContractStub struct{ token string }

func (s *journalContractStub) Query(_ context.Context, request contract.QueryRequest) (contract.QueryResponse, error) {
	s.token = request.AccessToken
	return contract.QueryResponse{}, nil
}

func (s *journalContractStub) Units(_ context.Context, request contract.UnitsRequest) (contract.UnitsResponse, error) {
	s.token = request.AccessToken
	return contract.UnitsResponse{}, nil
}

func (s *journalContractStub) Vacuum(_ context.Context, request contract.VacuumRequest) error {
	s.token = request.AccessToken
	return nil
}

// TestJournalControllersForwardAccessToken verifies journal reads and retention use the authenticated credential.
func TestJournalControllersForwardAccessToken(t *testing.T) {
	ctx := middlewares.NewAuthenticatedContext(context.Background(), "caller-token", authtoken.AccessClaims{})
	cases := []struct {
		name string
		call func(JournalController) error
	}{
		{"query", func(c JournalController) error {
			_, err := c.Query(ctx, &journal.QueryInput{Priority: "3"})
			return err
		}},
		{"units", func(c JournalController) error {
			_, err := c.Units(ctx, &journal.UnitsInput{Priority: "3"})
			return err
		}},
		{"vacuum", func(c JournalController) error { _, err := c.Vacuum(ctx, &journal.VacuumInput{}); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &journalContractStub{}
			mustRunControllerAction(t, func() error { return tc.call(NewJournalController(stub)) })
			if stub.token != "caller-token" {
				t.Fatalf("forwarded token = %q", stub.token)
			}
		})
	}
}
