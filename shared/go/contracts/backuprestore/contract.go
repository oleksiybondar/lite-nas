// Package backuprestore defines the on-demand backup RPC boundary.
package backuprestore

const (
	// ApplicationsSubject lists configured application IDs.
	ApplicationsSubject = "backup-restore.applications"
	// BackedUpSubject lists applications with stored history.
	BackedUpSubject = "backup-restore.backed-up"
	// VersionsSubject lists completed versions for one application.
	VersionsSubject = "backup-restore.versions"
	// BackupSubject creates one application snapshot.
	BackupSubject = "backup-restore.backup"
	// BackupAllSubject visits each configured application once.
	BackupAllSubject = "backup-restore.backup-all"
	// DiffSubject compares a version with the live application files.
	DiffSubject = "backup-restore.diff"
	// RestoreSubject restores a selected version.
	RestoreSubject = "backup-restore.restore"
)

// ApplicationRequest selects a configured application without supplying host paths.
type ApplicationRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
}

// BackupRequest supplies optional labels for a new application snapshot.
type BackupRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
	Name        string `json:"name,omitempty" validate:"max=128"`
	Description string `json:"description,omitempty" validate:"max=1024"`
}

// BackupAllRequest applies the same optional labels to each application snapshot.
type BackupAllRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Name        string `json:"name,omitempty" validate:"max=128"`
	Description string `json:"description,omitempty" validate:"max=1024"`
}

// RestoreRequest selects a stored version and its directory reconciliation mode.
type RestoreRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
	Version     string `json:"version" validate:"required,max=64"`
	Mode        string `json:"mode" validate:"oneof=replace merge"`
}

// VersionRequest selects one stored version for read-only comparison.
type VersionRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
	Version     string `json:"version" validate:"required,max=64"`
}

// Difference identifies one changed path without disclosing its contents.
type Difference struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// DiffResponse contains changed paths relative to the selected version.
type DiffResponse struct {
	Differences []Difference `json:"differences"`
}

// ApplicationsRequest authenticates an inventory query.
type ApplicationsRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
}

// ApplicationsResponse contains configured application IDs.
type ApplicationsResponse struct {
	Applications []string `json:"applications"`
}

// VersionsResponse contains completed versions, newest first.
type VersionsResponse struct {
	Versions []string `json:"versions"`
}

// BackupResponse identifies the newly created version.
type BackupResponse struct {
	Version string `json:"version"`
}

// BackupAllResult reports one completed, skipped, or failed application.
type BackupAllResult struct {
	Application string `json:"application"`
	Status      string `json:"status"`
	Version     string `json:"version,omitempty"`
	Error       string `json:"error,omitempty"`
}

// BackupAllResponse contains the result for every configured application.
type BackupAllResponse struct {
	Results []BackupAllResult `json:"results"`
}

// RestoreResponse confirms a completed restore.
type RestoreResponse struct {
	Restored bool `json:"restored"`
}
