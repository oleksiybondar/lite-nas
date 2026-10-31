package main

import (
	"context"
	"errors"
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
	calls   int
}

// Request supplies the expected RPC response to an isolated command test.
func (client *commandClient) Request(_ context.Context, subject string, request any, response any) error {
	client.subject, client.request = subject, request
	client.calls++
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
