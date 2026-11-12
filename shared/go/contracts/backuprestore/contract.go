// Package backuprestore defines the on-demand backup RPC boundary.
package backuprestore

import "time"

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
	// MetadataSubject reads labels and attribution stored inside one archive.
	MetadataSubject = "backup-restore.metadata"
	// DeleteSubject permanently removes one stored archive.
	DeleteSubject = "backup-restore.delete"
	// DeleteAllByNameSubject removes every archive with an exact snapshot name.
	DeleteAllByNameSubject = "backup-restore.delete-all-by-name"
	// BackupSubject creates one application snapshot.
	BackupSubject = "backup-restore.backup"
	// BackupAllSubject visits each configured application once.
	BackupAllSubject = "backup-restore.backup-all"
	// BackupTaskStartSubject starts an asynchronous backup task.
	BackupTaskStartSubject = "backup-restore.backup-task.start"
	// BackupTaskStatusSubject reads one asynchronous backup task.
	BackupTaskStatusSubject = "backup-restore.backup-task.status"
	// DiffSubject compares a version with the live application files.
	DiffSubject = "backup-restore.diff"
	// RestoreSubject restores a selected version.
	RestoreSubject = "backup-restore.restore"
	// RestoreByNameSubject restores the newest version with a name for one application.
	RestoreByNameSubject = "backup-restore.restore-by-name"
	// RestoreAllByNameSubject restores every configured application with a matching name.
	RestoreAllByNameSubject = "backup-restore.restore-all-by-name"
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
	Mode        string `json:"mode" validate:"required,oneof=replace merge"`
}

// RestoreByNameRequest selects one application's newest matching archive.
type RestoreByNameRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
	Name        string `json:"name" validate:"required,max=128"`
	Mode        string `json:"mode" validate:"required,oneof=replace merge"`
}

// RestoreAllByNameRequest selects matching archives across configured applications.
type RestoreAllByNameRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Name        string `json:"name" validate:"required,max=128"`
	Mode        string `json:"mode" validate:"required,oneof=replace merge"`
}

// VersionRequest selects one stored version for read-only comparison.
type VersionRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application" validate:"required,max=64"`
	Version     string `json:"version" validate:"required,max=64"`
}

// DeleteAllByNameRequest removes snapshots with one exact name across applications.
type DeleteAllByNameRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Name        string `json:"name" validate:"required,max=128"`
}

// DeleteAllByNameResult reports one matching snapshot's deletion outcome.
type DeleteAllByNameResult struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Status      string `json:"status" validate:"required,oneof=deleted failed"`
	Error       string `json:"error,omitempty" validate:"omitempty,max=2048"`
}

// DeleteAllByNameResponse reports every matched snapshot, including partial failures.
type DeleteAllByNameResponse struct {
	Results []DeleteAllByNameResult `json:"results"`
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

// SnapshotSummary identifies a stored snapshot by its stable version and optional display name.
type SnapshotSummary struct {
	Version string `json:"version"`
	Name    string `json:"name,omitempty"`
}

// VersionsResponse contains completed versions, newest first, with their optional names.
type VersionsResponse struct {
	Versions  []string          `json:"versions"`
	Snapshots []SnapshotSummary `json:"snapshots"`
}

// BackupResponse identifies the newly created version.
type BackupResponse struct {
	Version string `json:"version"`
}

// BackupMetadataResponse returns descriptive data embedded in an archive.
type BackupMetadataResponse struct {
	Application  string   `json:"application"`
	Version      string   `json:"version"`
	RestartUnits []string `json:"restart_units,omitempty"`
	Name         string   `json:"name,omitempty"`
	Description  string   `json:"description,omitempty"`
	CreatedAt    string   `json:"created_at"`
	CreatedBy    string   `json:"created_by"`
}

// DeleteResponse confirms removal of one stored archive.
type DeleteResponse struct {
	Deleted bool `json:"deleted"`
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

// BackupTaskRequest selects one queued or completed background backup task.
type BackupTaskRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	TaskID      string `json:"task_id" validate:"required,len=32,hexadecimal"`
}

// BackupTaskStartRequest starts one configured application backup or the all-applications task.
type BackupTaskStartRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Application string `json:"application,omitempty" validate:"omitempty,max=64,required_without=All,excluded_with=All"`
	All         bool   `json:"all,omitempty"`
	Name        string `json:"name,omitempty" validate:"max=128"`
	Description string `json:"description,omitempty" validate:"max=1024"`
}

// BackupTaskStartResponse identifies an accepted asynchronous backup task.
type BackupTaskStartResponse struct {
	TaskID string `json:"task_id" validate:"required,len=32,hexadecimal"`
	State  string `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
}

// BackupTaskProgress reports measurable completion across configured selections or targets.
type BackupTaskProgress struct {
	Percent *int   `json:"percent,omitempty" validate:"omitempty,min=0,max=100"`
	Message string `json:"message,omitempty" validate:"omitempty,max=512"`
}

// BackupTaskStatusResponse is a pollable snapshot of one background backup.
type BackupTaskStatusResponse struct {
	TaskID      string              `json:"task_id" validate:"required,len=32,hexadecimal"`
	Application string              `json:"application" validate:"required,max=64"`
	State       string              `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
	StartedAt   time.Time           `json:"started_at" validate:"required"`
	UpdatedAt   time.Time           `json:"updated_at" validate:"required"`
	FinishedAt  *time.Time          `json:"finished_at,omitempty"`
	Progress    *BackupTaskProgress `json:"progress,omitempty"`
	Version     string              `json:"version,omitempty"`
	Results     []BackupAllResult   `json:"results,omitempty"`
	Error       string              `json:"error,omitempty" validate:"omitempty,max=2048"`
}

// RestoreResponse confirms a completed restore.
type RestoreResponse struct {
	Restored bool     `json:"restored"`
	Warnings []string `json:"warnings,omitempty"`
}

// RestoreByNameResponse identifies the version chosen by a name-based restore.
type RestoreByNameResponse struct {
	Version  string   `json:"version"`
	Restored bool     `json:"restored"`
	Warnings []string `json:"warnings,omitempty"`
}

// RestoreAllByNameResult reports one configured application's outcome.
type RestoreAllByNameResult struct {
	Application string   `json:"application"`
	Status      string   `json:"status"`
	Version     string   `json:"version,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// RestoreAllByNameResponse includes completed, skipped, and failed applications.
type RestoreAllByNameResponse struct {
	Results []RestoreAllByNameResult `json:"results"`
}
