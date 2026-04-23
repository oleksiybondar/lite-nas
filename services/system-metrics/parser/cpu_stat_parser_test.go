package parser

import (
	"testing"

	"lite-nas/shared/metrics"
)

func TestCPUStatParserTotalFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func() uint64
		want uint64
	}{
		{name: "total", got: func() uint64 { return loadCPUSampleFixture(t).Total.Total }, want: 126},
		{name: "idle", got: func() uint64 { return loadCPUSampleFixture(t).Total.Idle }, want: 45},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.got(); got != testCase.want {
				t.Fatalf("%s = %d, want %d", testCase.name, got, testCase.want)
			}
		})
	}
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
