package status

import (
	"testing"
)

// TestParseZpoolStatusParsesHealthyPool verifies a standard healthy pool block.
func TestParseZpoolStatusParsesHealthyPool(t *testing.T) {
	input := `pool: LiteNAS
 state: ONLINE
  scan: scrub repaired 0B in 00:02:51 with 0 errors on Sun May 10 00:26:53 2026
config:

	NAME           STATE     READ WRITE CKSUM
	LiteNAS        ONLINE       0     0     0
	  mirror-0     ONLINE       0     0     0
	    /dev/sdb1  ONLINE       0     0     0
	    /dev/sda1  ONLINE       0     0     0

errors: No known data errors
`

	document, diagnostics, err := ParseZpoolStatus(input, ParseModeStrict)
	if err != nil {
		t.Fatalf("expected strict parse to succeed, got error: %v, diagnostics: %+v", err, diagnostics)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got: %+v", diagnostics)
	}
	if len(document.Pools) != 1 {
		t.Fatalf("expected one pool block, got %d", len(document.Pools))
	}

	pool := document.Pools[0]
	assertPoolSummary(t, pool)
	assertConfigHeader(t, pool.Config.Header)
	assertConfigTree(t, pool.Config.Roots)
}

// TestParseZpoolStatusStrictFailsOnInvalidInput verifies strict mode behavior.
func TestParseZpoolStatusStrictFailsOnInvalidInput(t *testing.T) {
	input := "pool: LiteNAS\nstate: ONLINE\n"

	_, diagnostics, err := ParseZpoolStatus(input, ParseModeStrict)
	if err == nil {
		t.Fatal("expected strict parse error for invalid payload")
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for invalid payload")
	}
}

// TestParseZpoolStatusTolerantReturnsDiagnostics verifies tolerant mode behavior.
func TestParseZpoolStatusTolerantReturnsDiagnostics(t *testing.T) {
	input := "pool: LiteNAS\nstate: ONLINE\n"

	document, diagnostics, err := ParseZpoolStatus(input, ParseModeTolerant)
	if err != nil {
		t.Fatalf("expected tolerant parse not to return error, got: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for invalid payload")
	}
	if len(document.Pools) != 1 {
		t.Fatalf("expected one best-effort pool from malformed input, got %d", len(document.Pools))
	}
}

func assertPoolSummary(t *testing.T, pool PoolBlock) {
	t.Helper()

	if pool.PoolName != "LiteNAS" {
		t.Fatalf("expected pool name LiteNAS, got %q", pool.PoolName)
	}
	if pool.Metadata.State != "ONLINE" {
		t.Fatalf("expected state ONLINE, got %q", pool.Metadata.State)
	}
	if pool.ErrorsSummary != "No known data errors" {
		t.Fatalf("expected no known data errors summary, got %q", pool.ErrorsSummary)
	}
}

func assertConfigHeader(t *testing.T, header []string) {
	t.Helper()

	if len(header) != 5 {
		t.Fatalf("expected 5 config header columns, got %d", len(header))
	}
	if header[0] != "NAME" || header[1] != "STATE" {
		t.Fatalf("unexpected header prefix: %+v", header)
	}
}

func assertConfigTree(t *testing.T, roots []ConfigNode) {
	t.Helper()

	if len(roots) != 1 {
		t.Fatalf("expected one root config node, got %d", len(roots))
	}
	root := roots[0]
	if root.Name != "LiteNAS" {
		t.Fatalf("expected root node LiteNAS, got %q", root.Name)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected one mirror child, got %d", len(root.Children))
	}
	mirror := root.Children[0]
	if mirror.Name != "mirror-0" {
		t.Fatalf("expected mirror-0 child, got %q", mirror.Name)
	}
	if len(mirror.Children) != 2 {
		t.Fatalf("expected two leaf disks, got %d", len(mirror.Children))
	}
}
