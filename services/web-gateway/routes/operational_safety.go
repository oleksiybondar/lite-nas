package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountOperationalSafetyRouter exposes security-role operational safety RPCs.
func mountOperationalSafetyRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/system/automations")
	group.UseSimpleModifier(func(operation *huma.Operation) { operation.Tags = []string{"operational-safety-automation"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireSecurity(api))
	c := controllers.OperationalSafety
	huma.Register(group, huma.Operation{OperationID: "operational-safety-rules", Method: http.MethodGet, Path: "/rules", Summary: "List operational safety rules"}, c.Rules)
	huma.Register(group, huma.Operation{OperationID: "operational-safety-tasks", Method: http.MethodGet, Path: "/tasks", Summary: "List operational safety tasks"}, c.Tasks)
	huma.Register(group, huma.Operation{OperationID: "operational-safety-rule-state", Method: http.MethodPut, Path: "/rules/{rule_id}/enabled", Summary: "Temporarily enable or disable a rule"}, c.SetEnabled)
	huma.Register(group, huma.Operation{OperationID: "operational-safety-force-set", Method: http.MethodPost, Path: "/tasks/{task_id}/force-set", Summary: "Apply a configured level to a task"}, c.ForceSet)
	huma.Register(group, huma.Operation{OperationID: "operational-safety-force-stop", Method: http.MethodPost, Path: "/tasks/{task_id}/force-stop", Summary: "Start staged recovery of a task"}, c.ForceStop)
}
