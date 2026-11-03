package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountSecurityAutomationRouter exposes authenticated operator RPCs over HTTP.
func mountSecurityAutomationRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/security/automations")
	group.UseSimpleModifier(func(operation *huma.Operation) { operation.Tags = []string{"security-automation"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireSecurity(api))
	c := controllers.SecurityAutomation
	huma.Register(group, huma.Operation{OperationID: "security-automation-rules", Method: http.MethodGet, Path: "/rules", Summary: "List security automation rules"}, c.Rules)
	huma.Register(group, huma.Operation{OperationID: "security-automation-active", Method: http.MethodGet, Path: "/active", Summary: "List active security automations"}, c.Active)
	huma.Register(group, huma.Operation{OperationID: "security-automation-rollback", Method: http.MethodPost, Path: "/active/{incident_id}/rollback", Summary: "Restore an automation effect"}, c.Rollback)
	huma.Register(group, huma.Operation{OperationID: "security-automation-force-normalize", Method: http.MethodPost, Path: "/active/{incident_id}/force-normalize", Summary: "Restore and normalize a verified incident"}, c.ForceNormalize)
	huma.Register(group, huma.Operation{OperationID: "security-automation-force-mitigate", Method: http.MethodPost, Path: "/rules/{rule_id}/mitigate", Summary: "Manually trigger an automation rule"}, c.ForceMitigate)
	huma.Register(group, huma.Operation{OperationID: "security-automation-cooldown-reset", Method: http.MethodPost, Path: "/rules/{rule_id}/cooldown/reset", Summary: "Reset one keyed detection state"}, c.ResetCooldown)
	huma.Register(group, huma.Operation{OperationID: "security-automation-rule-state", Method: http.MethodPut, Path: "/rules/{rule_id}/enabled", Summary: "Temporarily enable or disable a rule"}, c.SetEnabled)
}
