package checks

import "context"

// CanExec reports whether one UID has execute access for one path.
func CanExec(ctx context.Context, runner Runner, uid string, path string) (bool, error) {
	return canAccess(ctx, runner, uid, path, "execute")
}
