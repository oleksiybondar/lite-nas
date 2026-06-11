package checks

import "context"

// CanRead reports whether one UID has read access for one path.
func CanRead(ctx context.Context, runner Runner, uid string, path string) (bool, error) {
	return canAccess(ctx, runner, uid, path, "read")
}
