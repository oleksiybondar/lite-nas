package parser

import (
	"testing"

	"lite-nas/shared/metrics"
)

func TestMemStatParserFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func() any
		want any
	}{
		{name: "total bytes", got: func() any { return loadMemSampleFixture(t).TotalBytes }, want: uint64(1024 * 1000)},
		{name: "used bytes", got: func() any { return loadMemSampleFixture(t).UsedBytes }, want: uint64(1024 * 750)},
		{name: "used pct", got: func() any { return loadMemSampleFixture(t).UsedPct }, want: 75.0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.got(); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestMemStatParserParseRejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	if _, err := (MemStatParser{}).Parse("MemTotal: 1000 kB\n"); err == nil {
		t.Fatal("expected missing meminfo lines error")
	}
}

func loadMemSampleFixture(t *testing.T) metrics.MemSample {
	t.Helper()

	sample, err := (MemStatParser{}).Parse(
		"MemTotal:      1000 kB\n" +
			"MemFree:        100 kB\n" +
			"MemAvailable:   250 kB\n",
	)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	return sample
}
