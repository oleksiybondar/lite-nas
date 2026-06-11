package identity

import (
	"fmt"
	"strings"
)

// Parse parses default id output, for example:
// uid=1002(testuser) gid=1002(testgroup) groups=1002(testgroup),27(sudo)
func Parse(input string) (Identity, error) {
	segments := strings.Fields(strings.TrimSpace(input))
	uidSegment, gidSegment, groupsSegment, err := parseSegments(segments)
	if err != nil {
		return Identity{}, err
	}

	uid, username, err := parseFieldIDName(uidSegment, "uid")
	if err != nil {
		return Identity{}, err
	}

	gid, primaryGroup, err := parseFieldIDName(gidSegment, "gid")
	if err != nil {
		return Identity{}, err
	}

	groups, err := parseGroupsList(groupsSegment)
	if err != nil {
		return Identity{}, err
	}
	if len(groups) == 0 {
		groups = []string{primaryGroup}
	}

	return Identity{
		UID:          uid,
		GID:          gid,
		Username:     username,
		PrimaryGroup: primaryGroup,
		Groups:       groups,
	}, nil
}

func parseSegments(segments []string) (string, string, string, error) {
	if len(segments) == 0 {
		return "", "", "", fmt.Errorf("empty id output")
	}
	if len(segments) != 3 {
		return "", "", "", fmt.Errorf("invalid id output segment count")
	}
	return segments[0], segments[1], segments[2], nil
}

func parseFieldIDName(segment string, field string) (string, string, error) {
	value, err := parseFieldValue(segment, field)
	if err != nil {
		return "", "", err
	}

	return parseIDName(value, field)
}

func parseIDName(value string, field string) (string, string, error) {
	parts := strings.SplitN(value, "(", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid %s pair field", field)
	}

	idValue, err := parseIDPart(parts[0])
	if err != nil {
		return "", "", fmt.Errorf("empty %s id", field)
	}

	name, err := parseNamePart(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("invalid %s pair field", field)
	}

	return idValue, name, nil
}

func parseIDPart(raw string) (string, error) {
	idValue := strings.TrimSpace(raw)
	if idValue == "" {
		return "", fmt.Errorf("id missing")
	}
	return idValue, nil
}

func parseNamePart(raw string) (string, error) {
	namePart := strings.TrimSpace(raw)
	if !strings.HasSuffix(namePart, ")") {
		return "", fmt.Errorf("name suffix missing")
	}

	name := strings.TrimSpace(strings.TrimSuffix(namePart, ")"))
	if name == "" {
		return "", fmt.Errorf("name missing")
	}

	return name, nil
}

func parseGroupsList(segment string) ([]string, error) {
	value, err := parseGroupsSegmentValue(segment)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return nil, nil
	}

	entries := strings.Split(value, ",")
	groups := make([]string, 0, len(entries))

	for _, entry := range entries {
		_, name, err := parseIDName(entry, "groups")
		if err != nil {
			return nil, err
		}
		groups = append(groups, name)
	}

	return groups, nil
}

func parseGroupsSegmentValue(segment string) (string, error) {
	value, err := parseFieldValue(segment, "groups")
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", nil
	}
	return value, nil
}

func parseFieldValue(segment string, field string) (string, error) {
	if segment == "" {
		return "", fmt.Errorf("missing %s field", field)
	}

	prefix := field + "="
	if !strings.HasPrefix(segment, prefix) {
		return "", fmt.Errorf("unexpected %s segment", field)
	}

	return strings.TrimPrefix(segment, prefix), nil
}
