package main

import (
	"errors"
	"strconv"
	"strings"
)

// singleRange parses the one RFC 9110 byte range supported for original media.
func singleRange(header string, size int64) (int64, int64, bool, error) {
	if header == "" {
		return 0, size - 1, false, nil
	}
	if !strings.HasPrefix(header, "bytes=") || strings.Contains(header, ",") || size == 0 {
		return 0, 0, false, errors.New("invalid range")
	}
	parts := strings.Split(strings.TrimPrefix(header, "bytes="), "-")
	if len(parts) != 2 {
		return 0, 0, false, errors.New("invalid range")
	}
	if parts[0] == "" {
		count, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || count <= 0 {
			return 0, 0, false, errors.New("invalid suffix")
		}
		if count > size {
			count = size
		}
		return size - count, size - 1, true, nil
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false, errors.New("invalid start")
	}
	end := size - 1
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false, errors.New("invalid end")
		}
		if end >= size {
			end = size - 1
		}
	}
	return start, end, true, nil
}
