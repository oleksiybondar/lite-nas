package identity

// Identity describes user and group identity data resolved from id command output.
type Identity struct {
	UID          string
	GID          string
	Username     string
	PrimaryGroup string
	Groups       []string
}
