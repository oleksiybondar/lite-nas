package auth

// SetLockdownRequest requests an explicit lockdown state transition and
// requires audit-oriented operator context.
type SetLockdownRequest struct {
	Lockdown  bool   `json:"lockdown"`
	Reason    string `json:"reason"`
	Initiator string `json:"initiator"`
}

// SetLockdownResponse returns the applied lockdown state together with the
// audit context used to request it.
type SetLockdownResponse struct {
	Lockdown  bool   `json:"lockdown"`
	Reason    string `json:"reason"`
	Initiator string `json:"initiator"`
}

// LockdownChangedEvent publishes an auth-service lockdown state transition.
type LockdownChangedEvent struct {
	Lockdown  bool   `json:"lockdown"`
	Reason    string `json:"reason"`
	Initiator string `json:"initiator"`
}
