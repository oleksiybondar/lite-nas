package messagingtest

import (
	"encoding/json"
	"testing"

	"lite-nas/shared/messaging"
)

// MustEnvelope serializes a request into a transport envelope for boundary tests.
func MustEnvelope(t *testing.T, request any) messaging.Envelope {
	t.Helper()
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.Envelope{Payload: payload}
}
