package services

import (
	"context"
	"reflect"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
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

func assertAlertsListRequest(
	t *testing.T,
	request any,
	wantToken string,
	wantPage int,
	wantPageSize int,
	wantFilters []loggingmanagerdto.Filter,
) {
	t.Helper()

	typed, ok := request.(loggingmanagercontract.ListAlertsInput)
	if !ok {
		t.Fatalf("request type = %T, want ListAlertsInput", request)
	}
	if typed.AccessToken != wantToken || typed.Page != wantPage || typed.PageSize != wantPageSize {
		t.Fatalf("request = %#v, want access token/page/page_size forwarded", typed)
	}
	if !reflect.DeepEqual(typed.Filters, wantFilters) {
		t.Fatalf("filters = %#v, want %#v", typed.Filters, wantFilters)
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

func assertAlertsListSubject(t *testing.T, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("subject = %q, want %q", got, want)
	}
}

func assertAlertsListResult(
	t *testing.T,
	name string,
	got AlertListPage,
	wantItems []loggingmanagercontract.ListAlertItem,
	wantTotalCount int,
) {
	t.Helper()

	if !reflect.DeepEqual(got.Items, wantItems) || got.TotalCount != wantTotalCount {
		t.Fatalf("%s = %#v, want items=%#v total_count=%d", name, got, wantItems, wantTotalCount)
	}
}

func runAlertsServiceListTest(
	t *testing.T,
	name string,
	invoke func(AlertsService) (AlertListPage, error),
	wantSubject string,
	wantPage int,
	wantSize int,
	wantFilters []loggingmanagerdto.Filter,
	wantItems []loggingmanagercontract.ListAlertItem,
	wantTotalCount int,
) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		client := newAlertsListClientStub(t, wantItems, wantTotalCount)
		service := NewSystemAlertsService(client)

		got, err := invoke(service)
		if err != nil {
			t.Fatalf("%s error = %v", name, err)
		}

		assertAlertsListSubject(t, client.subject, wantSubject)
		assertAlertsListRequest(t, client.request, "AT", wantPage, wantSize, wantFilters)
		assertAlertsListResult(t, name, got, wantItems, wantTotalCount)
	})
}
