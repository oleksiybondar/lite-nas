package authtoken

import (
	"context"
	"errors"

	authcontract "lite-nas/shared/contracts/auth"
)

// LiveRequester sends token validation requests to the Auth service.
type LiveRequester interface {
	Request(context.Context, string, any, any) error
}

// VerifyLive checks current Auth state and returns locally verified role claims.
func VerifyLive(ctx context.Context, client LiveRequester, verifier Verifier, token string) (AccessClaims, error) {
	if client == nil {
		return AccessClaims{}, errors.New("auth requester unavailable")
	}
	claims, err := verifier.Verify(token)
	if err != nil {
		return AccessClaims{}, err
	}
	var response authcontract.ValidateAccessTokenResponse
	if err := client.Request(ctx, authcontract.ValidateAccessTokenRPCSubject, authcontract.ValidateAccessTokenRequest{AccessToken: token}, &response); err != nil {
		return AccessClaims{}, err
	}
	if !response.Valid || response.Status != authcontract.StatusAuthenticated || response.Username != claims.Login {
		return AccessClaims{}, errors.New("access token is not active")
	}
	return claims, nil
}
