package checks

import "context"

func canAccess(ctx context.Context, runner Runner, uid string, path string, operation string) (bool, error) {
	identity, err := ResolveIdentityByUID(ctx, runner, uid)
	if err != nil {
		return false, err
	}

	acl, err := readACLForPath(ctx, runner, path)
	if err != nil {
		return false, err
	}

	return evaluateAccess(identity, acl, operation), nil
}
