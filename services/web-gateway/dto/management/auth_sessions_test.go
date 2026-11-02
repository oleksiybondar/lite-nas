package management

import (
	"testing"

	authcontract "lite-nas/shared/contracts/auth"
)

// TestAuthSessionsBodyPreservesSource covers the browser inventory mapping.
func TestAuthSessionsBodyPreservesSource(t *testing.T) {
	got := NewAuthSessionsBody([]authcontract.SessionSummary{{SessionID: "session_1", Source: "moments"}})
	if got.Data[0].Source != "moments" {
		t.Fatalf("source = %q, want moments", got.Data[0].Source)
	}
}
