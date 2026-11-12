package backuprestore

import (
	"errors"
	"reflect"
	"testing"
)

// mustMergeINI runs a merge expected to parse both inputs successfully.
func mustMergeINI(t *testing.T, snapshot string, current string) INIMergeResult {
	t.Helper()
	result, err := MergeINI(snapshot, current)
	if err != nil {
		t.Fatalf("MergeINI() error = %v", err)
	}
	return result
}

// Requirements: backup-restore/FR-011.
func TestMergeINIContent(t *testing.T) {
	cases := []struct {
		name     string
		snapshot string
		current  string
		want     string
	}{
		{
			name:     "preserves live-only keys and comments",
			snapshot: "[app]\nold=yes\n",
			current:  "# live comment\n[app]\nold=yes\nnew=keep\n",
			want:     "# live comment\n[app]\nold=yes\nnew=keep\n",
		},
		{
			name:     "adds snapshot-only key before next section",
			snapshot: "[app]\nold=yes\nadded=restore\n",
			current:  "[app]\nold=yes\n[next]\nkeep=1\n",
			want:     "[app]\nold=yes\nadded=restore\n[next]\nkeep=1\n",
		},
		{
			name:     "adds snapshot-only section",
			snapshot: "[added]\nkey=value\n",
			current:  "[current]\nkey=keep\n",
			want:     "[current]\nkey=keep\n[added]\nkey=value\n",
		},
		{
			name:     "uses current line ending",
			snapshot: "[app]\nnew=restore\n",
			current:  "[app]\r\nold=keep\r\n",
			want:     "[app]\r\nold=keep\r\nnew=restore\r\n",
		},
		{
			name:     "adds newline before appended section",
			snapshot: "[added]\nkey=value\n",
			current:  "[current]\nkey=keep",
			want:     "[current]\nkey=keep\n[added]\nkey=value\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := mustMergeINI(t, testCase.snapshot, testCase.current)
			if result.Content != testCase.want {
				t.Errorf("Content = %q, want %q", result.Content, testCase.want)
			}
		})
	}
}

// Requirements: backup-restore/FR-011.
func TestMergeINIReportsDifferentValues(t *testing.T) {
	result := mustMergeINI(t, "[app]\nkey=snapshot\n", "[app]\nkey=current\n")
	want := []INIConflict{{
		Section: "app", Key: "key", SnapshotValue: "snapshot", CurrentValue: "current",
	}}
	if !reflect.DeepEqual(result.Conflicts, want) {
		t.Errorf("Conflicts = %#v, want %#v", result.Conflicts, want)
	}
}

// Requirements: backup-restore/FR-011.
func TestMergeINILeavesConflictingCurrentValueUntouched(t *testing.T) {
	result := mustMergeINI(t, "[app]\nkey=snapshot\n", "[app]\nkey=current\n")
	if result.Content != "[app]\nkey=current\n" {
		t.Errorf("Content = %q, want current value", result.Content)
	}
}

// Requirements: backup-restore/FR-011.
func TestMergeINIRequiresManualResolutionForAmbiguousInput(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{name: "duplicate key", content: "[app]\nkey=one\nkey=two\n"},
		{name: "duplicate section", content: "[app]\nkey=one\n[app]\nother=two\n"},
		{name: "malformed header", content: "[app\nkey=one\n"},
		{name: "unsupported entry", content: "[app]\nkey: value\n"},
		{name: "invalid UTF-8", content: "[app]\nkey=\xff\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := MergeINI(testCase.content, "")
			if !errors.Is(err, ErrManualINIMerge) {
				t.Errorf("error = %v, want ErrManualINIMerge", err)
			}
		})
	}
}
