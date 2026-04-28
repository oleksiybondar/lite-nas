module lite-nas/services/auth

go 1.25.0

// lite-nas/shared is a local module dependency used by this service.
require lite-nas/shared v0.0.1

require github.com/msteinert/pam/v2 v2.1.0

require gopkg.in/ini.v1 v1.67.1 // indirect

// These indirect dependencies are pulled in through shared packages and other
// direct dependencies of this service.
require (
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/nats-io/nats.go v1.51.0 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

replace lite-nas/shared => ./../../shared/go
