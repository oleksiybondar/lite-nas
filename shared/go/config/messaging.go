package config

import (
	"time"

	"gopkg.in/ini.v1"
)

// MessagingConfig defines settings required to connect to NATS.
type MessagingConfig struct {
	URL        string
	ClientName string
	CA         string
	Cert       string
	Key        string
	Timeout    time.Duration
}

// LoadMessagingConfig extracts and parses the [messaging] section from the INI file.
//
// Expected keys:
//   - url: NATS server address (e.g. "nats://localhost:4222")
//   - client_name: logical name of the client/service
//   - ca: path to CA certificate
//   - cert: path to client certificate
//   - key: path to client private key
//   - timeout: duration string (e.g. "5s")
//
// timeout defaults to "5s" if not provided.
// An error is returned if parsing fails.
func LoadMessagingConfig(cfgFile *ini.File) (MessagingConfig, error) {
	section := cfgFile.Section("messaging")

	timeout, err := time.ParseDuration(section.Key("timeout").MustString("5s"))
	if err != nil {
		return MessagingConfig{}, err
	}

	return MessagingConfig{
		URL:        section.Key("url").String(),
		ClientName: section.Key("client_name").String(),
		CA:         section.Key("ca").String(),
		Cert:       section.Key("cert").String(),
		Key:        section.Key("key").String(),
		Timeout:    timeout,
	}, nil
}
