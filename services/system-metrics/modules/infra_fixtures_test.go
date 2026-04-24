package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	sharedlogger "lite-nas/shared/logger"
)

func loadInfraFixture() (Infra, *recordingMessagingClient, *recordingMessagingServer, *int) {
	client := &recordingMessagingClient{}
	server := &recordingMessagingServer{}
	log := sharedlogger.NewNop()
	cleanupCalls := 0

	return Infra{
		config: serviceconfig.Config{},
		logger: log,
		logCleanup: func() {
			cleanupCalls++
		},
		client: client,
		server: server,
	}, client, server, &cleanupCalls
}
