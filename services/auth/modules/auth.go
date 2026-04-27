package modules

import (
	"lite-nas/services/auth/pamauth"
)

// Auth groups PAM-backed authentication dependencies.
type Auth struct {
	ServiceName   string
	Authenticator pamauth.Authenticator
}

// NewAuthModule constructs the PAM-backed authenticator for the auth-service.
func NewAuthModule(serviceName string) (Auth, error) {
	authenticator, err := pamauth.NewAuthenticator(serviceName)
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		ServiceName:   serviceName,
		Authenticator: authenticator,
	}, nil
}
