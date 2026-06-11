package config

import "gopkg.in/ini.v1"

// AuthConfig defines auth identity certificate settings shared by services/apps
// that request service tokens.
type AuthConfig struct {
	CA           string
	Cert         string
	Key          string
	ServiceName  string
	ServiceLogin string
}

// LoadAuthConfig extracts the [auth] section from an INI document.
func LoadAuthConfig(cfgFile *ini.File) (AuthConfig, error) {
	section := cfgFile.Section("auth")

	certificate := loadCertificateConfig(section)
	if certificate.CA == "" {
		// Backward compatibility during migration from root_ca to ca.
		certificate.CA = section.Key("root_ca").String()
	}

	return AuthConfig{
		CA:           certificate.CA,
		Cert:         certificate.Cert,
		Key:          certificate.Key,
		ServiceName:  section.Key("service_name").String(),
		ServiceLogin: section.Key("service_login").String(),
	}, nil
}
