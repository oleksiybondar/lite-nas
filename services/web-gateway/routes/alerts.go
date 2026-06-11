package routes

import (
	"net/http"

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

	registerAlertsDomainOperations(group, basePath, controller)
}

func operationID(basePath string, suffix string) string {
	switch basePath {
	case "/alerts/system":
		return "system-alerts-" + suffix
	case "/alerts/security":
		return "security-alerts-" + suffix
	default:
		return suffix
	}
}

func registerAlertsDomainOperations(group huma.API, basePath string, controller controllers.AlertsController) {
	for _, register := range alertsOperations(controller) {
		register(group, basePath)
	}
}

func alertsOperations(controller controllers.AlertsController) []func(huma.API, string) {
	return []func(huma.API, string){
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "list", http.MethodGet, "", "List alerts", `Returns one page of alerts. Optional repeated "filters" query parameters accept JSON-encoded filter objects using the logging-manager filter contract.`), controller.List)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "count", http.MethodGet, "/count", "Count alerts", "Returns the total alert count for the configured domain."), controller.Count)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "list-active", http.MethodGet, "/active", "List active alerts", `Returns one page of active alerts. Optional repeated "filters" query parameters accept JSON-encoded filter objects using the logging-manager filter contract.`), controller.ListActive)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "count-active", http.MethodGet, "/active/count", "Count active alerts", "Returns the total active alert count for the configured domain."), controller.CountActive)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "list-unacknowledged", http.MethodGet, "/unacknowledged", "List unacknowledged alerts", `Returns one page of active unacknowledged alerts. Optional repeated "filters" query parameters accept JSON-encoded filter objects using the logging-manager filter contract.`), controller.ListUnacknowledged)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "count-unacknowledged", http.MethodGet, "/unacknowledged/count", "Count unacknowledged alerts", "Returns the total active unacknowledged alert count for the configured domain."), controller.CountUnacknowledged)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "get", http.MethodGet, "/{id}", "Get alert", "Returns one alert detail item by business record ID."), controller.Get)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "acknowledge", http.MethodPost, "/{id}/acknowledge", "Acknowledge alert", "Acknowledges one alert in the configured domain."), controller.Acknowledge)
		},
		func(group huma.API, basePath string) {
			huma.Register(group, newAlertsOperation(basePath, "mute", http.MethodPost, "/{id}/mute", "Mute alert", "Mutes one alert in the configured domain."), controller.Mute)
		},
	}
}

func newAlertsOperation(basePath string, suffix string, method string, path string, summary string, description string) huma.Operation {
	return huma.Operation{
		OperationID: operationID(basePath, suffix),
		Method:      method,
		Path:        path,
		Summary:     summary,
		Description: description,
	}
}
