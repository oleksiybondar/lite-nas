package routes

import (
	"lite-nas/services/web-gateway/controllers"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"

	"github.com/danielgtaylor/huma/v2"
)

// mountAlertsRouters registers the documented alert route slices on the shared API.
func mountAlertsRouters(
	api huma.API,
	controllerModule modules.Controllers,
	authentication middlewares.AuthenticationOptions,
) {
	mountAlertsDomainRouter(
		api,
		"/alerts/system",
		[]string{"system-alerts"},
		controllerModule.SystemAlerts,
		authentication,
		middlewares.RequireOperator(api),
	)
	mountAlertsDomainRouter(
		api,
		"/alerts/security",
		[]string{"security-alerts"},
		controllerModule.SecurityAlerts,
		authentication,
		middlewares.RequireSecurity(api),
	)
}

func mountAlertsDomainRouter(
	api huma.API,
	basePath string,
	tags []string,
	controller controllers.AlertsController,
	authentication middlewares.AuthenticationOptions,
	authorize func(huma.Context, func(huma.Context)),
) {
	group := huma.NewGroup(api, basePath)
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.Tags = tags
	})
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(authorize)

	huma.Get(group, "", controller.List)
	huma.Get(group, "/count", controller.Count)
	huma.Get(group, "/unacknowledged", controller.ListUnacknowledged)
	huma.Get(group, "/unacknowledged/count", controller.CountUnacknowledged)
	huma.Get(group, "/active", controller.ListUnacknowledged)
	huma.Get(group, "/active/count", controller.CountUnacknowledged)
	huma.Get(group, "/{id}", controller.Get)
	huma.Post(group, "/{id}/acknowledge", controller.Acknowledge)
	huma.Post(group, "/{id}/mute", controller.Mute)
}
