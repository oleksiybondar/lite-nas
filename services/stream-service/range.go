package main

import "lite-nas/shared/httprange"

// singleRange parses the supported byte range for original media responses.
func singleRange(header string, size int64) (int64, int64, bool, error) {
	return httprange.ParseSingle(header, size)
}
