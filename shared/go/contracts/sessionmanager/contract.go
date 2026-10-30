// Package sessionmanager defines authenticated login session control RPCs.
package sessionmanager

const (
	ListSubject      = "session-manager.rpc.list"
	TerminateSubject = "session-manager.rpc.terminate"
	LockSubject      = "session-manager.rpc.lock"
	UnlockSubject    = "session-manager.rpc.unlock"
)

// ListRequest authenticates a host session inventory request.
type ListRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
}

// ActionRequest identifies one loginctl session to manage.
type ActionRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Session     string `json:"session" validate:"required,max=64"`
}

// Session identifies a login session and its owning user.
type Session struct {
	ID   string `json:"id"`
	UID  string `json:"uid"`
	User string `json:"user"`
}

// ListResponse contains the current loginctl session inventory.
type ListResponse struct {
	Sessions []Session `json:"sessions"`
}

// ActionResponse confirms that loginctl completed an action successfully.
type ActionResponse struct {
	Completed bool `json:"completed"`
}
