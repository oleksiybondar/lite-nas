package litenasintegrity

// RegenerateBaselineRequest requests a privileged AIDE baseline regeneration.
type RegenerateBaselineRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// RegenerateBaselineResponse reports whether baseline regeneration completed.
type RegenerateBaselineResponse struct {
	OK bool `json:"ok"`
}
