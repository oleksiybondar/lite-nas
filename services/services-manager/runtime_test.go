package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"

	contract "lite-nas/shared/contracts/servicesmanager"
	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/hostcommandtest"
	"lite-nas/shared/testutil/messagingtest"
)

// newTestManager injects a real token verifier and recording command runner.
func newTestManager(t *testing.T, roles []string) (*manager, string) {
	t.Helper()
	verifier, token := authtokentest.MustIssueAccessToken(t, roles)
	return &manager{verifier: verifier, validator: validator.New(validator.WithRequiredStructEnabled()), runner: &hostcommandtest.RecordingRunner{}, jobs: make(chan commandJob, 1), ctx: context.Background()}, token
}

// Requirements: services-manager/SR-001.
func TestListRequiresAdministratorRole(t *testing.T) {
	m, token := newTestManager(t, []string{"lite-nas-security"})
	_, err := m.list(context.Background(), messagingtest.MustEnvelope(t, contract.ListRequest{AccessToken: token}))
	if err == nil {
		t.Fatal("security officer listed services")
	}
}

// Requirements: services-manager/SR-002.
func TestActionRejectsOptionShapedUnit(t *testing.T) {
	m, token := newTestManager(t, []string{"admin"})
	_, err := m.action(context.Background(), messagingtest.MustEnvelope(t, contract.ActionRequest{AccessToken: token, Unit: "--host=evil.service"}), "start")
	if err == nil {
		t.Fatal("option-shaped unit accepted")
	}
}

// Requirements: services-manager/FR-002, services-manager/SR-001.
func TestActionRunsForAdministrator(t *testing.T) {
	m, token := newTestManager(t, []string{"admin"})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.work(ctx)
	response, err := m.action(context.Background(), messagingtest.MustEnvelope(t, contract.ActionRequest{AccessToken: token, Unit: "ssh.service"}), "start")
	if err != nil || !response.(contract.ActionResponse).Completed {
		t.Fatalf("action = %#v, %v", response, err)
	}
}

// Requirements: services-manager/FR-002, services-manager/IR-001.
func TestExecuteUsesFixedSudoCommand(t *testing.T) {
	m, _ := newTestManager(t, []string{"admin"})
	runner := m.runner.(*hostcommandtest.RecordingRunner)
	m.execute(commandJob{ctx: context.Background(), action: "restart", unit: "ssh.service"})
	want := []string{"-n", "/usr/bin/systemctl", "--system", "restart", "--", "ssh.service"}
	if runner.Name != "/usr/bin/sudo" || !reflect.DeepEqual(runner.Args, want) {
		t.Fatalf("command = %q %v", runner.Name, runner.Args)
	}
}

// Requirements: services-manager/FR-001.
func TestParseServicesReturnsServiceState(t *testing.T) {
	services, err := parseServices([]byte("ssh.service loaded active running OpenSSH server\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Name != "ssh.service" || services[0].Description != "OpenSSH server" {
		t.Fatalf("services = %#v", services)
	}
}

// Requirements: services-manager/FR-001.
func TestMergeServiceFilesIncludesDisabledService(t *testing.T) {
	services, err := mergeServiceFiles(nil, []byte("cups.service disabled enabled\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 || services[0].Name != "cups.service" || services[0].EnabledState != "disabled" {
		t.Fatalf("services = %#v", services)
	}
}
