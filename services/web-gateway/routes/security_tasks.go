package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountSecurityTasksRouter registers the security-role task endpoints.
func mountSecurityTasksRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/security/tasks")
	group.UseSimpleModifier(func(operation *huma.Operation) { operation.Tags = []string{"security-tasks"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireSecurity(api))
	c := controllers.SecurityTasks
	huma.Register(group, huma.Operation{OperationID: "antivirus-scan-start", Method: http.MethodPost, Path: "/antivirus/scans", Summary: "Start an antivirus scan", Description: "Scans a folder under a configured eligible mount. Returns a task ID for status polling. Requires a security-equivalent access token."}, c.StartAntivirusScan)
	huma.Register(group, huma.Operation{OperationID: "antivirus-info", Method: http.MethodGet, Path: "/antivirus/info", Summary: "Read antivirus signature status", Description: "Returns a compact ClamAV engine and signature database status."}, c.AntivirusInfo)
	huma.Register(group, huma.Operation{OperationID: "antivirus-signatures-update", Method: http.MethodPost, Path: "/antivirus/signatures/update", Summary: "Update antivirus signatures", Description: "Starts a background FreshClam signature database update and returns its state."}, c.UpdateAntivirusSignatures)
	huma.Register(group, huma.Operation{OperationID: "antivirus-scan-status", Method: http.MethodGet, Path: "/antivirus/tasks/{task_id}", Summary: "Read antivirus task status", Description: "Returns the current state and bounded command output for an antivirus scan task."}, c.AntivirusTaskStatus)
	huma.Register(group, huma.Operation{OperationID: "antivirus-scan-retry", Method: http.MethodPost, Path: "/antivirus/tasks/{task_id}/retry", Summary: "Retry an antivirus scan", Description: "Starts a new attempt using the folder recorded by a completed scan task."}, c.RetryAntivirusTask)
	huma.Register(group, huma.Operation{OperationID: "integrity-check-start", Method: http.MethodPost, Path: "/integrity/checks", Summary: "Start an AIDE integrity check", Description: "Runs an on-demand AIDE check and returns a task ID for status polling."}, c.StartIntegrityCheck)
	huma.Register(group, huma.Operation{OperationID: "integrity-info", Method: http.MethodGet, Path: "/integrity/info", Summary: "Read AIDE baseline status", Description: "Returns whether an active AIDE baseline is available and its metadata."}, c.IntegrityInfo)
	huma.Register(group, huma.Operation{OperationID: "integrity-baseline-start", Method: http.MethodPost, Path: "/integrity/baselines", Summary: "Generate an AIDE baseline", Description: "Generates and activates a new AIDE baseline as an on-demand task."}, c.StartIntegrityBaseline)
	huma.Register(group, huma.Operation{OperationID: "integrity-check-status", Method: http.MethodGet, Path: "/integrity/tasks/{task_id}", Summary: "Read integrity task status", Description: "Returns the current state and bounded command output for an AIDE check task."}, c.IntegrityTaskStatus)
	huma.Register(group, huma.Operation{OperationID: "integrity-check-retry", Method: http.MethodPost, Path: "/integrity/tasks/{task_id}/retry", Summary: "Retry an AIDE integrity check", Description: "Starts a new attempt using the completed task's saved check request."}, c.RetryIntegrityTask)
}
