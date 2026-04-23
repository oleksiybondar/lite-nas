package messaging

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lite-nas/shared/config"

	"github.com/nats-io/nats.go"
)

func TestNewConnectionRejectsMissingURL(t *testing.T) {
	t.Parallel()

	_, err := newConnection(config.MessagingConfig{}, &recordingLogger{})
	assertErrorIs(t, err, ErrInvalidConfig)
}

func TestNewConnectionReturnsTLSConfigErrorBeforeConnect(t *testing.T) {
	t.Parallel()

	_, err := newConnection(config.MessagingConfig{
		URL:  "nats://localhost:4222",
		Cert: "/missing/client.crt",
		Key:  "/missing/client.key",
	}, &recordingLogger{})
	if err == nil {
		t.Fatal("expected TLS config error")
	}
}

func TestConnectionPublishRequiresConnection(t *testing.T) {
	t.Parallel()

	err := (&connection{}).publish("subject", []byte("payload"))
	assertErrorIs(t, err, ErrNotConnected)
}

func TestConnectionRequestRequiresConnection(t *testing.T) {
	t.Parallel()

	_, err := (&connection{}).request("subject", []byte("payload"), time.Second)
	assertErrorIs(t, err, ErrNotConnected)
}

func TestConnectionSubscribeRequiresConnection(t *testing.T) {
	t.Parallel()

	err := (&connection{}).subscribe("subject", func(_ *nats.Msg) {})
	assertErrorIs(t, err, ErrNotConnected)
}

func TestConnectionDrainIsSafeForNilState(t *testing.T) {
	t.Parallel()

	var nilConn *connection
	if err := nilConn.drain(); err != nil {
		t.Fatalf("drain() error = %v", err)
	}
}

func TestConnectionDrainIsSafeForEmptyState(t *testing.T) {
	t.Parallel()

	if err := (&connection{}).drain(); err != nil {
		t.Fatalf("drain() error = %v", err)
	}
}

func TestConnectionCloseIsSafeForNilState(t *testing.T) {
	t.Parallel()

	var nilConn *connection
	nilConn.close()
}

func TestConnectionCloseIsSafeForEmptyState(t *testing.T) {
	t.Parallel()

	(&connection{}).close()
}

func TestConnectionIsConnectedReturnsFalseForNilConnection(t *testing.T) {
	t.Parallel()

	var nilConn *connection
	if nilConn.isConnected() {
		t.Fatal("expected nil connection to report disconnected")
	}
}

func TestConnectionIsConnectedReturnsFalseForEmptyConnection(t *testing.T) {
	t.Parallel()

	if (&connection{}).isConnected() {
		t.Fatal("expected empty connection to report disconnected")
	}
}

func TestBuildConnectionOptionsWithoutTLS(t *testing.T) {
	t.Parallel()

	options := loadConnectionOptionsFixture(t, config.MessagingConfig{
		ClientName: "system-metrics",
		Timeout:    5 * time.Second,
	})

	if len(options) != 8 {
		t.Fatalf("len(options) = %d, want 8", len(options))
	}
}

func TestBuildConnectionOptionsWithTLS(t *testing.T) {
	t.Parallel()

	certPath, keyPath, caPath := writeTLSFixture(t)
	options := loadConnectionOptionsFixture(t, config.MessagingConfig{
		ClientName: "system-metrics",
		Timeout:    5 * time.Second,
		Cert:       certPath,
		Key:        keyPath,
		CA:         caPath,
	})

	if len(options) != 9 {
		t.Fatalf("len(options) = %d, want 9", len(options))
	}
}

func TestHasTLSConfigDetectsAnyTLSSetting(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cfg  config.MessagingConfig
		want bool
	}{
		{name: "empty", cfg: config.MessagingConfig{}, want: false},
		{name: "ca only", cfg: config.MessagingConfig{CA: "/tmp/ca.pem"}, want: true},
		{name: "cert only", cfg: config.MessagingConfig{Cert: "/tmp/cert.pem"}, want: true},
		{name: "key only", cfg: config.MessagingConfig{Key: "/tmp/key.pem"}, want: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := hasTLSConfig(testCase.cfg); got != testCase.want {
				t.Fatalf("hasTLSConfig() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestBuildTLSConfigWithoutCertificatesAndRootCA(t *testing.T) {
	t.Parallel()

	tlsConfig := loadTLSConfigFixture(t, config.MessagingConfig{})
	if len(tlsConfig.Certificates) != 0 || tlsConfig.RootCAs != nil {
		t.Fatalf("unexpected empty TLS config: %#v", tlsConfig)
	}
}

func TestBuildTLSConfigWithCertificatesAndRootCA(t *testing.T) {
	t.Parallel()

	certPath, keyPath, caPath := writeTLSFixture(t)
	tlsConfig := loadTLSConfigFixture(t, config.MessagingConfig{
		Cert: certPath,
		Key:  keyPath,
		CA:   caPath,
	})

	if len(tlsConfig.Certificates) != 1 {
		t.Fatalf("len(Certificates) = %d, want 1", len(tlsConfig.Certificates))
	}

	if tlsConfig.RootCAs == nil {
		t.Fatal("expected RootCAs to be configured")
	}
}

func TestLoadClientCertificatesReturnsNilWhenUnset(t *testing.T) {
	t.Parallel()

	certificates := loadClientCertificatesFixture(t, config.MessagingConfig{})
	if certificates != nil {
		t.Fatalf("loadClientCertificates() = %#v, want nil", certificates)
	}
}

func TestLoadClientCertificatesRejectsPartialConfig(t *testing.T) {
	t.Parallel()

	_, err := loadClientCertificates(config.MessagingConfig{Cert: "/tmp/cert.pem"})
	assertErrorIs(t, err, ErrInvalidConfig)
}

func TestLoadClientCertificatesLoadsConfiguredPair(t *testing.T) {
	t.Parallel()

	certPath, keyPath, _ := writeTLSFixture(t)
	certificates := loadClientCertificatesFixture(t, config.MessagingConfig{
		Cert: certPath,
		Key:  keyPath,
	})

	if len(certificates) != 1 {
		t.Fatalf("len(certificates) = %d, want 1", len(certificates))
	}
}

func TestLoadRootCAsReturnsAbsentWhenUnset(t *testing.T) {
	t.Parallel()

	rootCAs, ok := loadRootCAsFixture(t, config.MessagingConfig{})
	if rootCAs != nil || ok {
		t.Fatalf("loadRootCAs() = (%v, %v), want (nil, false)", rootCAs, ok)
	}
}

func TestLoadRootCAsRejectsUnreadableCA(t *testing.T) {
	t.Parallel()

	_, _, err := loadRootCAs(config.MessagingConfig{CA: "/missing/ca.pem"})
	if err == nil {
		t.Fatal("expected CA read error")
	}
}

func TestLoadRootCAsRejectsInvalidCA(t *testing.T) {
	t.Parallel()

	badCAPath := filepath.Join(t.TempDir(), "bad-ca.pem")
	if err := os.WriteFile(badCAPath, []byte("not-a-cert"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, _, err := loadRootCAs(config.MessagingConfig{CA: badCAPath})
	if err == nil {
		t.Fatal("expected CA parse error")
	}
}

func TestLoadRootCAsLoadsConfiguredCA(t *testing.T) {
	t.Parallel()

	_, _, caPath := writeTLSFixture(t)
	rootCAs, ok := loadRootCAsFixture(t, config.MessagingConfig{CA: caPath})
	if rootCAs == nil || !ok {
		t.Fatalf("unexpected valid CA result: (%v, %v)", rootCAs, ok)
	}
}

func TestBuildConnHandlerForwardsLoggerAndConnection(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	nc := &nats.Conn{}
	capture := &connHandlerCapture{}

	buildConnHandler(log, capture.record)(nc)

	if !capture.called {
		t.Fatal("expected conn handler to be called")
	}

	if capture.log != log || capture.conn != nc {
		t.Fatalf("unexpected forwarded values: %#v", capture)
	}
}

func TestBuildConnErrHandlerForwardsLoggerConnectionAndError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	nc := &nats.Conn{}
	testErr := errors.New("boom")
	capture := &connErrHandlerCapture{}

	buildConnErrHandler(log, capture.record)(nc, testErr)

	if !capture.called {
		t.Fatal("expected conn error handler to be called")
	}

	if capture.log != log || capture.conn != nc || !errors.Is(capture.err, testErr) {
		t.Fatalf("unexpected forwarded values: %#v", capture)
	}
}

func TestBuildAsyncErrHandlerForwardsLoggerConnectionSubscriptionAndError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	nc := &nats.Conn{}
	sub := &nats.Subscription{Subject: "system.metrics"}
	testErr := errors.New("boom")
	capture := &asyncErrHandlerCapture{}

	buildAsyncErrHandler(log, capture.record)(nc, sub, testErr)

	if !capture.called {
		t.Fatal("expected async error handler to be called")
	}

	if capture.log != log || capture.conn != nc || capture.sub != sub || !errors.Is(capture.err, testErr) {
		t.Fatalf("unexpected forwarded values: %#v", capture)
	}
}

func TestHandleDisconnectErrLogsError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handleDisconnectErr(log, &nats.Conn{}, errors.New("disconnect"))

	assertSingleLogEntry(t, log, logEntry{
		level: "warn",
		msg:   "nats disconnected",
		args:  []any{"url", "", "error", "disconnect"},
	})
}

func TestHandleDisconnectErrLogsWithoutError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handleDisconnectErr(log, &nats.Conn{}, nil)

	assertSingleLogEntry(t, log, logEntry{
		level: "warn",
		msg:   "nats disconnected",
		args:  []any{"url", ""},
	})
}

func TestHandleReconnectLogsURL(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handleReconnect(log, &nats.Conn{})

	assertSingleLogEntry(t, log, logEntry{
		level: "info",
		msg:   "nats reconnected",
		args:  []any{"url", ""},
	})
}

func TestHandleClosedLogsLastError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handleClosed(log, &nats.Conn{})

	assertSingleLogEntry(t, log, logEntry{
		level: "warn",
		msg:   "nats connection closed",
		args:  []any{"last_error", ""},
	})
}

func TestHandleAsyncErrorLogsSubjectAndError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handleAsyncError(log, &nats.Conn{}, &nats.Subscription{Subject: "system.metrics"}, errors.New("async"))

	assertSingleLogEntry(t, log, logEntry{
		level: "error",
		msg:   "nats async error",
		args:  []any{"subject", "system.metrics", "error", "async"},
	})
}

func TestHandleAsyncErrorLogsErrorWithoutSubject(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handleAsyncError(log, &nats.Conn{}, nil, errors.New("async"))

	assertSingleLogEntry(t, log, logEntry{
		level: "error",
		msg:   "nats async error",
		args:  []any{"error", "async"},
	})
}

func TestErrorStringReturnsEmptyStringForNil(t *testing.T) {
	t.Parallel()

	if got := errorString(nil); got != "" {
		t.Fatalf("errorString(nil) = %q, want empty string", got)
	}
}

func TestErrorStringReturnsErrorText(t *testing.T) {
	t.Parallel()

	if got := errorString(errors.New("boom")); got != "boom" {
		t.Fatalf("errorString(err) = %q, want boom", got)
	}
}
