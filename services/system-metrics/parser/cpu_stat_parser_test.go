package parser

import (
	"testing"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/testcasetest"
)

func TestCPUStatParserTotalFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[metrics.CPURawSample]{
		{Name: "total", Got: func(sample metrics.CPURawSample) any { return sample.Total.Total }, Want: uint64(126)},
		{Name: "idle", Got: func(sample metrics.CPURawSample) any { return sample.Total.Idle }, Want: uint64(45)},
	}

	testcasetest.RunFieldCases(t, loadCPUSampleFixture, testCases)
}

func TestCPUStatParserCoreCount(t *testing.T) {
	t.Parallel()

	if got := len(loadCPUSampleFixture(t).Cores); got != 2 {
		t.Fatalf("len(Cores) = %d, want 2", got)
	}
}

func TestCPUStatParserParseRejectsMissingAggregateLine(t *testing.T) {
	t.Parallel()

	if _, err := (CPUStatParser{}).Parse("cpu0 1 2 3 4 5\n"); err == nil {
		t.Fatal("expected missing aggregate cpu line error")
	}
}

func loadCPUSampleFixture(t *testing.T) metrics.CPURawSample {
	t.Helper()

	sample, err := (CPUStatParser{}).Parse(
		"cpu  10 20 30 40 5 6 7 8 0 0\n" +
			"cpu0 1 2 3 4 1 0 0 0 0 0\n" +
			"cpu1 2 3 4 5 2 0 0 0 0 0\n" +
			"intr 1\n",
	)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	return sample
}
