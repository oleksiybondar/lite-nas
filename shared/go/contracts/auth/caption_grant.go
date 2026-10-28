package auth

import "time"

// CaptionGrantIssueRequest requests read access to one fixed set of gallery files.
type CaptionGrantIssueRequest struct {
	UserAccessToken string   `json:"user_access_token" validate:"required,min=1,max=8192"`
	JobID           string   `json:"job_id" validate:"required,uuid"`
	PathHashes      []string `json:"path_hashes" validate:"required,min=1,max=10000,dive,len=64,hexadecimal"`
}

// CaptionGrantIssueResponse returns a File Service-only opaque grant.
type CaptionGrantIssueResponse struct {
	Grant     string    `json:"grant,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// CaptionGrantValidateRequest checks one exact source path against a grant.
type CaptionGrantValidateRequest struct {
	Grant string `json:"grant" validate:"required,min=1,max=128"`
	Path  string `json:"path" validate:"required,min=1,max=4096"`
}

// CaptionGrantValidateResponse identifies the user only for an allowed path.
type CaptionGrantValidateResponse struct {
	Valid    bool   `json:"valid"`
	Username string `json:"username,omitempty"`
}

// CaptionGrantRevokeRequest invalidates one grant held by Moments.
type CaptionGrantRevokeRequest struct {
	Grant string `json:"grant" validate:"required,min=1,max=128"`
}

// CaptionGrantRevokeResponse reports whether the grant was removed.
type CaptionGrantRevokeResponse struct {
	Revoked bool `json:"revoked"`
}
