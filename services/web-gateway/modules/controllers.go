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
	ServicesManagement controllers.ServicesManagementController
	SessionsManagement controllers.SessionsManagementController
	AuthSessions       controllers.AuthSessionsController
	Journal            controllers.JournalController
	USBManager         controllers.USBManagerController
	Auth               controllers.AuthController
	Static             controllers.StaticController
	SystemAlerts       controllers.AlertsController
	SecurityAlerts     controllers.AlertsController
	SystemMetrics      controllers.SystemMetricsController
	ProcessMetrics     controllers.ProcessMetricsController
	ServiceMetrics     controllers.ServiceMetricsController
	DiskMetrics        controllers.DiskMetricsController
	NetworkMetrics     controllers.NetworkMetricsController
	ZFSMetrics         controllers.ZFSMetricsController
	BackupRestore      controllers.BackupRestoreController
	SecurityTasks      controllers.SecurityTasksController
	SecurityAutomation controllers.SecurityAutomationController
	LogAudit           controllers.LogAuditController
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
	cookieDomain ...string,
) Controllers {
	domain := ""
	if len(cookieDomain) > 0 {
		domain = cookieDomain[0]
	}
	return Controllers{
		ServicesManagement: controllers.NewServicesManagementController(services.ServicesManagement),
		SessionsManagement: controllers.NewSessionsManagementController(services.SessionsManagement),
		AuthSessions:       controllers.NewAuthSessionsController(services.AuthSessions),
		Journal:            controllers.NewJournalController(services.Journal),
		USBManager:         controllers.NewUSBManagerController(services.USBManager),
		Auth:               controllers.NewAuthController(services.Auth, domain),
		Static:             controllers.NewStaticController(staticFiles, log),
		SystemAlerts:       controllers.NewSystemAlertsController(services.SystemAlerts),
		SecurityAlerts:     controllers.NewSecurityAlertsController(services.SecurityAlerts),
		SystemMetrics:      controllers.NewSystemMetricsController(services.SystemMetrics),
		ProcessMetrics:     controllers.NewProcessMetricsController(services.ProcessMetrics),
		ServiceMetrics:     controllers.NewServiceMetricsController(services.ServiceMetrics),
		DiskMetrics:        controllers.NewDiskMetricsController(services.DiskMetrics),
		NetworkMetrics:     controllers.NewNetworkMetricsController(services.NetworkMetrics),
		ZFSMetrics:         controllers.NewZFSMetricsController(services.ZFSMetrics),
		BackupRestore:      controllers.NewBackupRestoreController(services.BackupRestore),
		SecurityTasks:      controllers.NewSecurityTasksController(services.SecurityTasks),
		SecurityAutomation: controllers.NewSecurityAutomationController(services.SecurityAutomation),
		LogAudit:           controllers.NewLogAuditController(services.LogAudit),
	}
}
