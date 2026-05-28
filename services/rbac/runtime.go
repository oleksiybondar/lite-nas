package main

import "context"

// run assembles and starts the RBAC runtime.
func run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
