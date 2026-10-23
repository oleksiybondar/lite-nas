package browserauth

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"lite-nas/shared/httpcookie"
)

const (
	// AccessTokenCookieName is the browser access cookie shared by LiteNAS apps.
	AccessTokenCookieName = "lite-nas-at"
	// RefreshTokenCookieName is the browser refresh cookie shared by LiteNAS apps.
	RefreshTokenCookieName = "lite-nas-rt"
)

// ValidateCookieDomain accepts a DNS parent domain or an empty host-only scope.
func ValidateCookieDomain(domain string) error {
	if domain == "" {
		return nil
	}
	if strings.TrimSpace(domain) != domain || len(domain) > 253 || strings.ContainsAny(domain, ";/\\\r\n\t ") || net.ParseIP(domain) != nil {
		return errors.New("invalid browser cookie domain")
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return errors.New("browser cookie domain must be a parent DNS domain")
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return errors.New("invalid browser cookie domain label")
		}
		for _, character := range label {
			if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '-') {
				return errors.New("invalid browser cookie domain character")
			}
		}
	}
	return nil
}

// AccessCookie returns the secure browser cookie for a verified session.
func AccessCookie(session Session, domain string) http.Cookie {
	return http.Cookie{Name: AccessTokenCookieName, Value: session.AccessToken, Domain: domain, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, Expires: session.AccessExpires}
}

// RefreshCookie returns the secure browser refresh cookie for a session.
func RefreshCookie(session Session, domain string) http.Cookie {
	return http.Cookie{Name: RefreshTokenCookieName, Value: session.RefreshToken, Domain: domain, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, Expires: session.RefreshExpiry}
}

// ClearAccessCookie expires the access cookie in the same browser scope.
func ClearAccessCookie(now time.Time, domain string) http.Cookie {
	cookie := httpcookie.Expired(AccessTokenCookieName, now)
	cookie.Domain = domain
	return cookie
}

// ClearRefreshCookie expires the refresh cookie in the same browser scope.
func ClearRefreshCookie(now time.Time, domain string) http.Cookie {
	cookie := httpcookie.Expired(RefreshTokenCookieName, now)
	cookie.Domain = domain
	return cookie
}
