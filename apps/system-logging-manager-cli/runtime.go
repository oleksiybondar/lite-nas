package main

import (
	"context"
	"fmt"
)

const (
	defaultConfigPath = "/etc/lite-nas/system-logging-manager-cli.conf"
	appName           = "system-logging-manager-cli"
)

// run starts the initial app stub runtime.
func run(ctx context.Context, _ []string) error {
	fmt.Println("hello " + appName)
	<-ctx.Done()
	return ctx.Err()
}
