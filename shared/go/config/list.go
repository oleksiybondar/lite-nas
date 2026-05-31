package config

import "strings"

// parseCommaSeparatedValues splits one comma-delimited config value into
// normalized non-empty items.
func parseCommaSeparatedValues(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		if normalized == "" {
			continue
		}
		values = append(values, normalized)
	}

	return values
}
