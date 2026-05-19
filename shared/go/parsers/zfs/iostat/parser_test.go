package iostat

import "testing"

func TestParse(t *testing.T) {
	input := "LiteNAS\t30923764531\t1957130697441\t0\t0\t44544\t384\n"
	statsByPool, err := Parse(input)
	if err != nil {
		t.Fatalf("expected parse to succeed, got: %v", err)
	}
	stats, ok := statsByPool["LiteNAS"]
	if !ok {
		t.Fatal("expected LiteNAS entry")
	}
	if stats.Bandwidth.Read != 44544 {
		t.Fatalf("unexpected read bandwidth: %d", stats.Bandwidth.Read)
	}
}
