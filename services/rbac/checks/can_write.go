package checks

import "context"

// CanWrite reports whether one UID has write access for one path.
func CanWrite(ctx context.Context, runner Runner, uid uint32, path string) (bool, error) {
	return canAccess(ctx, runner, uid, path, "write")
}
