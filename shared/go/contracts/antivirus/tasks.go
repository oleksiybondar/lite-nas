// Package antivirus defines NATS contracts for scans, signature updates, and status information.
package antivirus

import (
	"time"

	backgroundtasks "lite-nas/shared/contracts/backgroundtasks"
)

const (
	// StartScanSubject begins a scan of one authorized folder.
	StartScanSubject = "antivirus.background-tasks.rpc.start"
	// RetryScanSubject retries a completed scan using its saved folder.
	RetryScanSubject = "antivirus.background-tasks.rpc.retry"
	// ScanStatusSubject reads one scan attempt.
	ScanStatusSubject = "antivirus.background-tasks.rpc.status-check"
	// InfoSubject reads ClamAV signature database and engine information.
	InfoSubject = "antivirus.rpc.info"
	// UpdateSignaturesSubject starts an authorized ClamAV signature update.
	UpdateSignaturesSubject = "antivirus.rpc.signatures.update"
)

// InfoRequest authorizes one read of antivirus installation state.
type InfoRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// InfoResponse is a compact antivirus state with bounded diagnostic details.
type InfoResponse struct {
	State                  string     `json:"state"`
	UpdateState            string     `json:"update_state"`
	UpdateDetails          string     `json:"update_details,omitempty"`
	EngineVersion          string     `json:"engine_version,omitempty"`
	SignatureUpdatedAt     *time.Time `json:"signature_updated_at,omitempty"`
	SignatureDatabaseCount int        `json:"signature_database_count"`
	Details                string     `json:"details,omitempty"`
}

// UpdateSignaturesRequest authorizes one manual signature database update.
type UpdateSignaturesRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// UpdateSignaturesResponse reports the accepted background update state.
type UpdateSignaturesResponse struct {
	State string `json:"state" validate:"required,oneof=queued updating already_updating"`
}

// StartScanRequest identifies a folder and the human access token authorizing it.
type StartScanRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	Folder      string `json:"folder" validate:"required,min=1,max=4096"`
}

// TaskRequest identifies one task and its authorized caller.
type TaskRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	TaskID      string `json:"task_id" validate:"required,len=32,hexadecimal"`
}

// StartScanResponse identifies the queued antivirus scan.
type StartScanResponse struct {
	TaskID string `json:"task_id" validate:"required,len=32,hexadecimal"`
	State  string `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
}

// StatusResponse is the shared pollable task snapshot.
type StatusResponse = backgroundtasks.StatusCheckResponse

// RetryResponse identifies the replacement attempt.
type RetryResponse = backgroundtasks.RetryResponse
