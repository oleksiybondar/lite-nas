package authtoken

import (
	"testing"

	"lite-nas/shared/testutil/cryptotest"
)

func TestParseEd25519PrivateKeyPEM(t *testing.T) {
	t.Parallel()

	_, privateKey := cryptotest.MustGenerateEd25519Key(t)
	keyPEM := cryptotest.MustMarshalEd25519PrivateKeyPEM(t, privateKey)

	got, err := ParseEd25519PrivateKeyPEM(keyPEM)
	if err != nil {
		t.Fatalf("ParseEd25519PrivateKeyPEM() error = %v", err)
	}

	if !got.Equal(privateKey) {
		t.Fatal("parsed private key does not match original key")
	}
}

func TestParseEd25519CertificatePublicKeyPEM(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := cryptotest.MustGenerateEd25519Key(t)
	certPEM := cryptotest.MustCreateEd25519CertificatePEM(t, publicKey, privateKey)

	got, err := ParseEd25519CertificatePublicKeyPEM(certPEM)
	if err != nil {
		t.Fatalf("ParseEd25519CertificatePublicKeyPEM() error = %v", err)
	}

	if !got.Equal(publicKey) {
		t.Fatal("parsed certificate public key does not match original key")
	}
}

func TestParsedPEMKeysCreateIssuerAndVerifier(t *testing.T) {
	t.Parallel()

	publicKey, privateKey := cryptotest.MustGenerateEd25519Key(t)
	signingKey, err := ParseEd25519PrivateKeyPEM(cryptotest.MustMarshalEd25519PrivateKeyPEM(t, privateKey))
	if err != nil {
		t.Fatalf("ParseEd25519PrivateKeyPEM() error = %v", err)
	}

	verificationKey, err := ParseEd25519CertificatePublicKeyPEM(
		cryptotest.MustCreateEd25519CertificatePEM(t, publicKey, privateKey),
	)
	if err != nil {
		t.Fatalf("ParseEd25519CertificatePublicKeyPEM() error = %v", err)
	}

	issuer := mustNewIssuer(t, signingKey)
	verifier := mustNewVerifier(t, verificationKey)
	tokenText := mustIssueToken(t, issuer)
	if _, err = verifier.Verify(tokenText); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}
