package main

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	contract "lite-nas/shared/contracts/sessionmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/hostcommandtest"
	"lite-nas/shared/testutil/hostmanagertest"
	"lite-nas/shared/testutil/messagingtest"
)

// runTestManager starts the session manager with a temporary config and fake host dependencies.
func runTestManager(t *testing.T, ctx context.Context, server *messagingtest.RecordingServer, log *hostmanagertest.RecordingLogger) error {
	t.Helper()
	return runWithDependencies(ctx, hostmanagertest.MustConfig(t), hostmanagertest.InfraFactory(server, log), &hostcommandtest.RecordingRunner{})
}

// Requirements: session-manager/IR-001.
func TestRunDrainsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	server := &messagingtest.RecordingServer{}
	err := runTestManager(t, ctx, server, &hostmanagertest.RecordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run error = %v", err)
	}
}

// Requirements: session-manager/IR-001.
func TestRunReturnsRegistrationFailure(t *testing.T) {
	want := errors.New("registration failed")
	server := &messagingtest.RecordingServer{RegisterErr: want}
	err := runTestManager(t, context.Background(), server, &hostmanagertest.RecordingLogger{})
	if !errors.Is(err, want) {
		t.Fatalf("run error = %v", err)
	}
}

// Requirements: session-manager/IR-001.
func TestRunReportsDrainFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	log := &hostmanagertest.RecordingLogger{}
	server := &messagingtest.RecordingServer{DrainErr: errors.New("drain failed")}
	_ = runTestManager(t, ctx, server, log)
	if len(log.Errors) != 1 {
		t.Fatalf("error logs = %v", log.Errors)
	}
}

// Requirements: session-manager/IR-001.
func TestRunRejectsMissingConfig(t *testing.T) {
	err := runWithDependencies(context.Background(), t.TempDir()+"/missing.conf", nil, nil)
	if err == nil {
		t.Fatal("missing config accepted")
	}
}

// Requirements: session-manager/FR-001, session-manager/FR-002.
func TestRegisterBindsAllRPCSubjects(t *testing.T) {
	m, _ := newTestManager(t, []string{"lite-nas-security"})
	server := &messagingtest.RecordingServer{}
	if err := m.register(server); err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(server.Handlers))
	for subject := range server.Handlers {
		got = append(got, subject)
	}
	sort.Strings(got)
	want := []string{contract.ListSubject, contract.TerminateSubject, contract.LockSubject, contract.UnlockSubject}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("subjects = %v", got)
	}
}

// Requirements: session-manager/FR-001.
func TestListReturnsCurrentSessions(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-security"})
	m.runner.(*hostcommandtest.RecordingRunner).OutputBytes = []byte("42 1000 alice seat0 1234 user tty2 no -\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	response, err := m.list(context.Background(), messagingtest.MustEnvelope(t, contract.ListRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	sessions := response.(contract.ListResponse).Sessions
	if len(sessions) != 1 || sessions[0].UID != "1000" {
		t.Fatalf("sessions = %#v", sessions)
	}
}

// Requirements: session-manager/SR-002.
func TestDecodeRejectsMalformedRequests(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-security"})
	for name, payload := range hostmanagertest.MalformedPayloads(token) {
		t.Run(name, func(t *testing.T) {
			_, err := decode(m, messaging.Envelope{Payload: payload}, func(request contract.ListRequest) string { return request.AccessToken })
			if err == nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
}

// Requirements: session-manager/FR-002.
func TestActionReturnsCommandFailure(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-security"})
	m.runner.(*hostcommandtest.RecordingRunner).Results = []hostcommandtest.Result{{Err: errors.New("loginctl failed")}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	_, err := m.action(context.Background(), messagingtest.MustEnvelope(t, contract.ActionRequest{AccessToken: token, Session: "42"}), "lock-session")
	if err == nil {
		t.Fatal("command failure was hidden")
	}
}

// Requirements: session-manager/SR-002.
func TestExecuteRejectsOversizedOutput(t *testing.T) {
	m, _ := newTestManager(t, []string{"lite-nas-security"})
	m.runner.(*hostcommandtest.RecordingRunner).OutputBytes = []byte(strings.Repeat("x", 1024*1024+1))
	result := m.execute(commandJob{ctx: context.Background(), action: "list"})
	if result.err == nil {
		t.Fatal("oversized output accepted")
	}
}

// Requirements: session-manager/FR-001.
func TestParseSessionsRejectsMalformedRow(t *testing.T) {
	_, err := parseSessions([]byte("--host=evil 1000 alice\n"))
	if err == nil {
		t.Fatal("malformed row accepted")
	}
}
