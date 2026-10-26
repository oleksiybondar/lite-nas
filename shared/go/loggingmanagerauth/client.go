// Package loggingmanagerauth adds service access tokens to logging-manager messages.
package loggingmanagerauth

import (
	"context"
	"time"

	"lite-nas/shared/messaging"
)

// TokenManager supplies and renews the service token used for outbound messages.
type TokenManager interface {
	Token() (string, time.Time, error)
	Refresh(context.Context) error
	Login(context.Context) error
}

// Injector returns a payload with the access token applied where its contract supports one.
type Injector func(payload any, accessToken string) any

// Client decorates a messaging client with service-token injection.
type Client struct {
	messaging.Client
	tokenManager TokenManager
	inject       Injector
}

// NewClient wraps outbound messages while preserving the client's lifecycle methods.
func NewClient(client messaging.Client, tokenManager TokenManager, inject Injector) *Client {
	return &Client{Client: client, tokenManager: tokenManager, inject: inject}
}

// Publish attaches the current token to a supported outbound payload.
func (client *Client) Publish(ctx context.Context, subject string, payload any) error {
	accessToken, err := client.CurrentToken(ctx)
	if err != nil {
		return err
	}
	return client.Client.Publish(ctx, subject, client.inject(payload, accessToken))
}

// Request attaches the current token to a supported outbound request.
func (client *Client) Request(ctx context.Context, subject string, request any, response any) error {
	accessToken, err := client.CurrentToken(ctx)
	if err != nil {
		return err
	}
	return client.Client.Request(ctx, subject, client.inject(request, accessToken), response)
}

// CurrentToken obtains a token, refreshing or logging in when necessary.
func (client *Client) CurrentToken(ctx context.Context) (string, error) {
	accessToken, _, err := client.tokenManager.Token()
	if err == nil {
		return accessToken, nil
	}
	if refreshErr := client.tokenManager.Refresh(ctx); refreshErr == nil {
		accessToken, _, err = client.tokenManager.Token()
		if err == nil {
			return accessToken, nil
		}
	}
	if loginErr := client.tokenManager.Login(ctx); loginErr != nil {
		return "", loginErr
	}
	accessToken, _, err = client.tokenManager.Token()
	return accessToken, err
}
