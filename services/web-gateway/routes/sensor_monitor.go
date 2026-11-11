package routes

import (
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountSensorMonitorRouter registers authenticated sensor snapshot and history routes.
func mountSensorMonitorRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/sensor-monitor")
	group.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"sensor-monitor"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	huma.Get(group, "/snapshot", controllers.SensorMonitor.GetSnapshot)
	huma.Get(group, "/history", controllers.SensorMonitor.GetHistory)
}
