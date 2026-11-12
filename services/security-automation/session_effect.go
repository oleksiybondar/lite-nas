package main

import (
	"context"
	"errors"
	"regexp"
	"time"

	contract "lite-nas/shared/contracts/sessionmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/servicetoken"
)

var (
	verifiedSessionID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
	verifiedTTY       = regexp.MustCompile(`^[A-Za-z0-9_./-]{1,64}$`)
)

// sessionTermination verifies a unique active loginctl owner and terminal
// before requesting one irreversible session termination.
type sessionTermination func(context.Context, string, string) (string, bool, error)

// newSessionTermination binds the existing Session Manager RPC contracts to
// the automation service's Security-role token.
func newSessionTermination(client messaging.Client, service string) (sessionTermination, error) {
	manager, err := servicetoken.NewManager(client, servicetoken.Options{Service: service})
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, username, tty string) (string, bool, error) {
		if !incidentUsername.MatchString(username) || !verifiedTTY.MatchString(tty) {
			return "", false, errors.New("invalid session evidence")
		}
		requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		token, err := serviceAccessToken(requestCtx, manager)
		if err != nil {
			return "", false, err
		}
		var inventory contract.ListResponse
		if err := client.Request(requestCtx, contract.ListSubject, contract.ListRequest{AccessToken: token}, &inventory); err != nil {
			return "", false, err
		}
		matched := ""
		for _, session := range inventory.Sessions {
			if session.User != username || session.TTY != tty {
				continue
			}
			if !verifiedSessionID.MatchString(session.ID) || matched != "" {
				return "", false, nil
			}
			matched = session.ID
		}
		if matched == "" {
			return "", false, nil
		}
		var result contract.ActionResponse
		if err := client.Request(requestCtx, contract.TerminateSubject,
			contract.ActionRequest{AccessToken: token, Session: matched}, &result); err != nil {
			return "", false, err
		}
		if !result.Completed {
			return "", false, errors.New("session termination was not acknowledged")
		}
		return matched, true, nil
	}, nil
}
