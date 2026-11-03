package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountProcessMetricsRouter registers the documented process metrics routes on
// the shared API.
func mountProcessMetricsRouter(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	group := huma.NewGroup(api, "/process-metrics")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = []string{"process-metrics"}
	})
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))

	controller := controllerModule.ProcessMetrics
	huma.Get(group, "/snapshot", controller.GetSnapshot)
}
