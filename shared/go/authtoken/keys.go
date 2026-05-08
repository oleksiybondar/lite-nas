package authtoken

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ParseEd25519PrivateKeyPEM parses a PKCS#8 Ed25519 private key PEM.
//
// Parameters:
//   - data: PEM-encoded private-key bytes, usually read by service wiring from
//     the configured signing-key path.
//
// It returns the parsed Ed25519 private key. The function intentionally accepts
// bytes rather than a path so filesystem access stays outside this pure helper
// package.
func ParseEd25519PrivateKeyPEM(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errMissingPEMBlock
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: %T", errUnsupportedSigningKey, key)
	}

	return edKey, nil
}

// ParseEd25519CertificatePublicKeyPEM parses an X.509 certificate PEM and
// extracts its Ed25519 public key.
//
// Parameters:
//   - data: PEM-encoded X.509 certificate bytes, usually read by service wiring
//     from the configured verification-certificate path.
//
// It returns the certificate public key when the certificate contains an
// Ed25519 key. Certificate chain and validity policy are intentionally outside
// this parser; callers can add those checks in service/module wiring when
// needed.
func ParseEd25519CertificatePublicKeyPEM(data []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errMissingPEMBlock
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	key, ok := certificate.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("%w: %T", errUnsupportedPublicKey, certificate.PublicKey)
	}

	return key, nil
}
