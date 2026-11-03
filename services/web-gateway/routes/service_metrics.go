package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountServiceMetricsRouter registers the documented service metrics routes on
// the shared API.
func mountServiceMetricsRouter(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	group := huma.NewGroup(api, "/service-metrics")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"service-metrics"}
	})
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))

	controller := controllerModule.ServiceMetrics
	huma.Get(group, "/snapshot", controller.GetSnapshot)
	huma.Get(group, "/history", controller.GetHistory)
}
