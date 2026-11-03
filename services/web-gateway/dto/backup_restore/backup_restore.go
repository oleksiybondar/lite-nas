// Package backup_restore defines the browser-facing backup and restore schemas.
package backup_restore

import (
	"time"

	"lite-nas/services/web-gateway/dto"
)

type (
	EmptyInput       struct{}
	ApplicationInput struct {
		Application string `path:"application" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_-]*$"`
	}
)

type VersionInput struct {
	Application string `path:"application" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_-]*$"`
	Version     string `path:"version" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_.-]*$"`
}
type BackupInput struct {
	Application string `path:"application" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_-]*$"`
	Body        BackupLabels
}
type BackupAllInput struct {
	Body BackupLabels
}

// BackupLabels are the optional human-readable fields in a backup request body.
type BackupLabels struct {
	Name        string `json:"name,omitempty" maxLength:"128"`
	Description string `json:"description,omitempty" maxLength:"1024"`
}
type RestoreInput struct {
	Application string `path:"application" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_-]*$"`
	Version     string `path:"version" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_.-]*$"`
	Body        struct {
		Mode string `json:"mode" required:"true" enum:"replace,merge"`
	}
}

// RestoreByNameInput selects an application's newest archive with the supplied name.
type RestoreByNameInput struct {
	Application string `path:"application" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_-]*$"`
	Body        RestoreByNameBody
}

// RestoreAllByNameInput selects matching archives across configured applications.
type RestoreAllByNameInput struct {
	Body RestoreByNameBody
}

// RestoreByNameBody validates the exact archive name and reconciliation mode.
type RestoreByNameBody struct {
	Name string `json:"name" required:"true" minLength:"1" maxLength:"128"`
	Mode string `json:"mode" required:"true" enum:"replace,merge"`
}
type MergeStartInput struct {
	Application string `path:"application" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_-]*$"`
	Version     string `path:"version" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_.-]*$"`
}
type SessionInput struct {
	SessionID string `path:"session_id" minLength:"32" maxLength:"32" pattern:"^[a-fA-F0-9]{32}$"`
}
type MergeReadInput struct {
	SessionID string `path:"session_id" minLength:"32" maxLength:"32" pattern:"^[a-fA-F0-9]{32}$"`
	Path      string `query:"path" minLength:"1" maxLength:"4096"`
	Side      string `query:"side" enum:"snapshot,current,result"`
	Offset    int    `query:"offset" minimum:"0"`
}
type MergeWriteInput struct {
	SessionID string `path:"session_id" minLength:"32" maxLength:"32" pattern:"^[a-fA-F0-9]{32}$"`
	Body      struct {
		Path   string `json:"path" minLength:"1" maxLength:"4096"`
		Offset int    `json:"offset" minimum:"0"`
		Data   string `json:"data" maxLength:"87384"`
		Final  bool   `json:"final"`
		SHA256 string `json:"sha256,omitempty" minLength:"64" maxLength:"64" pattern:"^[a-fA-F0-9]{64}$"`
	}
}
type MergeSaveInput struct {
	SessionID string `path:"session_id" minLength:"32" maxLength:"32" pattern:"^[a-fA-F0-9]{32}$"`
	Body      BackupLabels
}

type (
	Output[T any] struct{ Body Body[T] }
	Body[T any]   struct {
		dto.ResponseMeta
		Data T `json:"data"`
	}
)

func NewOutput[T any](data T) *Output[T] {
	return &Output[T]{Body: Body[T]{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}
