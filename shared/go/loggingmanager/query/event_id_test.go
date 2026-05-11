package query

import "testing"

func TestBuildEventID(t *testing.T) {
	t.Parallel()

	eventID, err := BuildEventID("perf", 1)
	if err != nil {
		t.Fatalf("BuildEventID() error = %v", err)
	}
	if eventID != "perf_1" {
		t.Fatalf("eventID = %q, want %q", eventID, "perf_1")
	}
}

func TestBuildEventIDRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		prefix string
		seq    uint32
	}{
		{name: "empty prefix", prefix: "", seq: 1},
		{name: "prefix too long", prefix: "abcdefghijk", seq: 1},
		{name: "sequence overflow", prefix: "event", seq: EventIDMaxSequence + 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if _, err := BuildEventID(testCase.prefix, testCase.seq); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
