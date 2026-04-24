module lite-nas/apps/system-metrics-cli

go 1.25.0

// lite-nas/shared is a local module dependency used by this application.
require lite-nas/shared v0.0.1

// These indirect dependencies are pulled in through shared packages and other
// direct dependencies of this application.
require (
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/nats-io/nats.go v1.51.0 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	gopkg.in/ini.v1 v1.67.1
)

replace lite-nas/shared => ./../../shared/go
