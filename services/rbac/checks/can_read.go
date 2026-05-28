package checks

import "context"

// CanRead reports whether one UID has read access for one path.
func CanRead(ctx context.Context, runner Runner, uid uint32, path string) (bool, error) {
	return canAccess(ctx, runner, uid, path, "read")
}
