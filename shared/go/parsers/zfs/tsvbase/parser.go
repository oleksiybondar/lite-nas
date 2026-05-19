package tsvbase

import (
	"fmt"
	"strings"
)

// ParseRows parses tab-separated rows into header-mapped records.
//
// Empty lines are ignored. Every non-empty row must have exactly len(headers)
// fields.
func ParseRows(input string, headers []string) ([]map[string]string, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("headers must not be empty")
	}

	lines := strings.Split(input, "\n")
	rows := make([]map[string]string, 0, len(lines))

	for lineIndex, line := range lines {
		fields, shouldSkip, err := parseLineFields(line, lineIndex, len(headers))
		if err != nil {
			return nil, err
		}
		if shouldSkip {
			continue
		}

		rows = append(rows, mapRow(headers, fields))
	}

	return rows, nil
}

// parseLineFields tokenizes one input line and validates field count.
func parseLineFields(line string, lineIndex int, expectedFields int) ([]string, bool, error) {
	if strings.TrimSpace(line) == "" {
		return nil, true, nil
	}

	fields := strings.Split(line, "\t")
	if len(fields) != expectedFields {
		return nil, false, fmt.Errorf(
			"invalid TSV row %d: expected %d fields, got %d",
			lineIndex+1,
			expectedFields,
			len(fields),
		)
	}

	return fields, false, nil
}

// mapRow maps one parsed field slice to header-keyed values.
func mapRow(headers []string, fields []string) map[string]string {
	row := make(map[string]string, len(headers))
	for index := range headers {
		row[headers[index]] = strings.TrimSpace(fields[index])
	}
	return row
}
