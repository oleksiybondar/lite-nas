// Package servicesmanager defines authenticated system service control RPCs.
package servicesmanager

const (
	ListSubject    = "services-manager.rpc.list"
	StartSubject   = "services-manager.rpc.start"
	StopSubject    = "services-manager.rpc.stop"
	RestartSubject = "services-manager.rpc.restart"
	EnableSubject  = "services-manager.rpc.enable"
	DisableSubject = "services-manager.rpc.disable"
)

// ListRequest authenticates a system service inventory request.
type ListRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
}

// ActionRequest identifies one system service unit to manage.
type ActionRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Unit        string `json:"unit" validate:"required,max=256"`
}

// Service describes one loaded or known systemd service unit.
type Service struct {
	Name         string `json:"name"`
	LoadState    string `json:"load_state"`
	ActiveState  string `json:"active_state"`
	SubState     string `json:"sub_state"`
	Description  string `json:"description"`
	EnabledState string `json:"enabled_state,omitempty"`
	Preset       string `json:"preset,omitempty"`
}

// ListResponse contains the current systemd service inventory.
type ListResponse struct {
	Services []Service `json:"services"`
}

// ActionResponse confirms that systemctl completed an action successfully.
type ActionResponse struct {
	Completed bool `json:"completed"`
}
