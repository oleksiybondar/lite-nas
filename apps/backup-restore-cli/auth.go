package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/term"

	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

// accessToken reuses a live per-user token or authenticates without retaining
// the refresh token returned by Auth Service.
func accessToken(ctx context.Context, client messaging.Client) (string, error) {
	path, err := tokenPath()
	if err != nil {
		return "", err
	}
	return accessTokenAtPath(ctx, client, path, login)
}

// accessTokenAtPath coordinates cache validation and interactive login.
func accessTokenAtPath(ctx context.Context, client messaging.Client, path string, authenticate func(context.Context, messaging.Client) (string, error)) (string, error) {
	cached, err := cachedToken(ctx, client, path)
	if err != nil {
		return "", err
	}
	if cached != "" {
		return cached, nil
	}
	token, err := authenticate(ctx, client)
	if err != nil {
		return "", err
	}
	if err := writeToken(path, token); err != nil {
		return "", err
	}
	return token, nil
}

// cachedToken asks Auth Service whether a stored access token is still valid.
func cachedToken(ctx context.Context, client messaging.Client, path string) (string, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > 8192 {
		return "", fmt.Errorf("unsafe token cache")
	}
	cached, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(cached) == 0 {
		return "", nil
	}
	var response authcontract.ValidateAccessTokenResponse
	if err := client.Request(ctx, authcontract.ValidateAccessTokenRPCSubject, authcontract.ValidateAccessTokenRequest{AccessToken: string(cached)}, &response); err != nil {
		return "", err
	}
	if response.Valid && response.Status == authcontract.StatusAuthenticated {
		return string(cached), nil
	}
	return "", nil
}

// tokenPath confines the cache to the caller-owned systemd runtime directory.
func tokenPath() (string, error) {
	uid := os.Geteuid()
	runtimeDir := filepath.Join("/run/user", strconv.Itoa(uid))
	return tokenPathInRuntime(runtimeDir, uid)
}

// tokenPathInRuntime validates ownership and permissions before creating a cache path.
func tokenPathInRuntime(runtimeDir string, uid int) (string, error) {
	info, err := os.Lstat(runtimeDir)
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || int(stat.Uid) != uid || info.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("unsafe user runtime directory")
	}
	dir := filepath.Join(runtimeDir, "lite-nas")
	if err := os.Mkdir(dir, 0o700); err != nil && !os.IsExist(err) {
		return "", err
	}
	info, err = os.Lstat(dir)
	if err != nil {
		return "", err
	}
	stat, ok = info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || int(stat.Uid) != uid || info.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("unsafe token directory")
	}
	return filepath.Join(dir, "backup-restore-cli.token"), nil
}

// writeToken atomically publishes one access token with owner-only permissions.
func writeToken(path, token string) error {
	if token == "" || len(token) > 8192 {
		return fmt.Errorf("invalid access token")
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".backup-restore-token-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.WriteString(token); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), path)
}

// login completes Auth Service's PAM conversation and keeps only its access token.
func login(ctx context.Context, client messaging.Client) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("interactive terminal required: %w", err)
	}
	defer tty.Close()
	return loginWithPrompt(ctx, client, func(label string, secret bool) (string, error) { return prompt(tty, label, secret) })
}

// loginWithPrompt handles Auth Service's login and PAM prompts without persisting refresh material.
func loginWithPrompt(ctx context.Context, client messaging.Client, ask func(string, bool) (string, error)) (string, error) {
	username, err := ask("Username: ", false)
	if err != nil {
		return "", err
	}
	password, err := ask("Password: ", true)
	if err != nil {
		return "", err
	}
	if username == "" || len(username) > 128 || password == "" || len(password) > 4096 {
		return "", fmt.Errorf("invalid credentials")
	}
	var response authcontract.LoginResponse
	err = client.Request(ctx, authcontract.LoginRPCSubject, authcontract.LoginRequest{Username: username, Password: password}, &response)
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
			return "", fmt.Errorf("invalid authentication answer")
		}
		request := authcontract.PasswordChangeReplyRequest{ConversationID: response.Conversation.ID, Sequence: response.Conversation.Sequence, Answer: answer}
		err = client.Request(ctx, authcontract.PasswordChangeReplyRPCSubject, request, &response)
		answer = ""
		if err != nil {
			return "", err
		}
	}
	if response.Status != authcontract.StatusAuthenticated || response.AccessToken == "" {
		return "", errors.New("authentication failed")
	}
	return response.AccessToken, nil
}

// prompt reads one credential from the controlling terminal, hiding secrets.
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
			return "", fmt.Errorf("authentication input too long")
		}
		value = append(value, character[0])
	}
}
