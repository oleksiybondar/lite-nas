package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/testutil/messagingtest"
)

// downloadClient serves checksummed chunks from a controlled byte slice.
type downloadClient struct {
	messagingtest.RecordingClient
	content []byte
	corrupt bool
}

// sessionStartClient serves the start request and its input transfers.
type sessionStartClient struct {
	downloadClient
	started bool
}

// Request returns one selected file for an interactive merge start.
func (client *sessionStartClient) Request(ctx context.Context, subject string, request any, response any) error {
	if subject == contract.MergeStartSubject {
		client.started = true
		*response.(*contract.MergeStartResponse) = contract.MergeStartResponse{SessionID: "0123456789abcdef0123456789abcdef", Files: []contract.MergeFile{{Path: "/etc/example.conf"}}, ApplyAvailable: true}
		return nil
	}
	return client.downloadClient.Request(ctx, subject, request, response)
}

// Request returns a bounded chunk at the requested offset.
func (client *downloadClient) Request(_ context.Context, subject string, request any, response any) error {
	if subject != contract.MergeReadSubject {
		return fmt.Errorf("unexpected subject")
	}
	offset := request.(contract.MergeReadRequest).Offset
	end := min(offset+mergeChunkSize, len(client.content))
	checksum := sha256.Sum256(client.content)
	result := contract.MergeReadResponse{Data: base64.StdEncoding.EncodeToString(client.content[offset:end]), NextOffset: end, Done: end == len(client.content), SHA256: hex.EncodeToString(checksum[:])}
	if client.corrupt {
		result.SHA256 = "bad"
	}
	*response.(*contract.MergeReadResponse) = result
	return nil
}

// recordingMergeClient collects submitted chunks without a NATS server.
type recordingMergeClient struct {
	messagingtest.RecordingClient
	content []byte
}

// Request records sequential upload bytes and confirms the next offset.
func (client *recordingMergeClient) Request(_ context.Context, subject string, request any, response any) error {
	if subject != contract.MergeWriteSubject {
		return nil
	}
	chunk := request.(contract.MergeWriteRequest)
	data, err := base64.StdEncoding.DecodeString(chunk.Data)
	if err != nil {
		return err
	}
	client.content = append(client.content, data...)
	response.(*contract.MergeWriteResponse).NextOffset = len(client.content)
	return nil
}

// Requirements: backup-restore/FR-005.
func TestUploadMergeChunksLargeResult(t *testing.T) {
	client := &recordingMergeClient{}
	content := make([]byte, mergeChunkSize+7)
	for index := range content {
		content[index] = 'x'
	}
	if err := uploadMerge(context.Background(), client, "token", "session", "/etc/example", content); err != nil {
		t.Fatal(err)
	}
	if string(client.content) != string(content) {
		t.Fatal("uploaded content differs")
	}
}

// Requirements: backup-restore/FR-005, backup-restore/IR-002.
func TestDownloadMergeAssemblesMultipleChunks(t *testing.T) {
	client := &downloadClient{content: make([]byte, mergeChunkSize+7)}
	for index := range client.content {
		client.content[index] = 'x'
	}
	got, err := downloadMerge(context.Background(), client, "token", "session", "/etc/example", "snapshot")
	if err != nil || string(got) != string(client.content) {
		t.Fatalf("download = %d bytes, %v", len(got), err)
	}
}

// Requirements: backup-restore/SR-003.
func TestDownloadMergeRejectsChecksumMismatch(t *testing.T) {
	client := &downloadClient{content: []byte("snapshot"), corrupt: true}
	_, err := downloadMerge(context.Background(), client, "token", "session", "/etc/example", "snapshot")
	if err == nil {
		t.Fatal("checksum mismatch accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestReadEditedResultRejectsOversizedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "result")
	if err := os.WriteFile(path, make([]byte, mergeFileLimit+1), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := readEditedResult(path)
	if err == nil {
		t.Fatal("oversized result accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestLaunchEditorRejectsCommandString(t *testing.T) {
	t.Setenv("EDITOR", "sh -c echo")
	err := launchEditor(context.Background(), "/tmp/snapshot", "/tmp/result", "/tmp/current")
	if err == nil {
		t.Fatal("command-shaped EDITOR accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestMergeRejectsMalformedSessionIDBeforeRPC(t *testing.T) {
	client := &recordingMergeClient{}
	err := runMerge(context.Background(), client, "token", []string{"discard", "--host=evil"})
	if err == nil {
		t.Fatal("malformed session ID accepted")
	}
}

// Requirements: backup-restore/FR-005, backup-restore/FR-006.
func TestMergeTerminalChoiceSelectsSubject(t *testing.T) {
	for _, testCase := range []struct {
		choice, subject string
		reply           any
	}{
		{"apply", contract.MergeApplySubject, contract.MergeActionResponse{Completed: true}},
		{"save", contract.MergeSaveSubject, contract.BackupResponse{Version: "version"}},
		{"discard", contract.MergeDiscardSubject, contract.MergeActionResponse{Completed: true}},
	} {
		t.Run(testCase.choice, func(t *testing.T) {
			client := &commandClient{reply: testCase.reply}
			if err := submitMergeChoice(context.Background(), client, "token", "session", testCase.choice, true); err != nil {
				t.Fatal(err)
			}
			if client.subject != testCase.subject {
				t.Fatalf("subject = %s", client.subject)
			}
		})
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeTerminalChoiceRejectsApplyAfterConflict(t *testing.T) {
	client := &commandClient{}
	err := submitMergeChoice(context.Background(), client, "token", "session", "apply", false)
	if err == nil || client.calls != 0 {
		t.Fatalf("unavailable apply = %v, calls = %d", err, client.calls)
	}
}

// Requirements: backup-restore/IR-003.
func TestEditorFailurePreservesMergeSession(t *testing.T) {
	t.Setenv("EDITOR", "sh -c echo")
	client := &downloadClient{content: []byte("text\n")}
	session := contract.MergeStartResponse{SessionID: "0123456789abcdef0123456789abcdef", Files: []contract.MergeFile{{Path: "/etc/example.conf"}}, ApplyAvailable: true}
	err := editMerge(context.Background(), client, "token", session)
	if err == nil || !strings.Contains(err.Error(), "remains active") {
		t.Fatalf("editor failure = %v", err)
	}
}

// Requirements: backup-restore/IR-003.
func TestMergeCommandStartsAndDownloadsSession(t *testing.T) {
	t.Setenv("EDITOR", "sh -c echo")
	client := &sessionStartClient{downloadClient: downloadClient{content: []byte("text\n")}}
	err := runMerge(context.Background(), client, "token", []string{"example", "20261031T120000.000000000Z"})
	if err == nil || !client.started {
		t.Fatalf("merge start = %v, started = %v", err, client.started)
	}
}
