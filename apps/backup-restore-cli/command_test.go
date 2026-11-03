package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/testutil/messagingtest"
)

// commandClient records one RPC and supplies a typed response.
type commandClient struct {
	messagingtest.RecordingClient
	subject string
	request any
	reply   any
	err     error
	calls   int
}

// Request supplies the expected RPC response to an isolated command test.
func (client *commandClient) Request(_ context.Context, subject string, request any, response any) error {
	client.subject, client.request = subject, request
	client.calls++
	if client.err != nil {
		return client.err
	}
	switch value := client.reply.(type) {
	case contract.ApplicationsResponse:
		*response.(*contract.ApplicationsResponse) = value
	case contract.VersionsResponse:
		*response.(*contract.VersionsResponse) = value
	case contract.BackupResponse:
		*response.(*contract.BackupResponse) = value
	case contract.BackupAllResponse:
		*response.(*contract.BackupAllResponse) = value
	case contract.DiffResponse:
		*response.(*contract.DiffResponse) = value
	case contract.RestoreResponse:
		*response.(*contract.RestoreResponse) = value
	case contract.RestoreByNameResponse:
		*response.(*contract.RestoreByNameResponse) = value
	case contract.RestoreAllByNameResponse:
		*response.(*contract.RestoreAllByNameResponse) = value
	case contract.MergeStartResponse:
		*response.(*contract.MergeStartResponse) = value
	case contract.MergeReadResponse:
		*response.(*contract.MergeReadResponse) = value
	case contract.MergeWriteResponse:
		*response.(*contract.MergeWriteResponse) = value
	case contract.MergeActionResponse:
		*response.(*contract.MergeActionResponse) = value
	default:
		return errors.New("unexpected response type")
	}
	return nil
}

// Requirements: backup-restore/IR-003.
func TestBackupCommandSendsLabels(t *testing.T) {
	client := &commandClient{reply: contract.BackupResponse{Version: "20261031T120000.000000000Z"}}
	err := executeCommand(context.Background(), client, "token", []string{"backup", "example", "--name", "named", "--description", "note"})
	if err != nil {
		t.Fatal(err)
	}
	request := client.request.(contract.BackupRequest)
	if client.subject != contract.BackupSubject || request.Name != "named" || request.Description != "note" {
		t.Fatalf("backup request = %#v on %s", request, client.subject)
	}
}

// Requirements: backup-restore/SR-003.
func TestCommandRejectsMalformedApplicationBeforeRPC(t *testing.T) {
	client := &commandClient{}
	_ = executeCommand(context.Background(), client, "token", []string{"backup", "--host=evil"})
	if client.calls != 0 {
		t.Fatal("malformed application reached NATS")
	}
}

// Requirements: backup-restore/IR-003.
func TestRestoreCommandReportsRestartWarningAsFailure(t *testing.T) {
	client := &commandClient{reply: contract.RestoreResponse{Restored: true, Warnings: []string{"example.service: denied"}}}
	err := executeCommand(context.Background(), client, "token", []string{"restore", "example", "20261031T120000.000000000Z"})
	if err == nil {
		t.Fatal("restart warning did not fail command")
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeDiscardSendsOwnedSessionAction(t *testing.T) {
	client := &commandClient{reply: contract.MergeActionResponse{Completed: true}}
	err := runMerge(context.Background(), client, "token", []string{"discard", "0123456789abcdef0123456789abcdef"})
	if err != nil {
		t.Fatal(err)
	}
	if client.subject != contract.MergeDiscardSubject {
		t.Fatalf("subject = %s", client.subject)
	}
}

// Requirements: backup-restore/IR-003.
func TestReadCommandsSelectExpectedNATSSubject(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		reply   any
		subject string
	}{
		{"applications", []string{"applications"}, contract.ApplicationsResponse{}, contract.ApplicationsSubject},
		{"backed-up", []string{"backed-up"}, contract.ApplicationsResponse{}, contract.BackedUpSubject},
		{"versions", []string{"versions", "example"}, contract.VersionsResponse{}, contract.VersionsSubject},
		{"diff", []string{"diff", "example", "20261031T120000.000000000Z"}, contract.DiffResponse{}, contract.DiffSubject},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			client := &commandClient{reply: testCase.reply}
			if err := executeCommand(context.Background(), client, "token", testCase.args); err != nil {
				t.Fatal(err)
			}
			if client.subject != testCase.subject {
				t.Fatalf("subject = %s", client.subject)
			}
		})
	}
}

// Requirements: backup-restore/IR-003.
func TestBackupAllReportsFailedApplication(t *testing.T) {
	client := &commandClient{reply: contract.BackupAllResponse{Results: []contract.BackupAllResult{{Application: "example", Status: "failed", Error: "source missing"}}}}
	err := executeCommand(context.Background(), client, "token", []string{"backup", "all"})
	if err == nil {
		t.Fatal("failed application backup was reported as success")
	}
}

// Requirements: backup-restore/IR-003.
func TestRestoreCommandSendsMergeMode(t *testing.T) {
	client := &commandClient{reply: contract.RestoreResponse{Restored: true}}
	if err := executeCommand(context.Background(), client, "token", []string{"restore", "example", "20261031T120000.000000000Z", "merge"}); err != nil {
		t.Fatal(err)
	}
	if client.request.(contract.RestoreRequest).Mode != "merge" {
		t.Fatalf("mode = %s", client.request.(contract.RestoreRequest).Mode)
	}
}

// Requirements: backup-restore/FR-012, backup-restore/IR-003.
// TestRestoreByNameCommandSendsName verifies the exact name reaches the new RPC.
func TestRestoreByNameCommandSendsName(t *testing.T) {
	client := &commandClient{reply: contract.RestoreByNameResponse{Version: "20261031T120000.000000000Z", Restored: true}}
	if err := executeCommand(context.Background(), client, "token", []string{"restore-by-name", "example", "release one", "merge"}); err != nil {
		t.Fatal(err)
	}
	request := client.request.(contract.RestoreByNameRequest)
	if client.subject != contract.RestoreByNameSubject || request.Application != "example" || request.Name != "release one" || request.Mode != "merge" {
		t.Fatalf("restore-by-name request = %#v on %s", request, client.subject)
	}
}

// Requirements: backup-restore/FR-012, backup-restore/IR-003.
// TestRestoreAllByNameCommandSendsName verifies all-target dispatch.
func TestRestoreAllByNameCommandSendsName(t *testing.T) {
	client := &commandClient{reply: contract.RestoreAllByNameResponse{Results: []contract.RestoreAllByNameResult{{Application: "example", Status: "completed"}}}}
	if err := executeCommand(context.Background(), client, "token", []string{"restore-by-name", "all", "release one"}); err != nil {
		t.Fatal(err)
	}
	request := client.request.(contract.RestoreAllByNameRequest)
	if client.subject != contract.RestoreAllByNameSubject || request.Name != "release one" || request.Mode != "replace" {
		t.Fatalf("restore-all-by-name request = %#v on %s", request, client.subject)
	}
}

// Requirements: backup-restore/FR-012, backup-restore/SR-003.
// TestRestoreByNameRejectsEmptyName checks the CLI boundary before NATS.
func TestRestoreByNameRejectsEmptyName(t *testing.T) {
	client := &commandClient{}
	_ = executeCommand(context.Background(), client, "token", []string{"restore-by-name", "example", ""})
	if client.calls != 0 {
		t.Fatal("empty backup name reached NATS")
	}
}

// Requirements: backup-restore/IR-003.
// TestCommandPropagatesRPCFailure checks failed requests cannot be reported as success.
func TestCommandPropagatesRPCFailure(t *testing.T) {
	for _, arguments := range [][]string{
		{"applications"},
		{"versions", "example"},
		{"backup", "example"},
		{"backup", "all"},
		{"diff", "example", "v1"},
		{"restore", "example", "v1"},
		{"restore-by-name", "example", "release"},
		{"restore-by-name", "all", "release"},
	} {
		t.Run(strings.Join(arguments, "/"), func(t *testing.T) {
			client := &commandClient{err: errors.New("rpc unavailable")}
			if err := executeCommand(context.Background(), client, "token", arguments); err == nil {
				t.Fatal("RPC failure was ignored")
			}
		})
	}
}

// Requirements: backup-restore/FR-012, backup-restore/IR-003.
// TestNamedRestoreReportsUnconfirmedAndWarningOutcomes checks operator-visible failures.
func TestNamedRestoreReportsUnconfirmedAndWarningOutcomes(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		args  []string
		reply any
	}{
		{"single unconfirmed", []string{"restore-by-name", "example", "release"}, contract.RestoreByNameResponse{}},
		{"single warning", []string{"restore-by-name", "example", "release"}, contract.RestoreByNameResponse{Restored: true, Warnings: []string{"restart failed"}}},
		{"all failed", []string{"restore-by-name", "all", "release"}, contract.RestoreAllByNameResponse{Results: []contract.RestoreAllByNameResult{{Status: "failed"}}}},
		{"all warning", []string{"restore-by-name", "all", "release"}, contract.RestoreAllByNameResponse{Results: []contract.RestoreAllByNameResult{{Status: "completed", Warnings: []string{"restart failed"}}}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client := &commandClient{reply: testCase.reply}
			if err := executeCommand(context.Background(), client, "token", testCase.args); err == nil {
				t.Fatal("restore outcome was reported as success")
			}
		})
	}
}

// Requirements: backup-restore/IR-003.
// TestBackupLabelsRejectMalformedArguments checks CLI labels before an RPC.
func TestBackupLabelsRejectMalformedArguments(t *testing.T) {
	for _, testCase := range []struct {
		name string
		args []string
	}{
		{"missing value", []string{"--name"}},
		{"duplicate", []string{"--name", "first", "--name", "second"}},
		{"option value", []string{"--name", "--help"}},
		{"long name", []string{"--name", strings.Repeat("n", 129)}},
		{"long description", []string{"--description", strings.Repeat("d", 1025)}},
		{"unknown flag", []string{"--unknown", "value"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client := &commandClient{}
			_ = executeCommand(context.Background(), client, "token", append([]string{"backup", "example"}, testCase.args...))
			if client.calls != 0 {
				t.Fatal("malformed label reached RPC")
			}
		})
	}
}

// Requirements: backup-restore/SR-003, backup-restore/IR-003.
// TestCommandRejectsInvalidOperands checks each command family before an RPC.
func TestCommandRejectsInvalidOperands(t *testing.T) {
	for _, testCase := range []struct {
		name string
		args []string
	}{
		{"no command", nil},
		{"inventory extra", []string{"applications", "example"}},
		{"versions missing", []string{"versions"}},
		{"versions extra", []string{"versions", "example", "extra"}},
		{"diff invalid version", []string{"diff", "example", "--help"}},
		{"restore invalid mode", []string{"restore", "example", "v1", "overwrite"}},
		{"named restore invalid mode", []string{"restore-by-name", "example", "release", "overwrite"}},
		{"named restore long name", []string{"restore-by-name", "example", strings.Repeat("n", 129)}},
		{"unknown command", []string{"unknown"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			client := &commandClient{}
			_ = executeCommand(context.Background(), client, "token", testCase.args)
			if client.calls != 0 {
				t.Fatal("invalid command reached RPC")
			}
		})
	}
}

// Requirements: backup-restore/IR-003.
// TestBackupAllAcceptsCompletedApplications verifies a successful all-target backup.
func TestBackupAllAcceptsCompletedApplications(t *testing.T) {
	client := &commandClient{reply: contract.BackupAllResponse{Results: []contract.BackupAllResult{{Application: "example", Status: "completed", Version: "v1"}}}}
	if err := executeCommand(context.Background(), client, "token", []string{"backup", "all"}); err != nil {
		t.Fatal(err)
	}
}

// Requirements: backup-restore/IR-003.
// TestRestoreCommandRequiresConfirmation prevents a false success message.
func TestRestoreCommandRequiresConfirmation(t *testing.T) {
	client := &commandClient{reply: contract.RestoreResponse{Restored: false}}
	if err := executeCommand(context.Background(), client, "token", []string{"restore", "example", "v1"}); err == nil {
		t.Fatal("unconfirmed restore accepted")
	}
}
