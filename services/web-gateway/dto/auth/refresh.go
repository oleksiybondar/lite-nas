package auth

// RefreshInput accepts the refresh token from either the payload or the
// persistent cookie transport.
type RefreshInput struct {
	RefreshTokenCookie string `cookie:"lite-nas-rt" doc:"Refresh token cookie."`
	UserAgent          string `header:"User-Agent" doc:"Client user agent bound to the refresh session."`
	Body               RefreshRequestBody
}

// RefreshRequestBody defines the public refresh-token payload transport.
type RefreshRequestBody struct {
	RefreshToken string `json:"refresh_token,omitempty" pattern:"^RT-[A-Za-z0-9-]+$" doc:"Explicit refresh token payload for non-cookie clients."`
}
