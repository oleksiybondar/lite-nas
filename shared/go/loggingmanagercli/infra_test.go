package loggingmanagercli

import (
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

func TestWithAccessTokenInjectsAcknowledgePayload(t *testing.T) {
	t.Parallel()

	payload := loggingmanagercontract.AcknowledgeAlertInput{
		EventID:        "evt-1",
		AcknowledgedBy: "operator",
	}

	output := withAccessToken(payload, "token-ack")
	typed, ok := output.(loggingmanagercontract.AcknowledgeAlertInput)
	if !ok {
		t.Fatalf("output type = %T, want AcknowledgeAlertInput", output)
	}
	if typed.AccessToken != "token-ack" {
		t.Fatalf("AccessToken = %q, want token-ack", typed.AccessToken)
	}
}

func TestWithAccessTokenInjectsListPayload(t *testing.T) {
	t.Parallel()

	payload := loggingmanagercontract.ListAlertsInput{Page: 1}

	output := withAccessToken(payload, "token-read")
	typed, ok := output.(loggingmanagercontract.ListAlertsInput)
	if !ok {
		t.Fatalf("output type = %T, want ListAlertsInput", output)
	}
	if typed.AccessToken != "token-read" {
		t.Fatalf("AccessToken = %q, want token-read", typed.AccessToken)
	}
}

func TestWithAccessTokenLeavesUnknownPayloadUntouched(t *testing.T) {
	t.Parallel()

	payload := struct{ Name string }{Name: "unchanged"}

	output := withAccessToken(payload, "token")
	typed, ok := output.(struct{ Name string })
	if !ok {
		t.Fatalf("output type = %T, want anonymous struct", output)
	}
	if typed.Name != "unchanged" {
		t.Fatalf("Name = %q, want unchanged", typed.Name)
	}
}
