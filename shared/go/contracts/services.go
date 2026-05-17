package contracts

// Service and app identity constants shared across LiteNAS code and deployment
// configuration.
const (
	ServiceAuth                   = "auth-service"
	ServiceSystemMetrics          = "system-metrics"
	ServiceResourcesMonitor       = "resources-monitor"
	ServiceSystemLoggingManager   = "system-logging-manager"
	ServiceSecurityLoggingManager = "security-logging-manager"
	ServiceWebGateway             = "web-gateway"

	AppSystemMetricsCLI        = "system-metrics-cli"
	AppSystemLoggingManagerCLI = "system-logging-manager-cli"
	AppSecurityLoggingMgrCLI   = "security-logging-manager-cli"
	AppAdminPanel              = "admin-panel"
)
