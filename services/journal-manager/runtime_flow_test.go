package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	contract "lite-nas/shared/contracts/journalmanager"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/hostcommandtest"
	"lite-nas/shared/testutil/hostmanagertest"
	"lite-nas/shared/testutil/messagingtest"
)

// recordingAlertClient records one journal security alert or a publication failure.
type recordingAlertClient struct {
	subject string
	payload loggingcontract.AlertPayload
	err     error
}

// Publish records the alert effect produced after vacuum succeeds.
func (c *recordingAlertClient) Publish(_ context.Context, subject string, payload any) error {
	c.subject = subject
	c.payload = payload.(loggingcontract.AlertPayload)
	return c.err
}

// Request is unused by the journal manager's alert publisher.
func (*recordingAlertClient) Request(context.Context, string, any, any) error { return nil }

// Drain is unused by the journal manager's alert publisher.
func (*recordingAlertClient) Drain() error { return nil }

// Close is unused by the journal manager's alert publisher.
func (*recordingAlertClient) Close() {}

// runTestManager starts the service with temporary credentials and in-memory NATS.
func runTestManager(t *testing.T, ctx context.Context, server *messagingtest.RecordingServer, log *hostmanagertest.RecordingLogger) error {
	t.Helper()
	return runWithDependencies(ctx, hostmanagertest.MustConfig(t), hostmanagertest.InfraFactory(server, log), &hostcommandtest.RecordingRunner{})
}

// Requirements: journal-manager/IR-001.
func TestRunDrainsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := runTestManager(t, ctx, &messagingtest.RecordingServer{}, &hostmanagertest.RecordingLogger{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("run error = %v", err)
	}
}

// Requirements: journal-manager/IR-001.
func TestRunReturnsRegistrationFailure(t *testing.T) {
	want := errors.New("registration failed")
	err := runTestManager(t, context.Background(), &messagingtest.RecordingServer{RegisterErr: want}, &hostmanagertest.RecordingLogger{})
	if !errors.Is(err, want) {
		t.Fatalf("run error = %v", err)
	}
}

// Requirements: journal-manager/IR-001.
func TestRunReportsDrainFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	log := &hostmanagertest.RecordingLogger{}
	_ = runTestManager(t, ctx, &messagingtest.RecordingServer{DrainErr: errors.New("drain failed")}, log)
	if len(log.Errors) != 1 {
		t.Fatalf("error logs = %v", log.Errors)
	}
}

// Requirements: journal-manager/IR-001.
func TestRunRejectsMissingConfig(t *testing.T) {
	if err := runWithDependencies(context.Background(), t.TempDir()+"/missing", nil, nil); err == nil {
		t.Fatal("missing config accepted")
	}
}

// Requirements: journal-manager/FR-001, journal-manager/FR-002, journal-manager/FR-003.
func TestRegisterBindsJournalSubjects(t *testing.T) {
	m, _ := testManager(t, []string{"lite-nas-security"})
	server := &messagingtest.RecordingServer{}
	if err := m.register(server); err != nil {
		t.Fatal(err)
	}
	if len(server.Handlers) != 3 || server.Handlers[contract.QuerySubject] == nil || server.Handlers[contract.UnitsSubject] == nil || server.Handlers[contract.VacuumSubject] == nil {
		t.Fatalf("handlers = %v", server.Handlers)
	}
}

// Requirements: journal-manager/FR-001.
func TestQueryReturnsStructuredChunk(t *testing.T) {
	m, token := testManager(t, []string{"lite-nas-security"})
	m.runner.(*hostcommandtest.RecordingRunner).OutputBytes = []byte(`{"__CURSOR":"s=one","MESSAGE":"ready"}` + "\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	response, err := m.query(context.Background(), messagingtest.MustEnvelope(t, contract.QueryRequest{AccessToken: token, Limit: 50}))
	if err != nil || len(response.(contract.QueryResponse).Entries) != 1 {
		t.Fatalf("response = %#v, %v", response, err)
	}
}

// Requirements: journal-manager/FR-002.
func TestVacuumPublishesSecurityAlert(t *testing.T) {
	m, token := testManager(t, []string{"lite-nas-security"})
	client := &recordingAlertClient{}
	m.client = client
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	_, err := m.vacuum(context.Background(), messagingtest.MustEnvelope(t, contract.VacuumRequest{AccessToken: token, Days: 30}))
	if err != nil || client.subject != "security-alert" || client.payload.EventID == "" {
		t.Fatalf("alert = %#v, %v", client, err)
	}
}

// Requirements: journal-manager/FR-002.
func TestVacuumSurfacesAlertFailure(t *testing.T) {
	m, token := testManager(t, []string{"lite-nas-security"})
	m.client = &recordingAlertClient{err: errors.New("publish failed")}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	_, err := m.vacuum(context.Background(), messagingtest.MustEnvelope(t, contract.VacuumRequest{AccessToken: token, Days: 30}))
	if err == nil {
		t.Fatal("alert publication failure hidden")
	}
}

// Requirements: journal-manager/FR-002, journal-manager/IR-001.
func TestExecuteUsesValidatedVacuumEntrypoint(t *testing.T) {
	m, _ := testManager(t, []string{"lite-nas-security"})
	m.execute(commandJob{ctx: context.Background(), days: 30})
	runner := m.runner.(*hostcommandtest.RecordingRunner)
	want := []string{"-n", vacuumHelper, "vacuum", "30"}
	if runner.Name != "/usr/bin/sudo" || !reflect.DeepEqual(runner.Args, want) {
		t.Fatalf("command = %s %v", runner.Name, runner.Args)
	}
}

// Requirements: journal-manager/SR-002.
func TestDecodeRejectsMalformedRequest(t *testing.T) {
	m, token := testManager(t, []string{"lite-nas-security"})
	for name, payload := range hostmanagertest.MalformedPayloads(token) {
		t.Run(name, func(t *testing.T) {
			_, _, err := decode(m, messaging.Envelope{Payload: payload}, func(v contract.QueryRequest) string { return v.AccessToken })
			if err == nil {
				t.Fatal("malformed request accepted")
			}
		})
	}
}
