package modules

import (
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/servicetoken"
	sharedworkers "lite-nas/shared/workers"
)

const defaultAuthRefreshTickBuffer = 1

// CoreClientAuthInfra extends CoreInfra with auth token runtime dependencies.
type CoreClientAuthInfra struct {
	CoreInfra
	AuthTokenManager *servicetoken.Manager
	AuthRefreshTimer sharedworkers.TimerWorker
	AuthRefreshTicks <-chan struct{}
}

// NewCoreClientAuthInfra constructs shared logger/client infra and wires auth
// token manager plus a refresh timer worker.
func NewCoreClientAuthInfra(
	serviceName string,
	loggingConfig sharedconfig.LoggingConfig,
	messagingConfig sharedconfig.MessagingConfig,
	authConfig sharedconfig.AuthConfig,
	refreshInterval time.Duration,
) (CoreClientAuthInfra, error) {
	core, err := NewCoreClientInfra(serviceName, loggingConfig, messagingConfig)
	if err != nil {
		return CoreClientAuthInfra{}, err
	}

	module, err := buildCoreClientAuthInfra(core, authConfig, refreshInterval)
	if err != nil {
		core.Close()
		return CoreClientAuthInfra{}, err
	}
	return module, nil
}

func buildCoreClientAuthInfra(
	core CoreInfra,
	authConfig sharedconfig.AuthConfig,
	refreshInterval time.Duration,
) (CoreClientAuthInfra, error) {
	serviceLogin := authConfig.ServiceLogin
	if serviceLogin == "" {
		serviceLogin = authConfig.ServiceName
	}

	manager, err := servicetoken.NewManager(core.Client, servicetoken.Options{
		Service: serviceLogin,
	})
	if err != nil {
		return CoreClientAuthInfra{}, err
	}

	refreshTimer, refreshTicks, err := sharedworkers.NewPollingTimerWorker(refreshInterval, defaultAuthRefreshTickBuffer)
	if err != nil {
		return CoreClientAuthInfra{}, err
	}

	return CoreClientAuthInfra{
		CoreInfra:        core,
		AuthTokenManager: manager,
		AuthRefreshTimer: refreshTimer,
		AuthRefreshTicks: refreshTicks,
	}, nil
}
