// Package management defines browser-facing service and session management schemas.
package management

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	servicecontract "lite-nas/shared/contracts/servicesmanager"
	sessioncontract "lite-nas/shared/contracts/sessionmanager"
)

// ListInput carries no parameters; authentication comes from gateway middleware.
type ListInput struct{}

// ServiceActionInput identifies a systemd unit in a service command URL.
type ServiceActionInput struct {
	Unit string `path:"unit" minLength:"9" maxLength:"256" pattern:"^[A-Za-z0-9][A-Za-z0-9_.@:\\\\-]*\\.service$" doc:"Systemd system service unit name, including the .service suffix."`
}

// SessionActionInput identifies a loginctl session in a session command URL.
type SessionActionInput struct {
	Session string `path:"session" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$" doc:"Login session ID returned by the session inventory."`
}

// Service describes one systemd system unit in the browser API.
type Service struct {
	Name         string `json:"name" doc:"Systemd unit name ending in .service."`
	LoadState    string `json:"load_state" doc:"Systemd load state, such as loaded or not-found."`
	ActiveState  string `json:"active_state" doc:"High-level runtime state, such as active or inactive."`
	SubState     string `json:"sub_state" doc:"Detailed runtime state reported by systemd."`
	Description  string `json:"description" doc:"Human-readable systemd unit description."`
	EnabledState string `json:"enabled_state,omitempty" doc:"Unit file enablement state when known."`
	Preset       string `json:"preset,omitempty" doc:"Vendor preset when known."`
}

// Session describes one loginctl session in the browser API.
type Session struct {
	ID   string `json:"id" doc:"Logind session identifier used by management commands."`
	UID  string `json:"uid" doc:"Numeric user ID owning the session."`
	User string `json:"user" doc:"User name owning the session."`
}

// ServicesOutput returns the current systemd inventory.
type ServicesOutput struct{ Body ServicesBody }

// ServicesBody contains service states and common response metadata.
type ServicesBody struct {
	dto.ResponseMeta
	Data []Service `json:"data" doc:"Loaded and installed system service units."`
}

// SessionsOutput returns the current host login sessions.
type SessionsOutput struct{ Body SessionsBody }

// SessionsBody contains session identities and common response metadata.
type SessionsBody struct {
	dto.ResponseMeta
	Data []Session `json:"data" doc:"Current host login sessions."`
}

// ManagementActionOutput confirms completion of a management command.
type ManagementActionOutput struct{ Body ManagementActionBody }

// ManagementActionBody contains the command result and common response metadata.
type ManagementActionBody struct {
	dto.ResponseMeta
	Completed bool `json:"completed" doc:"True only when the backend confirmed command completion."`
}

// NewServicesBody wraps an inventory with success metadata.
func NewServicesBody(items []servicecontract.Service) ServicesBody {
	return ServicesBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: mapServices(items)}
}

// NewSessionsBody wraps an inventory with success metadata.
func NewSessionsBody(items []sessioncontract.Session) SessionsBody {
	return SessionsBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: mapSessions(items)}
}

// NewManagementActionBody confirms a completed backend command.
func NewManagementActionBody() ManagementActionBody {
	return ManagementActionBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Completed: true}
}

func mapServices(items []servicecontract.Service) []Service {
	result := make([]Service, 0, len(items))
	for _, item := range items {
		result = append(result, Service{Name: item.Name, LoadState: item.LoadState, ActiveState: item.ActiveState, SubState: item.SubState, Description: item.Description, EnabledState: item.EnabledState, Preset: item.Preset})
	}
	return result
}

func mapSessions(items []sessioncontract.Session) []Session {
	result := make([]Session, 0, len(items))
	for _, item := range items {
		result = append(result, Session{ID: item.ID, UID: item.UID, User: item.User})
	}
	return result
}
