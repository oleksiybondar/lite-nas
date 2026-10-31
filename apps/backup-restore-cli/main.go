package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"lite-nas/shared/config"
	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
)

const configPath = "/etc/lite-nas/backup-restore-cli.conf"

// main runs one authenticated backup request and exits.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run uses the current user's short-lived access token for every request.
func run(ctx context.Context, arguments []string) error {
	if len(arguments) == 0 {
		return fmt.Errorf("usage: backup-restore-cli applications|backed-up|versions APP|backup APP|backup all|diff APP VERSION|restore APP VERSION [replace|merge]|merge APP VERSION|merge resume ID|merge discard ID; backup accepts --name TEXT and --description TEXT")
	}
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	iniFile, err := config.LoadINI(reader)
	if err != nil {
		return err
	}
	messagingConfig, err := config.LoadMessagingConfig(iniFile)
	if err != nil {
		return err
	}
	loggingConfig, err := config.LoadLoggingConfig(iniFile)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientInfra("backup-restore-cli", loggingConfig, messagingConfig)
	if err != nil {
		return err
	}
	defer infra.Close()
	token, err := accessToken(ctx, infra.Client)
	if err != nil {
		return err
	}
	return executeCommand(ctx, infra.Client, token, arguments)
}

// executeCommand validates and dispatches one authenticated CLI command.
func executeCommand(ctx context.Context, client messaging.Client, token string, arguments []string) error {
	if len(arguments) == 0 {
		return fmt.Errorf("missing command")
	}
	var err error
	switch arguments[0] {
	case "applications", "backed-up":
		if len(arguments) != 1 {
			return fmt.Errorf("inventory command takes no arguments")
		}
		var response contract.ApplicationsResponse
		subject := contract.ApplicationsSubject
		if arguments[0] == "backed-up" {
			subject = contract.BackedUpSubject
		}
		if err := client.Request(ctx, subject, contract.ApplicationsRequest{AccessToken: token}, &response); err != nil {
			return err
		}
		for _, application := range response.Applications {
			fmt.Fprintln(os.Stdout, application)
		}
	case "versions", "backup":
		if len(arguments) < 2 || !validApplication(arguments[1]) || arguments[0] == "versions" && len(arguments) != 2 {
			return fmt.Errorf("expected one valid application ID")
		}
		name, description := "", ""
		if arguments[0] == "backup" {
			name, description, err = backupLabels(arguments[2:])
			if err != nil {
				return err
			}
		}
		if arguments[0] == "backup" && arguments[1] == "all" {
			var response contract.BackupAllResponse
			if err := client.Request(ctx, contract.BackupAllSubject, contract.BackupAllRequest{AccessToken: token, Name: name, Description: description}, &response); err != nil {
				return err
			}
			failed := false
			for _, result := range response.Results {
				fmt.Fprintln(os.Stdout, result.Application, result.Status, result.Version, result.Error)
				failed = failed || result.Status == "failed"
			}
			if failed {
				return fmt.Errorf("some application backups failed")
			}
			return nil
		}
		request := contract.ApplicationRequest{AccessToken: token, Application: arguments[1]}
		if arguments[0] == "versions" {
			var response contract.VersionsResponse
			if err := client.Request(ctx, contract.VersionsSubject, request, &response); err != nil {
				return err
			}
			for _, version := range response.Versions {
				fmt.Fprintln(os.Stdout, version)
			}
		} else {
			var response contract.BackupResponse
			if err := client.Request(ctx, contract.BackupSubject, contract.BackupRequest{AccessToken: token, Application: request.Application, Name: name, Description: description}, &response); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, response.Version)
		}
	case "diff":
		if len(arguments) != 3 || !validApplication(arguments[1]) || !validVersion(arguments[2]) {
			return fmt.Errorf("expected application ID and version")
		}
		var response contract.DiffResponse
		request := contract.VersionRequest{AccessToken: token, Application: arguments[1], Version: arguments[2]}
		if err := client.Request(ctx, contract.DiffSubject, request, &response); err != nil {
			return err
		}
		for _, difference := range response.Differences {
			fmt.Fprintln(os.Stdout, difference.Status, difference.Path)
		}
	case "restore":
		if len(arguments) < 3 || len(arguments) > 4 || !validApplication(arguments[1]) || !validVersion(arguments[2]) {
			return fmt.Errorf("expected application ID, version, and optional replace|merge mode")
		}
		mode := "replace"
		if len(arguments) == 4 {
			mode = arguments[3]
		}
		if mode != "replace" && mode != "merge" {
			return fmt.Errorf("invalid restore mode")
		}
		var response contract.RestoreResponse
		request := contract.RestoreRequest{AccessToken: token, Application: arguments[1], Version: arguments[2], Mode: mode}
		if err := client.Request(ctx, contract.RestoreSubject, request, &response); err != nil {
			return err
		}
		if !response.Restored {
			return fmt.Errorf("restore was not confirmed")
		}
		if len(response.Warnings) > 0 {
			return fmt.Errorf("files restored, but service restart failed: %v", response.Warnings)
		}
		fmt.Fprintln(os.Stdout, "restored", arguments[1], arguments[2])
	case "merge":
		return runMerge(ctx, client, token, arguments[1:])
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
	return nil
}

// backupLabels validates optional human labels before sending a backup request.
func backupLabels(arguments []string) (string, string, error) {
	name, description := "", ""
	seen := map[string]bool{}
	for len(arguments) > 0 {
		if len(arguments) < 2 || seen[arguments[0]] {
			return "", "", fmt.Errorf("invalid backup label arguments")
		}
		flag, value := arguments[0], arguments[1]
		if value == "" || value[0] == '-' {
			return "", "", fmt.Errorf("invalid backup label value")
		}
		switch flag {
		case "--name":
			if len(value) > 128 {
				return "", "", fmt.Errorf("backup name too long")
			}
			name = value
		case "--description":
			if len(value) > 1024 {
				return "", "", fmt.Errorf("backup description too long")
			}
			description = value
		default:
			return "", "", fmt.Errorf("unknown backup option %q", flag)
		}
		seen[flag] = true
		arguments = arguments[2:]
	}
	return name, description, nil
}

// validApplication rejects option-shaped and malformed application arguments.
func validApplication(value string) bool {
	if value == "" || len(value) > 64 || value[0] == '-' {
		return false
	}
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}
	return true
}

// validVersion accepts only the version characters emitted by the service store.
func validVersion(value string) bool {
	if value == "" || len(value) > 64 || value[0] == '-' {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || char == 'T' || char == 'Z' || char == '.' || char == '-') {
			return false
		}
	}
	return true
}
