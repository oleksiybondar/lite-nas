package modules

import (
	"lite-nas/services/web-gateway/services"
	"lite-nas/shared/messaging"
)

// Services groups the gateway service-layer dependencies.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only by the runtime after construction.
type Services struct {
	Auth          services.AuthService
	SystemMetrics services.SystemMetricsService
}

// NewServicesModule assembles the service-layer dependencies used by the
// gateway runtime.
//
// Parameters:
//   - client: messaging client used by services that call backend modules
func NewServicesModule(client messaging.Client) Services {
	return Services{
		Auth:          services.NewAuthService(),
		SystemMetrics: services.NewSystemMetricsService(client),
	}
}
