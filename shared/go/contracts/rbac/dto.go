package rbac

type GetSubjectRolesRequest struct {
	Username string `json:"username" validate:"required,min=1,max=128"`
}

type GetSubjectRolesResponse struct {
	UID    string   `json:"uid"`
	Groups []string `json:"groups"`
}

type CheckPathRequest struct {
	UID  string `json:"uid" validate:"required,min=1,max=32"`
	Path string `json:"path" validate:"required,min=1,max=4096"`
}

type CheckSudoExecRequest struct {
	UID     string `json:"uid" validate:"required,min=1,max=32"`
	Command string `json:"command" validate:"required,min=1,max=4096"`
}

type DecisionResponse struct {
	Allowed bool `json:"allowed"`
}

type InvalidateCacheRequest struct {
	UID string `json:"uid,omitempty" validate:"max=32"`
}

type InvalidateCacheResponse struct {
	OK bool `json:"ok"`
}
