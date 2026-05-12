package main

import (
	"context"
	"fmt"
)

const (
	packagedConfigPath = "/etc/lite-nas/security-logging-manager.conf"
	serviceName        = "security-logging-manager"
)

// run starts the initial service stub runtime.
func run(ctx context.Context) error {
	fmt.Println("hello " + serviceName)
	<-ctx.Done()
	return ctx.Err()
}
