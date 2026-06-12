package modules

import (
	"lite-nas/services/web-gateway/controllers"
	sharedlogger "lite-nas/shared/logger"
)

// Controllers groups the gateway HTTP controllers.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only by the runtime after construction.
type Controllers struct {
	Auth           controllers.AuthController
	Static         controllers.StaticController
	SystemAlerts   controllers.AlertsController
	SecurityAlerts controllers.AlertsController
	SystemMetrics  controllers.SystemMetricsController
	NetworkMetrics controllers.NetworkMetricsController
	ZFSMetrics     controllers.ZFSMetricsController
}

// NewControllersModule assembles the HTTP controllers used by the route layer.
//
// Parameters:
//   - staticFiles: packaged frontend assets exposed by the static controller
//   - log: application logger used for static asset load failures
//   - services: service-layer dependencies consumed by the controllers
func NewControllersModule(
	staticFiles controllers.StaticFiles,
	log sharedlogger.Logger,
	services Services,
) Controllers {
	return Controllers{
		Auth:           controllers.NewAuthController(services.Auth),
		Static:         controllers.NewStaticController(staticFiles, log),
		SystemAlerts:   controllers.NewSystemAlertsController(services.SystemAlerts),
		SecurityAlerts: controllers.NewSecurityAlertsController(services.SecurityAlerts),
		SystemMetrics:  controllers.NewSystemMetricsController(services.SystemMetrics),
		NetworkMetrics: controllers.NewNetworkMetricsController(services.NetworkMetrics),
		ZFSMetrics:     controllers.NewZFSMetricsController(services.ZFSMetrics),
	}
}
