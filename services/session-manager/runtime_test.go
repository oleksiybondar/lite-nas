package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"

	contract "lite-nas/shared/contracts/sessionmanager"
	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/hostcommandtest"
	"lite-nas/shared/testutil/messagingtest"
)

// newTestManager injects a real token verifier and recording command runner.
func newTestManager(t *testing.T, roles []string) (*manager, string) {
	t.Helper()
	verifier, token := authtokentest.MustIssueAccessToken(t, roles)
	return &manager{verifier: verifier, client: &messagingtest.RecordingClient{AuthUsername: "tester"}, validator: validator.New(validator.WithRequiredStructEnabled()), runner: &hostcommandtest.RecordingRunner{}, jobs: make(chan commandJob, 1), ctx: context.Background()}, token
}

// Requirements: session-manager/SR-001.
func TestListRequiresSecurityOrAdministratorRole(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-operator"})
	_, err := m.list(context.Background(), messagingtest.MustEnvelope(t, contract.ListRequest{AccessToken: token}))
	if err == nil {
		t.Fatal("operator listed sessions")
	}
}

// Requirements: session-manager/SR-002.
func TestActionRejectsOptionShapedSession(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-security"})
	_, err := m.action(context.Background(), messagingtest.MustEnvelope(t, contract.ActionRequest{AccessToken: token, Session: "--host=evil"}), "terminate-session")
	if err == nil {
		t.Fatal("option-shaped session accepted")
	}
}

// Requirements: session-manager/FR-002, session-manager/SR-001.
func TestActionRunsForSecurityOfficer(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-security"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	response, err := m.action(context.Background(), messagingtest.MustEnvelope(t, contract.ActionRequest{AccessToken: token, Session: "42"}), "terminate-session")
	if err != nil || !response.(contract.ActionResponse).Completed {
		t.Fatalf("action = %#v, %v", response, err)
	}
}

// Requirements: session-manager/FR-002, session-manager/IR-001.
func TestExecuteUsesFixedSudoCommand(t *testing.T) {
	m, _ := newTestManager(t, []string{"lite-nas-security"})
	runner := m.runner.(*hostcommandtest.RecordingRunner)
	m.execute(commandJob{ctx: context.Background(), action: "terminate-session", session: "42"})
	want := []string{"-n", "/usr/bin/loginctl", "terminate-session", "--", "42"}
	if runner.Name != "/usr/bin/sudo" || !reflect.DeepEqual(runner.Args, want) {
		t.Fatalf("command = %q %v", runner.Name, runner.Args)
	}
}

// Requirements: session-manager/FR-001.
func TestParseSessionsReturnsOwner(t *testing.T) {
	sessions, err := parseSessions([]byte("42 1000 alice seat0 1234 user tty2 no -\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].ID != "42" || sessions[0].User != "alice" {
		t.Fatalf("sessions = %#v", sessions)
	}
}
