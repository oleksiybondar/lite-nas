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
		Config: serviceconfig.Config{},
		Logger: log,
		logCleanup: func() {
			cleanupCalls++
		},
		Client: client,
		Server: server,
	}, client, server, &cleanupCalls
}
