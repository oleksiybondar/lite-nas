package list

import "testing"

func TestParse(t *testing.T) {
	input := "LiteNAS\t1990111166792\t30923764531\t1957130697441\t1\tONLINE\n"
	usageByPool, err := Parse(input)
	if err != nil {
		t.Fatalf("expected parse to succeed, got: %v", err)
	}
	usage, ok := usageByPool["LiteNAS"]
	if !ok {
		t.Fatal("expected LiteNAS entry")
	}
	if usage.CapacityPct != 1 {
		t.Fatalf("unexpected capacity pct: %v", usage.CapacityPct)
	}
}
