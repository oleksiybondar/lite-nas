package main

import (
	"bytes"
	"context"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	contract "lite-nas/shared/contracts/journalmanager"
	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/hostcommandtest"
	"lite-nas/shared/testutil/messagingtest"
)

// Requirements: journal-manager/SR-002.
func TestLimitedWriterRejectsExcessCommandOutput(t *testing.T) {
	var buffer bytes.Buffer
	writer := &limitedWriter{writer: &buffer, remaining: 2}
	_, err := writer.Write([]byte("large"))
	if err == nil {
		t.Fatal("oversized command output accepted")
	}
}

// Requirements: journal-manager/IR-001.
func TestExecRunnerCapturesFixedCommandOutput(t *testing.T) {
	output, err := (execRunner{}).Output(context.Background(), "/usr/bin/printf", "ready")
	if err != nil || string(output) != "ready" {
		t.Fatalf("output = %q, %v", output, err)
	}
}

// testManager supplies a real verifier and recording command runner.
func testManager(t *testing.T, roles []string) (*manager, string) {
	t.Helper()
	verifier, token := authtokentest.MustIssueAccessToken(t, roles)
	return &manager{verifier: verifier, validator: validator.New(validator.WithRequiredStructEnabled()), runner: &hostcommandtest.RecordingRunner{}, jobs: make(chan commandJob, 1), ctx: context.Background()}, token
}

// Requirements: journal-manager/SR-001.
func TestQueryRejectsOperator(t *testing.T) {
	m, token := testManager(t, []string{"lite-nas-operator"})
	_, err := m.query(context.Background(), messagingtest.MustEnvelope(t, contract.QueryRequest{AccessToken: token, Limit: 50}))
	if err == nil {
		t.Fatal("operator queried journal")
	}
}

// Requirements: journal-manager/SR-001, journal-manager/FR-003.
func TestUnitsRejectsOperator(t *testing.T) {
	m, token := testManager(t, []string{"lite-nas-operator"})
	_, err := m.units(context.Background(), messagingtest.MustEnvelope(t, contract.UnitsRequest{AccessToken: token}))
	if err == nil {
		t.Fatal("operator listed journal units")
	}
}

// Requirements: journal-manager/SR-002.
func TestQueryRejectsInvalidFilters(t *testing.T) {
	for name, change := range map[string]func(*contract.QueryRequest){
		"option-shaped-unit":   func(q *contract.QueryRequest) { q.Unit = "--help" },
		"newline-search":       func(q *contract.QueryRequest) { q.Message = "x\ny" },
		"invalid-cursor":       func(q *contract.QueryRequest) { q.Cursor = "--bad cursor" },
		"option-shaped-cursor": func(q *contract.QueryRequest) { q.Cursor = "--help" },
		"reversed-time":        func(q *contract.QueryRequest) { q.Since = "2026-11-01T12:00:00Z"; q.Until = "2026-11-01T11:00:00Z" },
	} {
		t.Run(name, func(t *testing.T) {
			q := contract.QueryRequest{Limit: 50}
			change(&q)
			if err := validateQuery(q); err == nil {
				t.Fatal("invalid filter accepted")
			}
		})
	}
}

// Requirements: journal-manager/FR-001, journal-manager/IR-001.
func TestExecuteBuildsBoundedJournalCommand(t *testing.T) {
	m, _ := testManager(t, []string{"lite-nas-security"})
	q := contract.QueryRequest{Limit: 2, Message: "disk full", Cursor: "s=abc;i=42"}
	m.execute(commandJob{ctx: context.Background(), query: &q})
	runner := m.runner.(*hostcommandtest.RecordingRunner)
	want := []string{"--no-pager", "--quiet", "--utc", "--output=json", "--reverse", "--lines=3", "--grep=disk full", "--case-sensitive=no", "--after-cursor=s=abc;i=42"}
	if runner.Name != journalctlBinary || !reflect.DeepEqual(runner.Args, want) {
		t.Fatalf("command = %s %v", runner.Name, runner.Args)
	}
}

// Requirements: journal-manager/FR-003, journal-manager/SR-002.
func TestExecuteBuildsFilteredUnitDiscoveryCommand(t *testing.T) {
	m, _ := testManager(t, []string{"lite-nas-security"})
	q := contract.UnitsRequest{Message: "disk.full"}
	m.execute(commandJob{ctx: context.Background(), units: &q})
	runner := m.runner.(*hostcommandtest.RecordingRunner)
	want := []string{"--no-pager", "--quiet", "--output=json", "--output-fields=_SYSTEMD_UNIT", "--grep=disk\\.full", "--case-sensitive=no"}
	if runner.Name != journalctlBinary || !reflect.DeepEqual(runner.Args, want) {
		t.Fatalf("command = %s %v", runner.Name, runner.Args)
	}
}

// Requirements: journal-manager/FR-001, journal-manager/FR-003.
func TestJournalOutputTreatsVerifiedGrepMissAsEmpty(t *testing.T) {
	exitError := exec.Command("/usr/bin/false").Run()
	m, _ := testManager(t, []string{"lite-nas-security"})
	m.runner.(*hostcommandtest.RecordingRunner).Results = []hostcommandtest.Result{{Err: exitError}, {Output: []byte(`{"MESSAGE":"available"}`)}}
	output, err := m.journalOutput(context.Background(), []string{"--no-pager", "--grep=missing"}, true)
	if err != nil || len(output) != 0 {
		t.Fatalf("no-match output = %q, %v", output, err)
	}
}

// Requirements: journal-manager/FR-001, journal-manager/SR-002.
func TestJournalOutputPreservesAccessFailure(t *testing.T) {
	exitError := exec.Command("/usr/bin/false").Run()
	m, _ := testManager(t, []string{"lite-nas-security"})
	m.runner.(*hostcommandtest.RecordingRunner).Results = []hostcommandtest.Result{{Err: exitError}, {Err: exitError}}
	_, err := m.journalOutput(context.Background(), []string{"--no-pager", "--grep=missing"}, true)
	if err == nil {
		t.Fatal("journal access failure was treated as an empty search")
	}
}

// Requirements: journal-manager/FR-001.
func TestParseEntriesReturnsCursorAndStructuredFields(t *testing.T) {
	output := []byte(`{"__CURSOR":"s=one","__REALTIME_TIMESTAMP":"123","MESSAGE":"first","_SYSTEMD_UNIT":"ssh.service"}` + "\n" + `{"__CURSOR":"s=two","MESSAGE":"second"}` + "\n")
	response, err := parseEntries(output, 1)
	if err != nil || len(response.Entries) != 1 || response.Entries[0].Unit != "ssh.service" || response.NextCursor != "s=one" || !response.HasMore {
		t.Fatalf("response = %#v, %v", response, err)
	}
}

// Requirements: journal-manager/FR-001.
func TestParseEntriesRejectsMalformedJSON(t *testing.T) {
	_, err := parseEntries([]byte("not json"), 50)
	if err == nil {
		t.Fatal("invalid journal JSON accepted")
	}
}

// Requirements: journal-manager/FR-001.
func TestParseEntriesRejectsMissingCursor(t *testing.T) {
	_, err := parseEntries([]byte(`{"MESSAGE":"missing cursor"}`), 50)
	if err == nil {
		t.Fatal("cursorless journal entry accepted")
	}
}

// Requirements: journal-manager/FR-003.
func TestParseUnitsReturnsSortedDistinctServiceUnits(t *testing.T) {
	output := []byte("ssh.service\ncron.service\nssh.service\nsession-1.scope\n")
	response, err := parseUnits(output, false)
	if err != nil || !reflect.DeepEqual(response.Units, []string{"cron.service", "ssh.service"}) {
		t.Fatalf("units = %v, %v", response.Units, err)
	}
}

// Requirements: journal-manager/FR-003.
func TestParseFilteredUnitsUsesMatchingEntries(t *testing.T) {
	output := []byte(`{"_SYSTEMD_UNIT":"ssh.service"}` + "\n" + `{"_SYSTEMD_UNIT":"cron.service"}` + "\n")
	response, err := parseUnits(output, true)
	if err != nil || !reflect.DeepEqual(response.Units, []string{"cron.service", "ssh.service"}) {
		t.Fatalf("units = %v, %v", response.Units, err)
	}
}

// Requirements: journal-manager/FR-002, journal-manager/SR-002.
func TestVacuumHelperRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{{"vacuum", "0"}, {"vacuum", "366"}, {"vacuum", "--help"}, {"vacuum", "30", "--rotate"}, {"query", "30"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			if err := runVacuum(args); err == nil {
				t.Fatal("invalid vacuum arguments accepted")
			}
		})
	}
}
