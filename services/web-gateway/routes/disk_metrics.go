package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountDiskMetricsRouter registers the documented disk metrics routes on the
// shared API.
func mountDiskMetricsRouter(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	group := huma.NewGroup(api, "/disk-metrics")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"disk-metrics"}
	})
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))

	controller := controllerModule.DiskMetrics
	huma.Get(group, "/snapshot", controller.GetSnapshot)
	huma.Get(group, "/history", controller.GetHistory)
}
