package auth

const (
	LoginRPCSubject   = "auth.rpc.login"
	RefreshRPCSubject = "auth.rpc.refresh"
	LogoutRPCSubject  = "auth.rpc.logout"
	// ServiceTokenLoginRPCSubject mints a long-lived service-to-service token pair.
	// #nosec G101 -- NATS subject name, not a credential.
	ServiceTokenLoginRPCSubject = "auth.rpc.service_token.login"
	// ServiceTokenRefreshRPCSubject rotates a service-to-service token pair.
	// #nosec G101 -- NATS subject name, not a credential.
	ServiceTokenRefreshRPCSubject = "auth.rpc.service_token.refresh"
	// #nosec G101 -- NATS subject name, not a credential.
	ValidateAccessTokenRPCSubject = "auth.rpc.token.validate"
	SetLockdownRPCSubject         = "auth.rpc.lockdown.set"
	LockdownChangedEventSubject   = "auth.events.lockdown.changed"
)
