package auth

import "time"

// PasswordConversation is a pending PAM prompt; answers must never be logged.
type PasswordConversation struct {
	ID        string    `json:"id"`
	Sequence  int32     `json:"sequence"`
	Prompt    string    `json:"prompt"`
	Secret    bool      `json:"secret"`
	ExpiresAt time.Time `json:"expires_at"`
	Messages  []Message `json:"messages,omitempty"`
}

// PasswordChangeStartRequest starts a change for the verified token subject.
type PasswordChangeStartRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// PasswordChangeReplyRequest answers one pending PAM prompt.
type PasswordChangeReplyRequest struct {
	ConversationID string `json:"conversation_id" validate:"required,min=1,max=128"`
	Sequence       int32  `json:"sequence" validate:"required,min=1"`
	Answer         string `json:"answer" validate:"required,min=1,max=4096"`
}
