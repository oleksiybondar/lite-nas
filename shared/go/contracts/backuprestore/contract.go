// Package backuprestore defines the on-demand backup RPC boundary.
package backuprestore

const (
	// Merge subjects manage the one process-local text merge session.
	MergeStartSubject   = "backup-restore.merge.start"
	MergeReadSubject    = "backup-restore.merge.read"
	MergeStatusSubject  = "backup-restore.merge.status"
	MergeWriteSubject   = "backup-restore.merge.write"
	MergeApplySubject   = "backup-restore.merge.apply"
	MergeSaveSubject    = "backup-restore.merge.save"
	MergeDiscardSubject = "backup-restore.merge.discard"
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

// MergeStartRequest selects a snapshot to compare against current files.
type MergeStartRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
	Version     string `json:"version" validate:"required,max=64"`
}

// MergeFile describes one editable regular file.
type MergeFile struct {
	Path         string `json:"path"`
	SnapshotSize int    `json:"snapshot_size"`
	CurrentSize  int    `json:"current_size"`
	Status       string `json:"status"`
}

// MergeStartResponse returns the session ID and editable files.
type MergeStartResponse struct {
	SessionID      string      `json:"session_id"`
	Files          []MergeFile `json:"files"`
	ApplyAvailable bool        `json:"apply_available"`
}

// MergeReadRequest selects one chunk of a session input or result.
type MergeReadRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	SessionID   string `json:"session_id" validate:"required,len=32,hexadecimal"`
	Path        string `json:"path" validate:"required,max=4096"`
	Side        string `json:"side" validate:"oneof=snapshot current result"`
	Offset      int    `json:"offset" validate:"min=0"`
}

// MergeReadResponse carries a base64 chunk and full-content SHA-256 checksum.
type MergeReadResponse struct {
	Data       string `json:"data"`
	NextOffset int    `json:"next_offset"`
	Done       bool   `json:"done"`
	SHA256     string `json:"sha256"`
}

// MergeWriteRequest uploads one sequential chunk of resolved content.
type MergeWriteRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	SessionID   string `json:"session_id" validate:"required,len=32,hexadecimal"`
	Path        string `json:"path" validate:"required,max=4096"`
	Offset      int    `json:"offset" validate:"min=0"`
	Data        string `json:"data" validate:"max=87384"`
	Final       bool   `json:"final"`
	SHA256      string `json:"sha256,omitempty" validate:"omitempty,len=64,hexadecimal"`
}

// MergeActionRequest selects a session for apply or discard.
type MergeActionRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	SessionID   string `json:"session_id" validate:"required,len=32,hexadecimal"`
}

// MergeSaveRequest supplies labels for a new snapshot made from staged content.
type MergeSaveRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	SessionID   string `json:"session_id" validate:"required,len=32,hexadecimal"`
	Name        string `json:"name,omitempty" validate:"max=128"`
	Description string `json:"description,omitempty" validate:"max=1024"`
}

// MergeWriteResponse confirms one accepted chunk.
type MergeWriteResponse struct {
	NextOffset int `json:"next_offset"`
}

// MergeActionResponse confirms a completed apply or discard.
type MergeActionResponse struct {
	Completed bool     `json:"completed"`
	Warnings  []string `json:"warnings,omitempty"`
}

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
	Restored bool     `json:"restored"`
	Warnings []string `json:"warnings,omitempty"`
}
