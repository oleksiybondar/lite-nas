package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountLogAuditRouter exposes the security operator's Log Analysis rule catalog.
func mountLogAuditRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/security/log-audit")
	group.UseSimpleModifier(func(operation *huma.Operation) { operation.Tags = []string{"log-audit"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireSecurity(api))
	huma.Register(group, huma.Operation{OperationID: "log-audit-rules", Method: http.MethodGet, Path: "/rules", Summary: "List Log Analysis rules"}, controllers.LogAudit.Rules)
}
