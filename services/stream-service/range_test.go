package main

import "testing"

// Requirements: stream-service-svc/FR-004.
func TestUserReaderRangeBounds(t *testing.T) {
	for _, test := range []struct {
		header string
		start  int64
		end    int64
		valid  bool
	}{
		{"bytes=2-4", 2, 4, true},
		{"bytes=-3", 3, 5, true},
		{"bytes=4-", 4, 5, true},
		{"bytes=6-", 0, 0, false},
		{"bytes=0-1,3-4", 0, 0, false},
	} {
		start, end, _, err := singleRange(test.header, 6)
		if (err == nil) != test.valid || test.valid && (start != test.start || end != test.end) {
			t.Fatalf("range %q = %d-%d, err=%v", test.header, start, end, err)
		}
	}
}
