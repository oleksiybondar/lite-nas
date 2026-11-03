package auth

// RevokeMatchedSessionsRequest identifies browser sessions by verified account
// and client address for the Security Automation service RPC.
type RevokeMatchedSessionsRequest struct {
	Login    string `json:"login" validate:"required,max=128"`
	ClientIP string `json:"client_ip" validate:"required,ip"`
}

// RevokeMatchedSessionsResponse counts browser sessions revoked by Auth.
type RevokeMatchedSessionsResponse struct {
	Revoked int `json:"revoked"`
}
