//go:build !pam

package pamauth

type authenticator struct {
	serviceName string
}

// NewAuthenticator constructs a stub authenticator for builds that do not
// include PAM support.
func NewAuthenticator(serviceName string) (Authenticator, error) {
	if err := validateServiceName(serviceName); err != nil {
		return nil, err
	}

	return authenticator{
		serviceName: serviceName,
	}, nil
}

// Authenticate reports that PAM support is unavailable in the current build.
func (a authenticator) Authenticate(request AuthenticateRequest) (Result, error) {
	return newServiceUnavailableResult(request.Username, errPAMUnavailable.Error()), errPAMUnavailable
}

// ChangePassword reports that PAM support is unavailable in the current build.
func (a authenticator) ChangePassword(request PasswordChangeRequest) (Result, error) {
	return newServiceUnavailableResult(request.Username, errPAMUnavailable.Error()), errPAMUnavailable
}
