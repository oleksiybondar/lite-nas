package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestFirstPlaceSkipsInvalidAddressFields keeps overlong values out of tags.
func TestFirstPlaceSkipsInvalidAddressFields(t *testing.T) {
	address := map[string]string{"road": strings.Repeat("x", 257), "city": " Copenhagen "}
	if got := firstPlace(address, "road", "city"); got != "Copenhagen" {
		t.Fatalf("selected place = %q", got)
	}
}

// TestCityIndexRejectsMissingArchive checks startup asset handling.
func TestCityIndexRejectsMissingArchive(t *testing.T) {
	_, err := loadCityIndex(filepath.Join(t.TempDir(), "missing.zip"))
	if err == nil {
		t.Fatal("missing city archive was accepted")
	}
}
