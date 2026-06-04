package services

import (
	"context"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

func newAlertsListClientStub(
	t *testing.T,
	items []loggingmanagercontract.ListAlertItem,
	totalCount int,
) *alertsClientStub {
	t.Helper()

	return &alertsClientStub{
		requestFunc: func(_ context.Context, _ string, _ any, response any) error {
			typed, ok := response.(*loggingmanagercontract.ListAlertsResponse)
			if !ok {
				t.Fatalf("response type = %T, want *ListAlertsResponse", response)
			}
			typed.Items = items
			typed.TotalCount = totalCount
			return nil
		},
	}
}

func newAlertsActionClientStub(t *testing.T, okResponse bool) *alertsClientStub {
	t.Helper()

	return &alertsClientStub{
		requestFunc: func(_ context.Context, _ string, _ any, response any) error {
			typed, ok := response.(*loggingmanagercontract.OKResponse)
			if !ok {
				t.Fatalf("response type = %T, want *OKResponse", response)
			}
			typed.OK = okResponse
			return nil
		},
	}
}

func assertAlertsListRequest(t *testing.T, request any, wantToken string, wantPage int, wantPageSize int) {
	t.Helper()

	typed, ok := request.(loggingmanagercontract.ListAlertsInput)
	if !ok {
		t.Fatalf("request type = %T, want ListAlertsInput", request)
	}
	if typed.AccessToken != wantToken || typed.Page != wantPage || typed.PageSize != wantPageSize {
		t.Fatalf("request = %#v, want access token/page/page_size forwarded", typed)
	}
}

func assertAcknowledgeRequest(t *testing.T, request any, wantToken string, wantID string, wantActor string) {
	t.Helper()

	typed, ok := request.(loggingmanagercontract.AcknowledgeAlertInput)
	if !ok {
		t.Fatalf("request type = %T, want AcknowledgeAlertInput", request)
	}
	if typed.AccessToken != wantToken || typed.EventID != wantID || typed.AcknowledgedBy != wantActor {
		t.Fatalf("request = %#v, want forwarded access token, id, and actor", typed)
	}
}
