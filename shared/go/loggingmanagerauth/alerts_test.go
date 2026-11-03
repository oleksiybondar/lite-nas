package loggingmanagerauth

import (
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

// TestAlertTokenInjectionSupportsSingleAlertReads keeps authenticated manager
// lookups available to service clients that use the alert access-token wrapper.
func TestAlertTokenInjectionSupportsSingleAlertReads(t *testing.T) {
	t.Parallel()
	payload := InjectAlertAccessToken(
		loggingmanagercontract.GetAlertInput{EventID: "SEC123456_12345678"},
		"service-token",
	)
	input, ok := payload.(loggingmanagercontract.GetAlertInput)
	if !ok || input.AccessToken != "service-token" {
		t.Fatalf("injected read payload = %#v", payload)
	}
}
