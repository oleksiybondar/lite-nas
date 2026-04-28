package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	sharedlogger "lite-nas/shared/logger"
	sharedmodules "lite-nas/shared/modules"
	"lite-nas/shared/testutil/messagingtest"
)

func loadInfraFixture() (Infra, *messagingtest.RecordingClient, *messagingtest.RecordingServer, *int) {
	client := &messagingtest.RecordingClient{}
	server := &messagingtest.RecordingServer{}

	return Infra{
		CoreInfra: sharedmodules.CoreInfra{
			Logger: sharedlogger.NewNop(),
			Client: client,
			Server: server,
		},
		Config: serviceconfig.Config{},
	}, client, server, nil
}
