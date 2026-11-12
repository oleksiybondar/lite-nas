// Package cliauth implements interactive user login for LiteNAS command line clients.
package cliauth

import (
	"context"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"

	contract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

// Prompt asks one terminal question, hiding secret answers from echo.
type Prompt func(label string, secret bool) (string, error)

// LoginInteractive obtains a user access token without retaining the refresh token.
func LoginInteractive(ctx context.Context, client messaging.Client) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("interactive terminal required: %w", err)
	}
	defer tty.Close()
	return Login(ctx, client, func(label string, secret bool) (string, error) { return prompt(tty, label, secret) })
}

// Login completes Auth Service's login and any PAM password conversation.
func Login(ctx context.Context, client messaging.Client, ask Prompt) (string, error) {
	username, err := ask("Username: ", false)
	if err != nil {
		return "", err
	}
	password, err := ask("Password: ", true)
	if err != nil {
		return "", err
	}
	if username == "" || len(username) > 128 || password == "" || len(password) > 4096 {
		return "", errors.New("invalid credentials")
	}
	var response contract.LoginResponse
	err = client.Request(ctx, contract.LoginRPCSubject, contract.LoginRequest{Username: username, Password: password}, &response)
	password = ""
	if err != nil {
		return "", err
	}
	for response.Conversation != nil {
		answer, err := ask(response.Conversation.Prompt+": ", response.Conversation.Secret)
		if err != nil {
			return "", err
		}
		if answer == "" || len(answer) > 4096 {
			return "", errors.New("invalid authentication answer")
		}
		request := contract.PasswordChangeReplyRequest{ConversationID: response.Conversation.ID, Sequence: response.Conversation.Sequence, Answer: answer}
		err = client.Request(ctx, contract.PasswordChangeReplyRPCSubject, request, &response)
		answer = ""
		if err != nil {
			return "", err
		}
	}
	if response.Status != contract.StatusAuthenticated || response.AccessToken == "" {
		return "", errors.New("authentication failed")
	}
	return response.AccessToken, nil
}

func prompt(tty *os.File, label string, secret bool) (string, error) {
	if _, err := fmt.Fprint(tty, label); err != nil {
		return "", err
	}
	if secret {
		value, err := term.ReadPassword(int(tty.Fd()))
		_, _ = fmt.Fprintln(tty)
		return string(value), err
	}
	value := make([]byte, 0, 128)
	var character [1]byte
	for {
		_, err := tty.Read(character[:])
		if err != nil {
			return "", err
		}
		if character[0] == '\n' {
			return string(value), nil
		}
		if len(value) >= 4096 {
			return "", errors.New("authentication input too long")
		}
		value = append(value, character[0])
	}
}
