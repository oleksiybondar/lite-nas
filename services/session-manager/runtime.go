package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	contract "lite-nas/shared/contracts/sessionmanager"
	"lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/roleauth"
	"lite-nas/shared/workers"
)

const (
	packagedConfigPath = "/etc/lite-nas/session-manager.conf"
	loginctlBinary     = "/usr/bin/loginctl"
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

// commandRunner is the injected host command boundary.
type commandRunner interface {
	Output(context.Context, string, ...string) ([]byte, error)
}

// execRunner executes a command without a shell.
type execRunner struct{}

// Output returns command output after the caller validates arguments.
func (execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// commandJob joins one RPC callback to the single command worker.
type commandJob struct {
	ctx     context.Context
	action  string
	session string
	result  chan commandResult
}

// commandResult carries one completed result or error back to its caller.
type commandResult struct {
	output []byte
	err    error
}

// manager owns token validation and the bounded loginctl worker queue.
// RPC callbacks produce into jobs (capacity 32); the single worker consumes.
// Jobs is never closed because callbacks can still run during NATS drain.
type manager struct {
	verifier  authtoken.Verifier
	client    messaging.Client
	validator *validator.Validate
	runner    commandRunner
	jobs      chan commandJob
	ctx       context.Context
}

// run builds the service dependencies and serves RPCs until cancellation.
func run(ctx context.Context) error {
	return runWithDependencies(ctx, packagedConfigPath, modules.NewCoreClientServerInfra, execRunner{})
}

// runWithDependencies keeps configuration, messaging, and command execution
// explicit for lifecycle tests and alternate host wiring.
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
	infra, err := infraFactory(sharedcontracts.ServiceSessionManager, cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	service := &manager{verifier: verifier, client: infra.Client, validator: validator.New(validator.WithRequiredStructEnabled()), runner: runner, jobs: make(chan commandJob, 32), ctx: ctx}
	if err := service.register(infra.Server); err != nil {
		return err
	}
	done := make(chan struct{})
	workers.NewBaseWorker(func(ctx context.Context) { defer close(done); service.work(ctx) }).Start(ctx)
	infra.Logger.Info("session manager started")
	<-ctx.Done()
	if err := infra.Server.Drain(); err != nil {
		infra.Logger.Error("session manager drain failed", "error", err)
	}
	<-done
	return ctx.Err()
}

// register binds each fixed RPC subject to an authenticated handler.
func (m *manager) register(server messaging.Server) error {
	if err := server.RegisterRPC(contract.ListSubject, m.list); err != nil {
		return err
	}
	for _, operation := range []struct{ subject, action string }{
		{contract.TerminateSubject, "terminate-session"}, {contract.LockSubject, "lock-session"}, {contract.UnlockSubject, "unlock-session"},
	} {
		action := operation.action
		if err := server.RegisterRPC(operation.subject, func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			return m.action(ctx, envelope, action)
		}); err != nil {
			return err
		}
	}
	return nil
}

// decode validates the complete RPC payload and a security or administrator token.
func decode[T any](ctx context.Context, m *manager, envelope messaging.Envelope, accessToken func(T) string) (T, error) {
	var request T
	if len(envelope.Payload) > 16*1024 {
		return request, fmt.Errorf("request too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return request, fmt.Errorf("unexpected trailing JSON")
	}
	if err := m.validator.Struct(request); err != nil {
		return request, err
	}
	claims, err := authtoken.VerifyLive(ctx, m.client, m.verifier, accessToken(request))
	if err != nil || !roleauth.MatchesRequirement(claims.Roles, roleauth.RequirementSecurity) {
		return request, fmt.Errorf("security authorization required")
	}
	return request, nil
}

// list returns loginctl's current session inventory.
func (m *manager) list(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, err := decode(ctx, m, envelope, func(request contract.ListRequest) string { return request.AccessToken }); err != nil {
		return nil, err
	}
	result, err := m.submit(ctx, "list", "")
	if err != nil {
		return nil, err
	}
	var sessions []contract.Session
	if err := json.Unmarshal(result, &sessions); err != nil {
		return nil, fmt.Errorf("decode enriched loginctl sessions: %w", err)
	}
	return contract.ListResponse{Sessions: sessions}, nil
}

// action validates the session ID before requesting one fixed loginctl verb.
func (m *manager) action(ctx context.Context, envelope messaging.Envelope, action string) (any, error) {
	request, err := decode(ctx, m, envelope, func(request contract.ActionRequest) string { return request.AccessToken })
	if err != nil {
		return nil, err
	}
	if !sessionIDPattern.MatchString(request.Session) {
		return nil, fmt.Errorf("invalid session ID")
	}
	if _, err := m.submit(ctx, action, request.Session); err != nil {
		return nil, err
	}
	return contract.ActionResponse{Completed: true}, nil
}

// submit applies backpressure and cancellation to the worker queue. Each RPC
// owns its buffered result channel; the worker sends once and never closes it.
func (m *manager) submit(ctx context.Context, action, session string) ([]byte, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	job := commandJob{ctx: requestCtx, action: action, session: session, result: make(chan commandResult, 1)}
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

// work serializes commands; queued jobs receive cancellation on shutdown.
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

// execute uses a fixed executable, verb, and operand position.
func (m *manager) execute(job commandJob) commandResult {
	if err := job.ctx.Err(); err != nil {
		return commandResult{err: err}
	}
	ctx, cancel := context.WithTimeout(job.ctx, 20*time.Second)
	defer cancel()
	var output []byte
	var err error
	if job.action == "list" {
		output, err = m.runner.Output(ctx, loginctlBinary, "list-sessions", "--no-legend", "--no-pager")
		if err == nil {
			output, err = m.enrichSessions(ctx, output)
		}
	} else {
		output, err = m.runner.Output(ctx, "/usr/bin/sudo", "-n", loginctlBinary, job.action, "--", job.session)
	}
	if err != nil {
		return commandResult{err: fmt.Errorf("loginctl %s failed: %w", job.action, err)}
	}
	if len(output) > 1024*1024 {
		return commandResult{err: fmt.Errorf("loginctl output too large")}
	}
	return commandResult{output: output}
}

// enrichSessions retrieves the identifying logind properties for each listed session.
func (m *manager) enrichSessions(ctx context.Context, inventory []byte) ([]byte, error) {
	sessions, err := parseSessions(inventory)
	if err != nil {
		return nil, err
	}
	for index := range sessions {
		id := sessions[index].ID
		output, commandErr := m.runner.Output(ctx, loginctlBinary, "show-session", id,
			"-p", "Type", "-p", "Class", "-p", "Service", "-p", "Remote", "-p", "RemoteHost", "-p", "TTY", "-p", "Display", "-p", "Desktop", "-p", "Leader")
		if commandErr != nil {
			return nil, fmt.Errorf("loginctl show-session failed: %w", commandErr)
		}
		if len(output) > 64*1024 {
			return nil, fmt.Errorf("loginctl session properties too large")
		}
		if err := applySessionProperties(&sessions[index], output); err != nil {
			return nil, err
		}
	}
	return json.Marshal(sessions)
}

// applySessionProperties maps loginctl's key-value output into the session contract.
func applySessionProperties(session *contract.Session, output []byte) error {
	values := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("invalid loginctl session property")
		}
		values[key] = value
	}
	if values["Remote"] != "yes" && values["Remote"] != "no" {
		return fmt.Errorf("invalid loginctl Remote property")
	}
	session.Type = values["Type"]
	session.Class = values["Class"]
	session.Service = values["Service"]
	session.Remote = values["Remote"] == "yes"
	session.RemoteHost = values["RemoteHost"]
	session.TTY = values["TTY"]
	session.Display = values["Display"]
	session.Desktop = values["Desktop"]
	session.Leader = values["Leader"]
	return nil
}

// parseSessions maps identity and terminal metadata from loginctl list output.
func parseSessions(output []byte) ([]contract.Session, error) {
	sessions := make([]contract.Session, 0)
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 3 || !sessionIDPattern.MatchString(fields[0]) {
			return nil, fmt.Errorf("invalid loginctl session row")
		}
		session := contract.Session{ID: fields[0], UID: fields[1], User: fields[2]}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
