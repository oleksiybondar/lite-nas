package main

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	contract "lite-nas/shared/contracts/servicesmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/hostcommandtest"
	"lite-nas/shared/testutil/hostmanagertest"
	"lite-nas/shared/testutil/messagingtest"
)

// Requirements: services-manager/IR-001.
func TestRunDrainsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	server := &messagingtest.RecordingServer{}
	err := runWithDependencies(ctx, hostmanagertest.MustConfig(t), hostmanagertest.InfraFactory(server, &hostmanagertest.RecordingLogger{}), &hostcommandtest.RecordingRunner{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run error = %v", err)
	}
}

// Requirements: services-manager/IR-001.
func TestRunReturnsRegistrationFailure(t *testing.T) {
	want := errors.New("registration failed")
	server := &messagingtest.RecordingServer{RegisterErr: want}
	err := runWithDependencies(context.Background(), hostmanagertest.MustConfig(t), hostmanagertest.InfraFactory(server, &hostmanagertest.RecordingLogger{}), &hostcommandtest.RecordingRunner{})
	if !errors.Is(err, want) {
		t.Fatalf("run error = %v", err)
	}
}

// Requirements: services-manager/IR-001.
func TestRunReportsDrainFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	log := &hostmanagertest.RecordingLogger{}
	server := &messagingtest.RecordingServer{DrainErr: errors.New("drain failed")}
	_ = runWithDependencies(ctx, hostmanagertest.MustConfig(t), hostmanagertest.InfraFactory(server, log), &hostcommandtest.RecordingRunner{})
	if len(log.Errors) != 1 {
		t.Fatalf("error logs = %v", log.Errors)
	}
}

// Requirements: services-manager/IR-001.
func TestRunRejectsMissingConfig(t *testing.T) {
	err := runWithDependencies(context.Background(), t.TempDir()+"/missing.conf", nil, nil)
	if err == nil {
		t.Fatal("missing config accepted")
	}
}

// Requirements: services-manager/FR-001, services-manager/FR-002.
func TestRegisterBindsAllRPCSubjects(t *testing.T) {
	m, _ := newTestManager(t, []string{"admin"})
	server := &messagingtest.RecordingServer{}
	if err := m.register(server); err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(server.Handlers))
	for subject := range server.Handlers {
		got = append(got, subject)
	}
	sort.Strings(got)
	want := []string{contract.DisableSubject, contract.EnableSubject, contract.ListSubject, contract.RestartSubject, contract.StartSubject, contract.StopSubject}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("subjects = %v", got)
	}
}

// Requirements: services-manager/FR-001.
func TestListMergesLoadedAndInstalledServices(t *testing.T) {
	m, token := newTestManager(t, []string{"sudo"})
	runner := m.runner.(*hostcommandtest.RecordingRunner)
	runner.Results = []hostcommandtest.Result{
		{Output: []byte("ssh.service loaded active running OpenSSH server\n")},
		{Output: []byte("ssh.service enabled enabled\ncups.service disabled enabled\n")},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	response, err := m.list(context.Background(), messagingtest.MustEnvelope(t, contract.ListRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	services := response.(contract.ListResponse).Services
	if len(services) != 2 || services[0].Name != "cups.service" || services[1].ActiveState != "active" {
		t.Fatalf("services = %#v", services)
	}
}

// Requirements: services-manager/SR-002.
func TestDecodeRejectsMalformedRequests(t *testing.T) {
	m, token := newTestManager(t, []string{"admin"})
	for name, payload := range hostmanagertest.MalformedPayloads(token) {
		t.Run(name, func(t *testing.T) {
			_, err := decode(m, messaging.Envelope{Payload: payload}, func(request contract.ListRequest) string { return request.AccessToken })
			if err == nil {
				t.Fatal("invalid request accepted")
			}
		})
	}
}

// Requirements: services-manager/IR-001.
func TestActionReturnsCommandFailure(t *testing.T) {
	m, token := newTestManager(t, []string{"admin"})
	m.runner.(*hostcommandtest.RecordingRunner).Results = []hostcommandtest.Result{{Err: errors.New("systemctl failed")}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	_, err := m.action(context.Background(), messagingtest.MustEnvelope(t, contract.ActionRequest{AccessToken: token, Unit: "ssh.service"}), "stop")
	if err == nil {
		t.Fatal("command failure was hidden")
	}
}

// Requirements: services-manager/SR-002.
func TestExecuteRejectsOversizedOutput(t *testing.T) {
	m, _ := newTestManager(t, []string{"admin"})
	m.runner.(*hostcommandtest.RecordingRunner).OutputBytes = []byte(strings.Repeat("x", 1024*1024+1))
	result := m.execute(commandJob{ctx: context.Background(), action: "list-units"})
	if result.err == nil {
		t.Fatal("oversized output accepted")
	}
}

// Requirements: services-manager/FR-001.
func TestParseServicesRejectsMalformedRow(t *testing.T) {
	_, err := parseServices([]byte("bad-unit loaded active running\n"))
	if err == nil {
		t.Fatal("malformed row accepted")
	}
}

// Requirements: services-manager/FR-001.
func TestMergeServiceFilesRejectsMalformedRow(t *testing.T) {
	_, err := mergeServiceFiles(nil, []byte("bad-unit disabled\n"))
	if err == nil {
		t.Fatal("malformed unit file accepted")
	}
}
