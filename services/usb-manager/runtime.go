package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/config"
	sharedcontracts "lite-nas/shared/contracts"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	contract "lite-nas/shared/contracts/usbmanager"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/roleauth"
	"lite-nas/shared/workers"
)

const (
	packagedConfigPath = "/etc/lite-nas/usb-manager.conf"
	usbguardBinary     = "/usr/bin/usbguard"
	maxCommandErrorLen = 4096
)

var deviceIDPattern = regexp.MustCompile(`^[0-9]{1,10}$`)

// commandRunner is the injected host command boundary.
type commandRunner interface {
	Output(context.Context, string, ...string) ([]byte, error)
}

// execRunner executes a command without a shell.
type execRunner struct{}

// Output runs a validated fixed executable and argument vector.
func (execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// commandJob represents one validated operation queued for the USBGuard worker.
type commandJob struct {
	ctx       context.Context
	action    string
	deviceID  string
	permanent bool
	result    chan commandResult
}

// commandResult carries the worker output or error to its owning RPC callback.
type commandResult struct {
	output []byte
	err    error
}

// manager owns token validation and the bounded USBGuard command queue.
// RPC callbacks produce jobs (capacity 32); the worker consumes them.
type manager struct {
	verifier  authtoken.Verifier
	validator *validator.Validate
	runner    commandRunner
	jobs      chan commandJob
	ctx       context.Context
	client    messaging.Client
	logger    sharedlogger.Logger
}

// run loads the packaged service configuration and starts the NATS service.
func run(ctx context.Context) error {
	return runWithDependencies(ctx, packagedConfigPath, modules.NewCoreClientServerInfra, execRunner{})
}

// runWithDependencies keeps host integration explicit for lifecycle tests.
func runWithDependencies(ctx context.Context, configPath string, infraFactory modules.CoreInfraConstructor, runner commandRunner) error {
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	iniFile, err := config.LoadINI(reader)
	if err != nil {
		return err
	}
	cfg, err := config.LoadSharedAuthTokenConfig(iniFile)
	if err != nil {
		return err
	}
	certificate, err := os.ReadFile(cfg.AuthTokens.VerificationCert)
	if err != nil {
		return err
	}
	key, err := authtoken.ParseEd25519CertificatePublicKeyPEM(certificate)
	if err != nil {
		return err
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: cfg.AuthTokens.Issuer, Audience: cfg.AuthTokens.Audience, ClockSkew: cfg.AuthTokens.ClockSkew}, key)
	if err != nil {
		return err
	}
	infra, err := infraFactory(sharedcontracts.ServiceUSBManager, cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	service := &manager{verifier: verifier, validator: validator.New(validator.WithRequiredStructEnabled()), runner: runner, jobs: make(chan commandJob, 32), ctx: ctx, client: infra.Client, logger: infra.Logger}
	if err := service.register(infra.Server); err != nil {
		return err
	}
	done := make(chan struct{})
	workers.NewBaseWorker(func(ctx context.Context) { defer close(done); service.work(ctx) }).Start(ctx)
	infra.Logger.Info("usb manager started")
	<-ctx.Done()
	if err := infra.Server.Drain(); err != nil {
		infra.Logger.Error("usb manager drain failed", "error", err)
	}
	<-done
	return ctx.Err()
}

// register binds each fixed subject to an authenticated handler.
func (m *manager) register(server messaging.Server) error {
	if err := server.RegisterRPC(contract.ListSubject, m.list); err != nil {
		return err
	}
	for _, operation := range []struct{ subject, action string }{{contract.AllowSubject, "allow-device"}, {contract.BlockSubject, "block-device"}} {
		action := operation.action
		if err := server.RegisterRPC(operation.subject, func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			return m.action(ctx, envelope, action)
		}); err != nil {
			return err
		}
	}
	return nil
}

// decode rejects malformed payloads and callers without a security role.
func decode[T any](ctx context.Context, m *manager, envelope messaging.Envelope, accessToken func(T) string) (T, authtoken.AccessClaims, error) {
	var request T
	if len(envelope.Payload) > 16*1024 {
		return request, authtoken.AccessClaims{}, fmt.Errorf("request too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, authtoken.AccessClaims{}, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return request, authtoken.AccessClaims{}, fmt.Errorf("unexpected trailing JSON")
	}
	if err := m.validator.Struct(request); err != nil {
		return request, authtoken.AccessClaims{}, err
	}
	claims, err := authtoken.VerifyLive(ctx, m.client, m.verifier, accessToken(request))
	if err != nil || !roleauth.MatchesRequirement(claims.Roles, roleauth.RequirementSecurity) {
		return request, authtoken.AccessClaims{}, fmt.Errorf("security authorization required")
	}
	return request, claims, nil
}

// list returns the current usbguard device inventory.
func (m *manager) list(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, _, err := decode(ctx, m, envelope, func(request contract.ListRequest) string { return request.AccessToken }); err != nil {
		return nil, err
	}
	output, err := m.submit(ctx, "list-devices", "", false)
	if err != nil {
		return nil, err
	}
	return contract.ListResponse{Devices: strings.TrimSpace(string(output))}, nil
}

// action validates the USBGuard numeric device ID before queuing a decision.
func (m *manager) action(ctx context.Context, envelope messaging.Envelope, action string) (any, error) {
	request, claims, err := decode(ctx, m, envelope, func(request contract.DeviceRequest) string { return request.AccessToken })
	if err != nil {
		return nil, err
	}
	if !deviceIDPattern.MatchString(request.DeviceID) {
		return nil, fmt.Errorf("invalid USBGuard device ID")
	}
	if _, err := m.submit(ctx, action, request.DeviceID, request.Permanent); err != nil {
		return nil, err
	}
	message := fmt.Sprintf("USBGuard %s applied to device %s by %s", strings.TrimSuffix(action, "-device"), request.DeviceID, claims.Login)
	permanence := "temporary"
	if request.Permanent {
		permanence = "permanent"
	}
	payload := loggingmanagercontract.AlertPayload{
		AccessToken: request.AccessToken,
		Category:    "usbguard",
		Severity:    sharedloggingenum.SeverityInfo,
		Source:      sharedcontracts.ServiceUSBManager,
		Message:     message,
		TriggerValue: fmt.Sprintf("action=%s device_id=%s permanence=%s actor=%s", strings.TrimSuffix(action, "-device"),
			request.DeviceID, permanence, claims.Login),
	}
	if err := m.client.Publish(ctx, securityloggingmanagercontract.AlertSubject, payload); err != nil {
		m.logger.Error("failed to publish USBGuard action security event", "action", action, "device_id", request.DeviceID, "error", err)
	}
	return contract.ActionResponse{Completed: true}, nil
}

// submit applies backpressure and cancellation to the command worker queue.
func (m *manager) submit(ctx context.Context, action, deviceID string, permanent bool) ([]byte, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	job := commandJob{ctx: requestCtx, action: action, deviceID: deviceID, permanent: permanent, result: make(chan commandResult, 1)}
	select {
	case m.jobs <- job:
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}
	select {
	case result := <-job.result:
		return result.output, result.err
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}
}

// work serializes USBGuard commands and accounts for queued jobs on shutdown.
func (m *manager) work(ctx context.Context) {
	for {
		select {
		case job := <-m.jobs:
			if ctx.Err() != nil {
				job.result <- commandResult{err: ctx.Err()}
				continue
			}
			job.result <- m.execute(job)
		case <-ctx.Done():
			for {
				select {
				case job := <-m.jobs:
					job.result <- commandResult{err: ctx.Err()}
				default:
					return
				}
			}
		}
	}
}

// execute constructs only supported usbguard command forms.
func (m *manager) execute(job commandJob) commandResult {
	if err := job.ctx.Err(); err != nil {
		return commandResult{err: err}
	}
	ctx, cancel := context.WithTimeout(job.ctx, 20*time.Second)
	defer cancel()
	var output []byte
	var err error
	if job.action == "list-devices" {
		output, err = m.runner.Output(ctx, "/usr/bin/sudo", "-n", usbguardBinary, "list-devices")
	} else {
		args := []string{job.action}
		if job.permanent {
			args = append(args, "--permanent")
		}
		args = append(args, job.deviceID)
		output, err = m.runner.Output(ctx, "/usr/bin/sudo", append([]string{"-n", usbguardBinary}, args...)...)
	}
	if err != nil {
		detail := ""
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			detail = strings.TrimSpace(string(exitErr.Stderr))
			if len(detail) > maxCommandErrorLen {
				detail = detail[:maxCommandErrorLen]
			}
		}
		if detail != "" {
			return commandResult{err: fmt.Errorf("usbguard %s failed: %w: %s", job.action, err, detail)}
		}
		return commandResult{err: fmt.Errorf("usbguard %s failed: %w", job.action, err)}
	}
	if len(output) > 1024*1024 {
		return commandResult{err: fmt.Errorf("usbguard output too large")}
	}
	return commandResult{output: output}
}
