package loggingmanager

import (
	"testing"

	"lite-nas/shared/loggingmanager/enum"
)

type unmarshalTestCase struct {
	name        string
	payload     []byte
	wantEventID string
	wantStatus  enum.Status
}

func TestListAlertsResponseUnmarshalJSONSupportsLegacyAndFlatItems(t *testing.T) {
	t.Parallel()

	for _, testCase := range listAlertsUnmarshalTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertUnmarshaledItem(t, testCase)
		})
	}
}

func listAlertsUnmarshalTestCases() []unmarshalTestCase {
	return []unmarshalTestCase{
		{
			name:        "legacy nested items",
			payload:     legacyNestedItemsPayload(),
			wantEventID: "event_1",
			wantStatus:  enum.StatusActive,
		},
		{
			name:        "flat items",
			payload:     flatItemsPayload(),
			wantEventID: "event_1",
			wantStatus:  enum.StatusActive,
		},
	}
}

func legacyNestedItemsPayload() []byte {
	return []byte(`{
  "items": [
    {
      "Event": {
        "RecID": 1,
        "EventID": "event_1",
        "Category": "disk_health",
        "Severity": "warning",
        "Priority": 2,
        "CreatedAt": "2026-05-12T14:30:00Z",
        "Source": "raid-monitor"
      },
      "Lifecycle": {
        "RecID": 1,
        "EventID": "event_1",
        "EventRecID": 1,
        "Acknowledged": false,
        "AcknowledgedBy": "",
        "AcknowledgedAt": "2026-05-12T14:30:00Z",
        "Muted": false,
        "MutedBy": "",
        "MutedAt": "2026-05-12T14:30:00Z"
      },
      "State": {
        "RecID": 1,
        "EventID": "event_1",
        "EventRecID": 1,
        "Status": "active",
        "Message": ""
      },
      "LastValue": null,
      "Meta": null
    }
  ]
}`)
}

func flatItemsPayload() []byte {
	return []byte(`{
  "items": [
    {
      "RecID": 1,
      "EventID": "event_1",
      "Category": "disk_health",
      "Severity": "warning",
      "Priority": 2,
      "CreatedAt": "2026-05-12T14:30:00Z",
      "Source": "raid-monitor",
      "EventRecID": 1,
      "Acknowledged": false,
      "AcknowledgedBy": "",
      "AcknowledgedAt": "2026-05-12T14:30:00Z",
      "Muted": false,
      "MutedBy": "",
      "MutedAt": "2026-05-12T14:30:00Z",
      "Status": "active",
      "Message": ""
    }
  ]
}`)
}

func assertUnmarshaledItem(t *testing.T, testCase unmarshalTestCase) {
	t.Helper()

	item := mustUnmarshalSingleItem(t, testCase.payload)
	if item.EventID != testCase.wantEventID {
		t.Fatalf("event_id = %q, want %q", item.EventID, testCase.wantEventID)
	}
	if item.Status != testCase.wantStatus {
		t.Fatalf("status = %q, want %q", item.Status, testCase.wantStatus)
	}
}

func mustUnmarshalSingleItem(t *testing.T, payload []byte) ListAlertItem {
	t.Helper()

	var response ListAlertsResponse
	if err := response.UnmarshalJSON(payload); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if len(response.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(response.Items))
	}

	return response.Items[0]
}
