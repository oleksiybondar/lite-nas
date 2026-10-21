package rbac

type GetSubjectRolesRequest struct {
	Username string `json:"username" validate:"required,min=1,max=128"`
}

type GetSubjectRolesResponse struct {
	UID    string   `json:"uid"`
	Groups []string `json:"groups"`
}

type CheckPathRequest struct {
	UID  string `json:"uid" validate:"required,numeric,min=1,max=32"`
	Path string `json:"path" validate:"required,min=1,max=4096"`
}

// PathOperation names a path decision that the existing read, write, and exec RPCs cannot express.
type PathOperation string

const (
	PathOperationCreate   PathOperation = "create"
	PathOperationReplace  PathOperation = "replace"
	PathOperationDelete   PathOperation = "delete"
	PathOperationTraverse PathOperation = "traverse"
)

// CheckPathOperationRequest asks RBAC to evaluate one operation on an absolute path.
type CheckPathOperationRequest struct {
	UID       string        `json:"uid" validate:"required,numeric,min=1,max=32"`
	Path      string        `json:"path" validate:"required,min=1,max=4096"`
	Operation PathOperation `json:"operation" validate:"required,oneof=create replace delete traverse"`
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
