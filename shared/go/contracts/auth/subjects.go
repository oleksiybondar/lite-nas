package auth

const (
	LoginRPCSubject               = "auth.rpc.login"
	RefreshRPCSubject             = "auth.rpc.refresh"
	LogoutRPCSubject              = "auth.rpc.logout"
	PasswordChangeStartRPCSubject = "auth.rpc.password_change.start"
	PasswordChangeReplyRPCSubject = "auth.rpc.password_change.reply"
	// ServiceTokenLoginRPCSubject mints a long-lived service-to-service token pair.
	// #nosec G101 -- NATS subject name, not a credential.
	ServiceTokenLoginRPCSubject = "auth.rpc.service_token.login"
	// ServiceTokenRefreshRPCSubject rotates a service-to-service token pair.
	// #nosec G101 -- NATS subject name, not a credential.
	ServiceTokenRefreshRPCSubject = "auth.rpc.service_token.refresh"
	// CaptionGrantIssueRPCSubject creates a job-bound File Service read grant.
	CaptionGrantIssueRPCSubject = "auth.rpc.caption_grant.issue"
	// CaptionGrantValidateRPCSubject checks a grant for one File Service path.
	CaptionGrantValidateRPCSubject = "auth.rpc.caption_grant.validate"
	// CaptionGrantRevokeRPCSubject invalidates a completed or cancelled grant.
	CaptionGrantRevokeRPCSubject = "auth.rpc.caption_grant.revoke"
	// #nosec G101 -- NATS subject name, not a credential.
	ValidateAccessTokenRPCSubject = "auth.rpc.token.validate"
	SetLockdownRPCSubject         = "auth.rpc.lockdown.set"
	LockdownChangedEventSubject   = "auth.events.lockdown.changed"
)
