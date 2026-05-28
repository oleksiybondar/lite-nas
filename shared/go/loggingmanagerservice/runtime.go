package loggingmanagerservice

import (
	"context"
	"os"

	"lite-nas/shared/authtoken"
	sharedconfig "lite-nas/shared/config"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedworkers "lite-nas/shared/workers"
)

// Run starts and blocks a logging-manager service runtime.
func Run(
	ctx context.Context,
	infra Infra,
	subjects sharedloggingmanagernats.Subjects,
	authorizationPolicy sharedloggingmanagernats.AuthorizationPolicy,
	configPath string,
	logName string,
) error {
	if err := configureAuthMiddleware(infra, authorizationPolicy); err != nil {
		return err
	}

	if err := registerHandlers(infra, subjects); err != nil {
		return err
	}

	flushTimer, err := newFlushTimer(infra)
	if err != nil {
		return err
	}

	startWorkers(ctx, infra, flushTimer)
	logServiceLifecycle(ctx, infra, configPath, logName)
	return ctx.Err()
}

func configureAuthMiddleware(
	infra Infra,
	authorizationPolicy sharedloggingmanagernats.AuthorizationPolicy,
) error {
	authVerifier, err := newAuthTokenVerifier(infra.Config.AuthTokens)
	if err != nil {
		return err
	}

	infra.Server.UseSubscriptionMiddleware(
		sharedloggingmanagernats.NewAccessTokenValidationSubscriptionMiddleware(authVerifier),
		sharedloggingmanagernats.NewRoleAuthorizationSubscriptionMiddleware(authorizationPolicy),
	)
	infra.Server.UseRPCMiddleware(
		sharedloggingmanagernats.NewAccessTokenValidationRPCMiddleware(authVerifier),
		sharedloggingmanagernats.NewRoleAuthorizationRPCMiddleware(authorizationPolicy),
	)
	return nil
}

func registerHandlers(infra Infra, subjects sharedloggingmanagernats.Subjects) error {
	if err := sharedloggingmanagernats.RegisterSubscriptions(infra.Server, infra.LoggingManagerCore.Core, subjects); err != nil {
		return err
	}
	if err := sharedloggingmanagernats.RegisterRPCHandlers(infra.Server, infra.LoggingManagerCore.Core, subjects); err != nil {
		return err
	}
	return nil
}

func newFlushTimer(infra Infra) (sharedworkers.TimerWorker, error) {
	return sharedworkers.NewTimerWorker(
		sharedworkers.TimerConfig{
			Interval:    infra.Config.LoggingManager.Writer.FlushInterval,
			EmitOnStart: false,
		},
		infra.LoggingManagerCore.WriterFlushCh,
	)
}

func startWorkers(ctx context.Context, infra Infra, flushTimer sharedworkers.TimerWorker) {
	infra.LoggingManagerCore.CleanupTimer.Start(ctx)
	flushTimer.Start(ctx)
	go runCleanupWorker(ctx, infra.LoggingManagerCore.Core, infra.LoggingManagerCore.CleanupTicksCh, infra.Logger)
	go runWriterWorker(ctx, infra.LoggingManagerCore.Writer, infra.Logger)
}

func logServiceLifecycle(ctx context.Context, infra Infra, configPath string, logName string) {
	infra.Logger.Info(logName+" service started", "config", configPath)
	<-ctx.Done()
	infra.Logger.Info(logName + " service stopping")
}

func newAuthTokenVerifier(cfg sharedconfig.AuthTokenConfig) (authtoken.Verifier, error) {
	verificationCertData, err := os.ReadFile(cfg.VerificationCert) // #nosec G304 -- path comes from service config
	if err != nil {
		return authtoken.Verifier{}, err
	}

	verificationKey, err := authtoken.ParseEd25519CertificatePublicKeyPEM(verificationCertData)
	if err != nil {
		return authtoken.Verifier{}, err
	}

	return authtoken.NewVerifier(authTokenVerifierOptions(cfg), verificationKey)
}

func authTokenVerifierOptions(cfg sharedconfig.AuthTokenConfig) authtoken.VerifierOptions {
	return authtoken.VerifierOptions{
		Issuer:    cfg.Issuer,
		Audience:  cfg.Audience,
		ClockSkew: cfg.ClockSkew,
	}
}

func runCleanupWorker(
	ctx context.Context,
	core *sharedloggingmanager.Core,
	ticks <-chan struct{},
	log interface {
		Warn(msg string, args ...any)
	},
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			if err := core.Cleanup(ctx); err != nil {
				log.Warn("logging manager cleanup failed", "error", err)
			}
		}
	}
}

func runWriterWorker(
	ctx context.Context,
	writer *sharedloggingmanager.Writer,
	log interface {
		Error(msg string, args ...any)
	},
) {
	if err := writer.Run(ctx); err != nil {
		log.Error("logging manager writer stopped with error", "error", err.Error())
	}
}
