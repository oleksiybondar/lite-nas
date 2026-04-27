package pamauth

import (
	"errors"
	"strings"
)

var (
	errMissingServiceName = errors.New("pam service name is required")
	errPAMUnavailable     = errors.New("pam support is not compiled in")
)

// OutcomeCode identifies the stable auth outcome exposed by the service layer.
type OutcomeCode string

const (
	OutcomeAuthenticated        OutcomeCode = "authenticated"
	OutcomeInvalidCredentials   OutcomeCode = "invalid_credentials"
	OutcomePasswordChangeNeeded OutcomeCode = "password_change_required"
	OutcomeDenied               OutcomeCode = "denied"
	OutcomeServiceUnavailable   OutcomeCode = "service_unavailable"
)

// MessageLevel identifies the severity or purpose of an auth message.
type MessageLevel string

const (
	MessageLevelInfo  MessageLevel = "info"
	MessageLevelWarn  MessageLevel = "warn"
	MessageLevelError MessageLevel = "error"
)

// Message represents a structured PAM-originated or service-generated message.
type Message struct {
	Level MessageLevel
	Text  string
}

// AuthenticateRequest defines the input for a PAM-backed authentication
// attempt.
type AuthenticateRequest struct {
	Username string
	Password string
}

// PasswordChangeRequest defines the input for a PAM-backed password update
// flow.
type PasswordChangeRequest struct {
	Username    string
	OldPassword string
	NewPassword string
}

// Result captures the normalized auth outcome returned by the PAM adapter.
type Result struct {
	Code              OutcomeCode
	Username          string
	Messages          []Message
	CanChangePassword bool
}

// Authenticator defines the PAM-backed behavior required by the auth-service
// domain layer.
type Authenticator interface {
	Authenticate(request AuthenticateRequest) (Result, error)
	ChangePassword(request PasswordChangeRequest) (Result, error)
}

func validateServiceName(serviceName string) error {
	if strings.TrimSpace(serviceName) == "" {
		return errMissingServiceName
	}

	return nil
}

func newServiceUnavailableResult(username string, text string) Result {
	return Result{
		Code:     OutcomeServiceUnavailable,
		Username: username,
		Messages: []Message{{Level: MessageLevelError, Text: text}},
	}
}

func buildMessages(existing []Message, extra Message) []Message {
	if strings.TrimSpace(extra.Text) == "" {
		return existing
	}

	return append(existing, extra)
}
