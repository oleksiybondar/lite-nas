module lite-nas/services/resources-monitor

go 1.25.0

// lite-nas/shared is a local module dependency used by this service.
require lite-nas/shared v0.0.1

require gopkg.in/ini.v1 v1.67.1 // indirect

replace lite-nas/shared => ./../../shared/go
