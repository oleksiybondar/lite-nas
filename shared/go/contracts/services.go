package contracts

// Service and app identity constants shared across LiteNAS code and deployment
// configuration.
const (
	ServiceAuth                   = "auth-service"
	ServiceNetworkMetrics         = "network-metrics"
	ServiceSystemMetrics          = "system-metrics"
	ServiceZFSMetrics             = "zfs-metrics"
	ServiceResourcesMonitor       = "resources-monitor"
	ServiceSystemEmailNotifier    = "system-email-notifier"
	ServiceSecurityEmailNotifier  = "security-email-notifier"
	ServiceSystemLoggingManager   = "system-logging-manager"
	ServiceSecurityLoggingManager = "security-logging-manager"
	ServiceWebGateway             = "web-gateway"
	ServiceRBAC                   = "rbac-service"

	AppSystemMetricsCLI        = "system-metrics-cli"
	AppNetworkMetricsCLI       = "network-metrics-cli"
	AppZFSMetricsCLI           = "zfs-metrics-cli"
	AppSystemLoggingManagerCLI = "system-logging-manager-cli"
	AppSecurityLoggingMgrCLI   = "security-logging-manager-cli"
	AppAdminPanel              = "admin-panel"
)
