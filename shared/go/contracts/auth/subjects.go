package auth

const (
	LoginRPCSubject   = "auth.rpc.login"
	RefreshRPCSubject = "auth.rpc.refresh"
	LogoutRPCSubject  = "auth.rpc.logout"
	// #nosec G101 -- NATS subject name, not a credential.
	ValidateAccessTokenRPCSubject = "auth.rpc.token.validate"
	SetLockdownRPCSubject         = "auth.rpc.lockdown.set"
	LockdownChangedEventSubject   = "auth.events.lockdown.changed"
)
