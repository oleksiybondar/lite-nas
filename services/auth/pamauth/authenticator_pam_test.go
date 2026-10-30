package pamauth

import (
	"testing"

	pam "github.com/msteinert/pam/v2"
)

// TestAccountErrorsOnlyRequirePasswordChangeForThePAMExpiryStatus keeps
// unrelated account failures from being offered a password update.
func TestAccountErrorsOnlyRequirePasswordChangeForThePAMExpiryStatus(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want OutcomeCode
	}{
		{name: "password expired", err: pam.ErrNewAuthtokReqd, want: OutcomePasswordChangeNeeded},
		{name: "account expired", err: pam.ErrAcctExpired, want: OutcomeDenied},
		{name: "permission denied", err: pam.ErrPermDenied, want: OutcomeDenied},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := newAuthenticateConversation(AuthenticateRequest{}).resultForAccountError("alice", test.err)
			if result.Code != test.want {
				t.Fatalf("Code = %q, want %q", result.Code, test.want)
			}
		})
	}
}
