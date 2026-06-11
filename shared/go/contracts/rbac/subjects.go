package rbac

const (
	GetSubjectRolesRPCSubject = "rbac.rpc.subject.roles.get"
	CanReadPathRPCSubject     = "rbac.rpc.path.read.check"
	CanWritePathRPCSubject    = "rbac.rpc.path.write.check"
	CanExecPathRPCSubject     = "rbac.rpc.path.exec.check"
	CanSudoExecRPCSubject     = "rbac.rpc.command.sudo.check"
	InvalidateCacheRPCSubject = "rbac.rpc.cache.invalidate"
)
