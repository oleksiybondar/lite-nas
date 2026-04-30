package auth

// LoginInput documents the stub login request body.
type LoginInput struct {
	UserAgent string `header:"User-Agent" doc:"Client user agent bound to the refresh session."`
	Body      LoginRequestBody
}

// LoginRequestBody describes the draft login payload validated by the browser
// boundary before controller execution.
type LoginRequestBody struct {
	Login    string `json:"login" required:"true" pattern:"^(?:[A-Za-z0-9._-]+|[^@\\s]+@[^@\\s]+\\.[^@\\s]+)$" doc:"Local login or email address."`
	Password string `json:"password" required:"true" minLength:"4" doc:"Password input for the login flow."`
}
