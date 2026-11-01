// Package security_tasks contains browser-facing task request and response DTOs.
package security_tasks

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	antiviruscontract "lite-nas/shared/contracts/antivirus"
	backgroundtasks "lite-nas/shared/contracts/backgroundtasks"
	integritycontract "lite-nas/shared/contracts/litenasintegrity"
)

// StartAntivirusInput selects a folder for a configured eligible mount scan.
type StartAntivirusInput struct {
	Folder string `query:"folder" required:"true" minLength:"1" maxLength:"4096" doc:"Absolute folder path to scan."`
}

// EmptyInput represents a bodyless task-start request.
type EmptyInput struct{}

// TaskIDInput selects one pollable task attempt.
type TaskIDInput struct {
	TaskID string `path:"task_id" minLength:"32" maxLength:"32" pattern:"^[a-f0-9]{32}$"`
}

// TaskOutput wraps one task snapshot in gateway response metadata.
type TaskOutput struct{ Body TaskBody }

// TaskBody contains response metadata and a task snapshot.
type TaskBody struct {
	dto.ResponseMeta
	Data backgroundtasks.StatusCheckResponse `json:"data"`
}

// StartOutput wraps a newly queued task response.
type StartOutput struct{ Body StartBody }

// StartBody contains response metadata and the task ID.
type StartBody struct {
	dto.ResponseMeta
	Data backgroundtasks.StartResponse `json:"data"`
}

// RetryOutput wraps a replacement task response.
type RetryOutput struct{ Body RetryBody }

// AntivirusInfoOutput wraps antivirus engine and signature information.
type AntivirusInfoOutput struct{ Body AntivirusInfoBody }

// AntivirusInfoResponse gives the antivirus contract response a unique OpenAPI schema name.
type AntivirusInfoResponse antiviruscontract.InfoResponse

type AntivirusInfoBody struct {
	dto.ResponseMeta
	Data AntivirusInfoResponse `json:"data"`
}

// AntivirusUpdateOutput wraps the accepted signature update state.
type AntivirusUpdateOutput struct{ Body AntivirusUpdateBody }

// AntivirusUpdateBody contains response metadata and updater state.
type AntivirusUpdateBody struct {
	dto.ResponseMeta
	Data antiviruscontract.UpdateSignaturesResponse `json:"data"`
}

// IntegrityInfoOutput wraps AIDE baseline information.
type IntegrityInfoOutput struct{ Body IntegrityInfoBody }

// IntegrityInfoResponse gives the integrity contract response a unique OpenAPI schema name.
type IntegrityInfoResponse integritycontract.InfoResponse

type IntegrityInfoBody struct {
	dto.ResponseMeta
	Data IntegrityInfoResponse `json:"data"`
}

// RetryBody contains response metadata and the linked task attempts.
type RetryBody struct {
	dto.ResponseMeta
	Data backgroundtasks.RetryResponse `json:"data"`
}

// NewTaskOutput constructs the standard HTTP envelope for a task snapshot.
func NewTaskOutput(data backgroundtasks.StatusCheckResponse) *TaskOutput {
	return &TaskOutput{Body: TaskBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}

// NewStartOutput constructs the standard HTTP envelope for a started task.
func NewStartOutput(data backgroundtasks.StartResponse) *StartOutput {
	return &StartOutput{Body: StartBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}

// NewRetryOutput constructs the standard HTTP envelope for a retry response.
func NewRetryOutput(data backgroundtasks.RetryResponse) *RetryOutput {
	return &RetryOutput{Body: RetryBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}

// NewAntivirusInfoOutput constructs the standard HTTP envelope for antivirus information.
func NewAntivirusInfoOutput(data antiviruscontract.InfoResponse) *AntivirusInfoOutput {
	return &AntivirusInfoOutput{Body: AntivirusInfoBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: AntivirusInfoResponse(data)}}
}

// NewAntivirusUpdateOutput constructs the standard HTTP envelope for an accepted update.
func NewAntivirusUpdateOutput(data antiviruscontract.UpdateSignaturesResponse) *AntivirusUpdateOutput {
	return &AntivirusUpdateOutput{Body: AntivirusUpdateBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}

// NewIntegrityInfoOutput constructs the standard HTTP envelope for AIDE information.
func NewIntegrityInfoOutput(data integritycontract.InfoResponse) *IntegrityInfoOutput {
	return &IntegrityInfoOutput{Body: IntegrityInfoBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: IntegrityInfoResponse(data)}}
}
