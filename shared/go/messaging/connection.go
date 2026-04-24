package messaging

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"time"

	"lite-nas/shared/config"
	"lite-nas/shared/logger"

	"github.com/nats-io/nats.go"
)

const (
	// defaultReconnectWait defines the delay between reconnect attempts.
	defaultReconnectWait = 2 * time.Second

	// defaultMaxReconnects defines the reconnect attempt limit.
	// A negative value means reconnect indefinitely.
	defaultMaxReconnects = -1
)

// connection is a low-level wrapper around a NATS connection.
//
// Design choice:
//   - this type owns only runtime transport state
//   - it does not own callback wiring policy beyond using prepared options
//   - higher-level client/server abstractions should build on top of it
type connection struct {
	conn   *nats.Conn
	logger logger.Logger
}

// connHandlerFunc defines reusable logic for NATS callbacks with the
// nats.ConnHandler signature once logger injection is made explicit.
//
// Design choice:
//   - handler logic is kept in plain named functions
//   - logger is passed explicitly instead of being captured by inline logic
type connHandlerFunc func(
	log logger.Logger,
	nc *nats.Conn,
)

// connErrHandlerFunc defines reusable logic for NATS callbacks with the
// nats.ConnErrHandler signature once logger injection is made explicit.
type connErrHandlerFunc func(
	log logger.Logger,
	nc *nats.Conn,
	err error,
)

// asyncErrHandlerFunc defines reusable logic for NATS callbacks with the
// nats.ErrHandler signature once logger injection is made explicit.
type asyncErrHandlerFunc func(
	log logger.Logger,
	nc *nats.Conn,
	sub *nats.Subscription,
	err error,
)

// newConnection creates and establishes a NATS connection using the provided
// messaging configuration and logger.
//
// Design choice:
//   - configuration validation happens before transport creation
//   - reconnect and lifecycle logging are delegated to connection options
func newConnection(
	cfg config.MessagingConfig,
	log logger.Logger,
) (*connection, error) {
	if cfg.URL == "" {
		return nil, ErrInvalidConfig
	}

	opts, err := buildConnectionOptions(cfg, log)
	if err != nil {
		return nil, err
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotConnected, err)
	}

	logConnectionReady(log, conn, cfg)

	return &connection{
		conn:   conn,
		logger: log,
	}, nil
}

// publish sends a raw payload to the given subject.
func (c *connection) publish(subject string, payload []byte) error {
	if !c.isConnected() {
		return ErrNotConnected
	}

	if subject == "" {
		return ErrInvalidSubject
	}

	if err := c.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("%w: %w", ErrPublishFailed, err)
	}

	return nil
}

// request sends a request payload and waits for a reply within the provided
// timeout.
func (c *connection) request(
	subject string,
	payload []byte,
	timeout time.Duration,
) (*nats.Msg, error) {
	if !c.isConnected() {
		return nil, ErrNotConnected
	}

	if subject == "" {
		return nil, ErrInvalidSubject
	}

	msg, err := c.conn.Request(subject, payload, timeout)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			return nil, fmt.Errorf("%w: %w", ErrRequestTimeout, err)
		}

		return nil, fmt.Errorf("%w: %w", ErrPublishFailed, err)
	}

	return msg, nil
}

// subscribe registers a raw NATS message handler for the given subject.
func (c *connection) subscribe(
	subject string,
	handler nats.MsgHandler,
) error {
	if !c.canSubscribe() {
		return ErrNotConnected
	}

	if subject == "" {
		return ErrInvalidSubject
	}

	if handler == nil {
		return ErrHandlerFailed
	}

	_, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSubscribeFailed, err)
	}

	c.logger.Info("nats subscription registered", "subject", subject)

	return nil
}

// drain gracefully drains the connection before shutdown.
func (c *connection) drain() error {
	if c == nil || c.conn == nil {
		return nil
	}

	c.logger.Info("nats draining connection")

	if err := c.conn.Drain(); err != nil {
		return fmt.Errorf("messaging: drain failed: %w", err)
	}

	return nil
}

// close immediately closes the underlying connection.
func (c *connection) close() {
	if c == nil || c.conn == nil {
		return
	}

	c.logger.Info("nats connection closed")
	c.conn.Close()
}

// isConnected reports whether the underlying connection is active.
func (c *connection) isConnected() bool {
	return c != nil && c.conn != nil && c.conn.IsConnected()
}

// canSubscribe reports whether subscription registration can be attempted on
// the underlying connection.
//
// Design choice:
//   - subscriptions should be allowed while reconnecting so startup can
//     complete before the first successful broker connection
//   - closed or draining connections remain invalid for new subscriptions
func (c *connection) canSubscribe() bool {
	return c != nil &&
		c.conn != nil &&
		!c.conn.IsClosed() &&
		!c.conn.IsDraining()
}

// buildConnectionOptions constructs NATS connection options from messaging
// configuration.
//
// Design choice:
//   - lifecycle callback logic is attached through small decorators
//   - decorators only inject logger and forward arguments
//   - actual callback behavior lives in named handler functions
func buildConnectionOptions(
	cfg config.MessagingConfig,
	log logger.Logger,
) ([]nats.Option, error) {
	opts := []nats.Option{
		nats.Name(cfg.ClientName),
		nats.Timeout(cfg.Timeout),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(defaultMaxReconnects),
		nats.ReconnectWait(defaultReconnectWait),
		nats.DisconnectErrHandler(buildConnErrHandler(log, handleDisconnectErr)),
		nats.ReconnectHandler(buildConnHandler(log, handleReconnect)),
		nats.ClosedHandler(buildConnHandler(log, handleClosed)),
		nats.ErrorHandler(buildAsyncErrHandler(log, handleAsyncError)),
	}

	if hasTLSConfig(cfg) {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, err
		}

		opts = append(opts, nats.Secure(tlsConfig))
	}

	return opts, nil
}

func logConnectionReady(
	log logger.Logger,
	conn *nats.Conn,
	cfg config.MessagingConfig,
) {
	if conn.IsConnected() {
		log.Info(
			"nats connected",
			"url", cfg.URL,
			"client", cfg.ClientName,
		)
		return
	}

	log.Warn(
		"nats initial connect deferred",
		"url", cfg.URL,
		"client", cfg.ClientName,
		"status", conn.Status().String(),
	)
}

// hasTLSConfig reports whether any TLS-related settings are configured.
func hasTLSConfig(cfg config.MessagingConfig) bool {
	return cfg.CA != "" || cfg.Cert != "" || cfg.Key != ""
}

// buildTLSConfig constructs a TLS configuration from the provided messaging
// certificate settings.
func buildTLSConfig(cfg config.MessagingConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	certificates, err := loadClientCertificates(cfg)
	if err != nil {
		return nil, err
	}

	if len(certificates) > 0 {
		tlsConfig.Certificates = certificates
	}

	rootCAs, hasRootCAs, err := loadRootCAs(cfg)
	if err != nil {
		return nil, err
	}

	if hasRootCAs {
		tlsConfig.RootCAs = rootCAs
	}

	return tlsConfig, nil
}

// loadClientCertificates loads a client certificate pair when both certificate
// and key are configured.
func loadClientCertificates(cfg config.MessagingConfig) ([]tls.Certificate, error) {
	if cfg.Cert == "" && cfg.Key == "" {
		return nil, nil
	}

	if cfg.Cert == "" || cfg.Key == "" {
		return nil, fmt.Errorf("%w: both cert and key must be provided", ErrInvalidConfig)
	}

	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("messaging: failed to load client cert: %w", err)
	}

	return []tls.Certificate{cert}, nil
}

// loadRootCAs loads the configured CA certificate pool.
//
// Design choice:
//   - returns an explicit presence flag to avoid ambiguous (nil, nil) results
func loadRootCAs(cfg config.MessagingConfig) (*x509.CertPool, bool, error) {
	if cfg.CA == "" {
		return nil, false, nil
	}

	caCert, err := os.ReadFile(cfg.CA)
	if err != nil {
		return nil, false, fmt.Errorf("messaging: failed to read CA: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, false, fmt.Errorf("messaging: failed to parse CA")
	}

	return pool, true, nil
}

// buildConnHandler adapts a named connection handler into the callback shape
// required by NATS.
//
// Design choice:
//   - this is a thin decorator only
//   - it injects logger and forwards call arguments
//   - it intentionally contains no business or logging logic of its own
func buildConnHandler(
	log logger.Logger,
	handler connHandlerFunc,
) nats.ConnHandler {
	return func(nc *nats.Conn) {
		handler(log, nc)
	}
}

// buildConnErrHandler adapts a named connection error handler into the callback
// shape required by NATS.
func buildConnErrHandler(
	log logger.Logger,
	handler connErrHandlerFunc,
) nats.ConnErrHandler {
	return func(nc *nats.Conn, err error) {
		handler(log, nc, err)
	}
}

// buildAsyncErrHandler adapts a named async error handler into the callback
// shape required by NATS.
func buildAsyncErrHandler(
	log logger.Logger,
	handler asyncErrHandlerFunc,
) nats.ErrHandler {
	return func(nc *nats.Conn, sub *nats.Subscription, err error) {
		handler(log, nc, sub, err)
	}
}

// handleDisconnectErr logs NATS disconnect events.
func handleDisconnectErr(
	log logger.Logger,
	nc *nats.Conn,
	err error,
) {
	if err != nil {
		log.Warn(
			"nats disconnected",
			"url", nc.ConnectedUrl(),
			"error", err.Error(),
		)
		return
	}

	log.Warn(
		"nats disconnected",
		"url", nc.ConnectedUrl(),
	)
}

// handleReconnect logs successful reconnect events.
func handleReconnect(
	log logger.Logger,
	nc *nats.Conn,
) {
	log.Info(
		"nats reconnected",
		"url", nc.ConnectedUrl(),
	)
}

// handleClosed logs final connection closure events.
func handleClosed(
	log logger.Logger,
	nc *nats.Conn,
) {
	log.Warn(
		"nats connection closed",
		"last_error", errorString(nc.LastError()),
	)
}

// handleAsyncError logs asynchronous NATS callback and subscription errors.
func handleAsyncError(
	log logger.Logger,
	_ *nats.Conn,
	sub *nats.Subscription,
	err error,
) {
	if sub != nil {
		log.Error(
			"nats async error",
			"subject", sub.Subject,
			"error", errorString(err),
		)
		return
	}

	log.Error(
		"nats async error",
		"error", errorString(err),
	)
}

// errorString safely converts an error to string.
func errorString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
