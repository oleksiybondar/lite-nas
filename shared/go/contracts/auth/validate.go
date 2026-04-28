package auth

// ValidateAccessTokenRequest requests live validation of an access token.
type ValidateAccessTokenRequest struct {
	AccessToken string `json:"access_token"`
}

// ValidateAccessTokenResponse returns the current credibility of the submitted
// access token and any resolved auth context.
type ValidateAccessTokenResponse struct {
	Valid    bool      `json:"valid"`
	Status   Status    `json:"status,omitempty"`
	Username string    `json:"username,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}
