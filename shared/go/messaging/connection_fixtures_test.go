package messaging

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lite-nas/shared/config"

	"github.com/nats-io/nats.go"
)

func loadConnectionOptionsFixture(t *testing.T, cfg config.MessagingConfig) []nats.Option {
	t.Helper()

	options, err := buildConnectionOptions(cfg, &recordingLogger{})
	if err != nil {
		t.Fatalf("buildConnectionOptions() error = %v", err)
	}

	return options
}

func loadTLSConfigFixture(t *testing.T, cfg config.MessagingConfig) *tls.Config {
	t.Helper()

	tlsConfig, err := buildTLSConfig(cfg)
	if err != nil {
		t.Fatalf("buildTLSConfig() error = %v", err)
	}

	return tlsConfig
}

func loadClientCertificatesFixture(t *testing.T, cfg config.MessagingConfig) []tls.Certificate {
	t.Helper()

	certificates, err := loadClientCertificates(cfg)
	if err != nil {
		t.Fatalf("loadClientCertificates() error = %v", err)
	}

	return certificates
}

func loadRootCAsFixture(t *testing.T, cfg config.MessagingConfig) (*x509.CertPool, bool) {
	t.Helper()

	rootCAs, ok, err := loadRootCAs(cfg)
	if err != nil {
		t.Fatalf("loadRootCAs() error = %v", err)
	}

	return rootCAs, ok
}

func writeTLSFixture(t *testing.T) (string, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	certPEM, keyPEM := generateSelfSignedCertificate(t)

	certPath := filepath.Join(tempDir, "client.crt")
	keyPath := filepath.Join(tempDir, "client.key")
	caPath := filepath.Join(tempDir, "root-ca.crt")

	writeTestFile(t, certPath, certPEM)
	writeTestFile(t, keyPath, keyPEM)
	writeTestFile(t, caPath, certPEM)

	return certPath, keyPath, caPath
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}

func generateSelfSignedCertificate(t *testing.T) ([]byte, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "lite-nas-test",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() error = %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return certPEM, keyPEM
}
