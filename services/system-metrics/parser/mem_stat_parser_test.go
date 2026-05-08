package parser

import (
	"testing"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/testcasetest"
)

func TestMemStatParserFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[metrics.MemSample]{
		{Name: "total bytes", Got: func(sample metrics.MemSample) any { return sample.TotalBytes }, Want: uint64(1024 * 1000)},
		{Name: "used bytes", Got: func(sample metrics.MemSample) any { return sample.UsedBytes }, Want: uint64(1024 * 750)},
		{Name: "used pct", Got: func(sample metrics.MemSample) any { return sample.UsedPct }, Want: 75.0},
	}

	testcasetest.RunFieldCases(t, loadMemSampleFixture, testCases)
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
