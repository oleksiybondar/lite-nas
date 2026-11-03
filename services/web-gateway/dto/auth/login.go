package auth

// LoginInput documents the stub login request body.
type LoginInput struct {
	UserAgent  string `header:"User-Agent" doc:"Client user agent bound to the refresh session."`
	Ingress    string `header:"X-LiteNAS-Ingress" maxLength:"32" doc:"Nginx-set ingress class."`
	RealIP     string `header:"X-Real-IP" maxLength:"45" doc:"Nginx-observed peer address."`
	LocalIP    string `header:"X-LiteNAS-Local-IP" maxLength:"45" doc:"Nginx local listener address."`
	RedirectTo string `query:"redirect_to" maxLength:"2048" doc:"Optional app-local browser path to restore after login."`
	Body       LoginRequestBody
}

// LoginRequestBody describes the draft login payload validated by the browser
// boundary before controller execution.
type LoginRequestBody struct {
	Login    string `json:"login" required:"true" pattern:"^(?:[A-Za-z0-9._-]+|[^@\\s]+@[^@\\s]+\\.[^@\\s]+)$" doc:"Local login or email address."`
	Password string `json:"password" required:"true" minLength:"1" maxLength:"4096" doc:"Password input for the login flow."`
}
