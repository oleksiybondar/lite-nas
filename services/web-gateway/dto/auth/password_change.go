package auth

// PasswordChangeStartInput validates an on-demand change request.
type PasswordChangeStartInput struct {
	Authorization     string `header:"Authorization"`
	AccessTokenCookie string `cookie:"lite-nas-at"`
	Body              struct{}
}

// PasswordChangeReplyInput validates one PAM answer.
type PasswordChangeReplyInput struct {
	Body struct {
		ConversationID string `json:"conversation_id" required:"true" minLength:"1" maxLength:"128"`
		Sequence       int32  `json:"sequence" required:"true" minimum:"1"`
		Answer         string `json:"answer" required:"true" minLength:"1" maxLength:"4096"`
	}
}
