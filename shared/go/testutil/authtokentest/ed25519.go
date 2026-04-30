package authtokentest

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

// MustGenerateEd25519Key returns an Ed25519 keypair or fails the current test.
func MustGenerateEd25519Key(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() error = %v", err)
	}

	return publicKey, privateKey
}

// MustMarshalEd25519PrivateKeyPEM returns a PKCS#8 PEM block for an Ed25519
// private key or fails the current test.
func MustMarshalEd25519PrivateKeyPEM(t *testing.T, privateKey ed25519.PrivateKey) []byte {
	t.Helper()

	derBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("x509.MarshalPKCS8PrivateKey() error = %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: derBytes})
}

// MustCreateEd25519CertificatePEM returns a self-signed certificate PEM block
// for an Ed25519 public key or fails the current test.
func MustCreateEd25519CertificatePEM(
	t *testing.T,
	publicKey ed25519.PublicKey,
	privateKey ed25519.PrivateKey,
) []byte {
	t.Helper()

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "lite-nas-auth-token-signing"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() error = %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}
