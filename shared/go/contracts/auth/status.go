package auth

// Status identifies the stable auth outcome exposed by the auth-service
// contract.
type Status string

const (
	StatusAuthenticated          Status = "authenticated"
	StatusDenied                 Status = "denied"
	StatusPasswordChangeRequired Status = "password_change_required"
)

// MessageLevel identifies the severity or purpose of an auth message returned
// by the auth-service contract.
type MessageLevel string

const (
	MessageLevelInfo  MessageLevel = "info"
	MessageLevelWarn  MessageLevel = "warn"
	MessageLevelError MessageLevel = "error"
)

// Message carries a structured auth-service message intended for service
// consumers.
type Message struct {
	Level MessageLevel `json:"level"`
	Text  string       `json:"text"`
}
