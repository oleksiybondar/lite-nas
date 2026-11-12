package auth

// RefreshInput accepts the refresh token from the persistent browser cookie
// transport.
type RefreshInput struct {
	RefreshTokenCookie string `cookie:"lite-nas-rt" doc:"Refresh token cookie."`
	UserAgent          string `header:"User-Agent" doc:"Client user agent bound to the refresh session."`
	Ingress            string `header:"X-LiteNAS-Ingress" maxLength:"32"`
	RealIP             string `header:"X-Real-IP" maxLength:"45"`
	LocalIP            string `header:"X-LiteNAS-Local-IP" maxLength:"45"`
	Body               RefreshRequestBody
}

// RefreshRequestBody is intentionally empty for the BFF browser auth flow.
// Refresh token material is read from the HTTP-only refresh-token cookie.
type RefreshRequestBody struct{}
