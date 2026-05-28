package checks

import (
	"fmt"
	"strconv"
	"strings"
)

// Identity describes user and group identity data resolved from id command output.
type Identity struct {
	UID          uint32
	GID          uint32
	Username     string
	PrimaryGroup string
	Groups       []string
}

func parseUint32(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(parsed), nil
}

func parseGroups(raw string) []string {
	parts := strings.Fields(strings.TrimSpace(raw))
	groups := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, group := range parts {
		trimmed := strings.TrimSpace(group)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		groups = append(groups, trimmed)
	}

	return groups
}

func parseIdentity(uidRaw, gidRaw, usernameRaw, primaryGroupRaw, groupsRaw string) (Identity, error) {
	uid, err := parseUint32(uidRaw)
	if err != nil {
		return Identity{}, fmt.Errorf("invalid uid output: %w", err)
	}

	gid, err := parseUint32(gidRaw)
	if err != nil {
		return Identity{}, fmt.Errorf("invalid gid output: %w", err)
	}

	username := strings.TrimSpace(usernameRaw)
	if username == "" {
		return Identity{}, fmt.Errorf("empty username output")
	}

	primaryGroup := strings.TrimSpace(primaryGroupRaw)
	if primaryGroup == "" {
		return Identity{}, fmt.Errorf("empty primary-group output")
	}

	groups := parseGroups(groupsRaw)
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
