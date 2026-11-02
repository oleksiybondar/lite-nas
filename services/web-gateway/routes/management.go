package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountManagementRouters adds documented host management routes to the shared API.
func mountManagementRouters(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	services := huma.NewGroup(api, "/services")
	services.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"services-management"} })
	services.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	services.UseMiddleware(middlewares.RequireAdministrator(api))
	service := controllers.ServicesManagement
	huma.Register(services, managementOperation("services-list", http.MethodGet, "", "List system services", "Returns loaded and installed systemd system units with their load, active, enablement, and preset states. Requires an administrator-equivalent access token."), service.List)
	huma.Register(services, managementOperation("services-start", http.MethodPost, "/{unit}/start", "Start system service", "Starts the specified systemd system service unit. The unit must have a .service suffix. Requires an administrator-equivalent access token."), service.Start)
	huma.Register(services, managementOperation("services-stop", http.MethodPost, "/{unit}/stop", "Stop system service", "Stops the specified systemd system service unit. Requires an administrator-equivalent access token."), service.Stop)
	huma.Register(services, managementOperation("services-restart", http.MethodPost, "/{unit}/restart", "Restart system service", "Restarts the specified systemd system service unit. Requires an administrator-equivalent access token."), service.Restart)
	huma.Register(services, managementOperation("services-enable", http.MethodPost, "/{unit}/enable", "Enable system service", "Enables the specified systemd system service unit at boot. Requires an administrator-equivalent access token."), service.Enable)
	huma.Register(services, managementOperation("services-disable", http.MethodPost, "/{unit}/disable", "Disable system service", "Disables the specified systemd system service unit at boot. Requires an administrator-equivalent access token."), service.Disable)

	sessions := huma.NewGroup(api, "/sessions")
	sessions.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"sessions-management"} })
	sessions.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	sessions.UseMiddleware(middlewares.RequireSecurity(api))
	session := controllers.SessionsManagement
	huma.Register(sessions, managementOperation("sessions-list", http.MethodGet, "", "List host sessions", "Returns current loginctl sessions with session ID, UID, and user name. Requires a security-equivalent access token."), session.List)
	huma.Register(sessions, managementOperation("sessions-terminate", http.MethodPost, "/{session}/terminate", "Terminate host session", "Terminates the specified login session, including sessions owned by other users. Requires a security-equivalent access token."), session.Terminate)
	huma.Register(sessions, managementOperation("sessions-lock", http.MethodPost, "/{session}/lock", "Lock host session", "Locks the specified login session. Requires a security-equivalent access token."), session.Lock)
	huma.Register(sessions, managementOperation("sessions-unlock", http.MethodPost, "/{session}/unlock", "Unlock host session", "Unlocks the specified login session. Requires a security-equivalent access token."), session.Unlock)

	authSessions := huma.NewGroup(api, "/auth/sessions")
	authSessions.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"auth-sessions-management"} })
	authSessions.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	authSessions.UseMiddleware(middlewares.RequireAdministrator(api))
	authSession := controllers.AuthSessions
	huma.Register(authSessions, managementOperation("auth-sessions-list", http.MethodGet, "", "List LiteNAS Auth sessions", "Returns active authentication sessions. Requires an administrator access token."), authSession.List)
	huma.Register(authSessions, managementOperation("auth-sessions-revoke", http.MethodPost, "/{session_id}/revoke", "Revoke LiteNAS Auth session", "Revokes the selected authentication session. Requires an administrator access token."), authSession.Revoke)
}

func managementOperation(id, method, path, summary, description string) huma.Operation {
	return huma.Operation{OperationID: id, Method: method, Path: path, Summary: summary, Description: description, DefaultStatus: http.StatusOK}
}
