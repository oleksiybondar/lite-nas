package config

import "gopkg.in/ini.v1"

// CertificateConfig defines shared certificate file paths.
type CertificateConfig struct {
	CA   string
	Cert string
	Key  string
}

func loadCertificateConfig(section *ini.Section) CertificateConfig {
	return CertificateConfig{
		CA:   section.Key("ca").String(),
		Cert: section.Key("cert").String(),
		Key:  section.Key("key").String(),
	}
}
