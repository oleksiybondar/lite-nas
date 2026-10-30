package backuprestore

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/authtoken"
	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/messaging"
)

// Requirements: backup-restore/SR-003.
func TestDecodeRequestRequiresAdministratorRole(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := authtoken.NewIssuer(authtoken.IssuerOptions{Issuer: "test", Audience: "test", AccessLifetime: time.Minute}, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: "test", Audience: "test"}, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled())}
	cases := []struct {
		name  string
		login string
		roles []string
		allow bool
	}{
		{name: "admin", roles: []string{"admin"}, allow: true},
		{name: "sudo", roles: []string{"sudo"}, allow: true},
		{name: "dedicated CLI identity", login: "lite-nas-backup-restore-cli", allow: true},
		{name: "security officer", roles: []string{"lite-nas-security"}, allow: false},
		{name: "operator", roles: []string{"lite-nas-operator"}, allow: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			login := testCase.login
			if login == "" {
				login = "test"
			}
			token, _, err := issuer.Issue(authtoken.Principal{Subject: "123", Login: login, Roles: testCase.roles})
			if err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(contract.ApplicationsRequest{AccessToken: token})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeRequest[contract.ApplicationsRequest](runtime, messaging.Envelope{Payload: payload})
			if (err == nil) != testCase.allow {
				t.Fatalf("authorization error = %v, allow = %v", err, testCase.allow)
			}
		})
	}
}
