package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountSystemMetricsRouter registers the documented system metrics routes on
// the shared API.
func mountSystemMetricsRouter(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	group := huma.NewGroup(api, "/system-metrics")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"system-metrics"}
	})
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))

	controller := controllerModule.SystemMetrics
	huma.Get(group, "/snapshot", controller.GetSnapshot)
	huma.Get(group, "/history", controller.GetHistory)
}
