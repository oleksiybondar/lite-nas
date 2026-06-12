package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountNetworkMetricsRouter registers the documented network metrics routes on
// the shared API.
func mountNetworkMetricsRouter(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	group := huma.NewGroup(api, "/network-metrics")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"network-metrics"}
	})
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))

	controller := controllerModule.NetworkMetrics
	huma.Get(group, "/snapshot", controller.GetSnapshot)
	huma.Get(group, "/history", controller.GetHistory)
}
