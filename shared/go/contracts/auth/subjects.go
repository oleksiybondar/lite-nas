package auth

const (
	LoginRPCSubject               = "auth.rpc.login"
	RefreshRPCSubject             = "auth.rpc.refresh"
	LogoutRPCSubject              = "auth.rpc.logout"
	ValidateAccessTokenRPCSubject = "auth.rpc.token.validate"
	SetLockdownRPCSubject         = "auth.rpc.lockdown.set"
	LockdownChangedEventSubject   = "auth.events.lockdown.changed"
)
