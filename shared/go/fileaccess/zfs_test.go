package fileaccess

import (
	"strings"
	"testing"
)

// TestParseZFSRootsKeepsOnlyCanonicalMounts covers the media read-scope boundary.
func TestParseZFSRootsKeepsOnlyCanonicalMounts(t *testing.T) {
	input := "36 29 0:33 / /tank rw - zfs tank rw\n" +
		"37 29 0:34 / /mnt/other rw - ext4 other rw\n" +
		"38 29 0:35 / /tank/photos rw - zfs tank/photos rw\n"
	roots, err := ParseZFSRoots(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(roots, ",") != "/tank,/tank/photos" {
		t.Fatalf("ZFS roots = %v", roots)
	}
}

// TestParseZFSRootsRejectsMalformedMounts covers malformed and unsafe mount records.
func TestParseZFSRootsRejectsMalformedMounts(t *testing.T) {
	for _, input := range []string{
		"missing separator\n",
		"36 29 0:33 / /tank rw - \n",
		"36 29 0:33 / rw - zfs tank rw\n",
		"36 29 0:33 / / rw - zfs tank rw\n",
		"36 29 0:33 / /tank/../etc rw - zfs tank rw\n",
	} {
		t.Run(strings.TrimSpace(input), func(t *testing.T) {
			if _, err := ParseZFSRoots(strings.NewReader(input)); err == nil {
				t.Fatal("malformed mount record was accepted")
			}
		})
	}
}
