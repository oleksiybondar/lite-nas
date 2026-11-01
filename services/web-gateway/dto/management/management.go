// Package management defines browser-facing host and device management schemas.
package management

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	servicecontract "lite-nas/shared/contracts/servicesmanager"
	sessioncontract "lite-nas/shared/contracts/sessionmanager"
	usbcontract "lite-nas/shared/contracts/usbmanager"
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

// USBDeviceActionInput identifies a USBGuard device and optional persistent policy change.
type USBDeviceActionInput struct {
	DeviceID  string `path:"device_id" minLength:"1" maxLength:"10" pattern:"^[0-9]{1,10}$" doc:"Numeric USBGuard device ID returned by the inventory."`
	Permanent bool   `json:"permanent,omitempty" doc:"Persist this authorization decision in USBGuard policy."`
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
	ID         string `json:"id" doc:"Logind session identifier used by management commands."`
	UID        string `json:"uid" doc:"Numeric user ID owning the session."`
	User       string `json:"user" doc:"User name owning the session."`
	Seat       string `json:"seat" doc:"Seat associated with the session, when present."`
	Leader     string `json:"leader" doc:"Process ID of the session leader."`
	Class      string `json:"class" doc:"Logind session class."`
	TTY        string `json:"tty" doc:"Terminal associated with the session, when present."`
	Type       string `json:"type" doc:"Logind session type, such as wayland, x11, or tty."`
	Service    string `json:"service" doc:"PAM service that opened the session."`
	Remote     bool   `json:"remote" doc:"Whether logind marks the session as remote."`
	RemoteHost string `json:"remote_host" doc:"Remote host reported by logind, when present."`
	Display    string `json:"display" doc:"Display associated with the session, when present."`
	Desktop    string `json:"desktop" doc:"Desktop environment associated with the session, when present."`
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

// USBDevicesOutput returns USBGuard's current device inventory.
type USBDevicesOutput struct{ Body USBDevicesBody }

// USBDevicesBody contains the device inventory and common response metadata.
type USBDevicesBody struct {
	dto.ResponseMeta
	Data string `json:"data" doc:"Current USBGuard device inventory as returned by usbguard list-devices."`
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

// NewUSBDevicesBody wraps USBGuard's inventory with success metadata.
func NewUSBDevicesBody(items usbcontract.ListResponse) USBDevicesBody {
	return USBDevicesBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: items.Devices}
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
		result = append(result, Session{ID: item.ID, UID: item.UID, User: item.User, Seat: item.Seat, Leader: item.Leader, Class: item.Class, TTY: item.TTY, Type: item.Type, Service: item.Service, Remote: item.Remote, RemoteHost: item.RemoteHost, Display: item.Display, Desktop: item.Desktop})
	}
	return result
}
