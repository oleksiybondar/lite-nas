package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"lite-nas/shared/config"
	sharedcontracts "lite-nas/shared/contracts"
	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/fileio"
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

// run validates local authorization before reading the CLI identity credentials.
func run(ctx context.Context, arguments []string) error {
	if !localAdministrator() {
		return fmt.Errorf("root or host administrator required")
	}
	if len(arguments) == 0 {
		return fmt.Errorf("usage: backup-restore-cli applications|backed-up|versions APP|backup APP|backup all|diff APP VERSION|restore APP VERSION [replace|merge]")
	}
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	iniFile, err := config.LoadINI(reader)
	if err != nil {
		return err
	}
	settings, err := config.LoadSharedConfig(iniFile)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.AppBackupRestoreCLI, settings.Logging, settings.Messaging, settings.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	if err := infra.AuthTokenManager.Login(ctx); err != nil {
		return err
	}
	token, _, err := infra.AuthTokenManager.Token()
	if err != nil {
		return err
	}
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
		if err := infra.Client.Request(ctx, subject, contract.ApplicationsRequest{AccessToken: token}, &response); err != nil {
			return err
		}
		for _, application := range response.Applications {
			fmt.Fprintln(os.Stdout, application)
		}
	case "versions", "backup":
		if len(arguments) != 2 || !validApplication(arguments[1]) {
			return fmt.Errorf("expected one valid application ID")
		}
		if arguments[0] == "backup" && arguments[1] == "all" {
			var response contract.BackupAllResponse
			if err := infra.Client.Request(ctx, contract.BackupAllSubject, contract.ApplicationsRequest{AccessToken: token}, &response); err != nil {
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
			if err := infra.Client.Request(ctx, contract.VersionsSubject, request, &response); err != nil {
				return err
			}
			for _, version := range response.Versions {
				fmt.Fprintln(os.Stdout, version)
			}
		} else {
			var response contract.BackupResponse
			if err := infra.Client.Request(ctx, contract.BackupSubject, request, &response); err != nil {
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
		if err := infra.Client.Request(ctx, contract.DiffSubject, request, &response); err != nil {
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
		if err := infra.Client.Request(ctx, contract.RestoreSubject, request, &response); err != nil {
			return err
		}
		if !response.Restored {
			return fmt.Errorf("restore was not confirmed")
		}
		fmt.Fprintln(os.Stdout, "restored", arguments[1], arguments[2])
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
	return nil
}

// localAdministrator accepts root or membership in an existing host admin group.
func localAdministrator() bool {
	if os.Geteuid() == 0 {
		return true
	}
	groups, err := os.Getgroups()
	if err != nil {
		return false
	}
	for _, name := range []string{"admin", "sudo"} {
		group, err := user.LookupGroup(name)
		if err != nil {
			continue
		}
		id, err := strconv.Atoi(group.Gid)
		if err != nil {
			continue
		}
		for _, member := range groups {
			if member == id {
				return true
			}
		}
	}
	return false
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
