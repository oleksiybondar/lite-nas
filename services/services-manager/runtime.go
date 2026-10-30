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
	"sort"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/config"
	sharedcontracts "lite-nas/shared/contracts"
	contract "lite-nas/shared/contracts/servicesmanager"
	"lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/roleauth"
	"lite-nas/shared/systemd"
	"lite-nas/shared/workers"
)

const packagedConfigPath = "/etc/lite-nas/services-manager.conf"

var serviceUnitPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.@:\\-]*\.service$`)

// validServiceUnit limits operands to local system service names, including escaped unit names.
func validServiceUnit(unit string) bool {
	return len(unit) >= 9 && len(unit) <= 256 && serviceUnitPattern.MatchString(unit)
}

// commandRunner is the injected host command boundary.
type commandRunner interface {
	Output(context.Context, string, ...string) ([]byte, error)
}

// execRunner executes a command without a shell.
type execRunner struct{}

// Output returns bounded command output after the caller validates arguments.
func (execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// commandJob joins one RPC callback to the single command worker.
type commandJob struct {
	ctx    context.Context
	action string
	unit   string
	result chan commandResult
}

// commandResult carries one completed result or error back to its caller.
type commandResult struct {
	output []byte
	err    error
}

// manager owns token validation and the bounded systemctl worker queue.
// RPC callbacks produce into jobs (capacity 32); the single worker consumes.
// Jobs is never closed because callbacks can still run during NATS drain.
type manager struct {
	verifier  authtoken.Verifier
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
	infra, err := infraFactory(sharedcontracts.ServiceServicesManager, cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	service := &manager{verifier: verifier, validator: validator.New(validator.WithRequiredStructEnabled()), runner: runner, jobs: make(chan commandJob, 32), ctx: ctx}
	if err := service.register(infra.Server); err != nil {
		return err
	}
	done := make(chan struct{})
	workers.NewBaseWorker(func(ctx context.Context) { defer close(done); service.work(ctx) }).Start(ctx)
	infra.Logger.Info("services manager started")
	<-ctx.Done()
	if err := infra.Server.Drain(); err != nil {
		infra.Logger.Error("services manager drain failed", "error", err)
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
		{contract.StartSubject, "start"},
		{contract.StopSubject, "stop"},
		{contract.RestartSubject, "restart"},
		{contract.EnableSubject, "enable"},
		{contract.DisableSubject, "disable"},
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

// decode validates the complete RPC payload and an administrator access token.
func decode[T any](m *manager, envelope messaging.Envelope, accessToken func(T) string) (T, error) {
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
	claims, err := m.verifier.Verify(accessToken(request))
	if err != nil || !roleauth.MatchesRequirement(claims.Roles, roleauth.RequirementAdministrator) {
		return request, fmt.Errorf("administrator authorization required")
	}
	return request, nil
}

// list returns the current systemd service inventory.
func (m *manager) list(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, err := decode(m, envelope, func(request contract.ListRequest) string { return request.AccessToken }); err != nil {
		return nil, err
	}
	result, err := m.submit(ctx, "list-units", "")
	if err != nil {
		return nil, err
	}
	services, err := parseServices(result)
	if err != nil {
		return nil, err
	}
	files, err := m.submit(ctx, "list-unit-files", "")
	if err != nil {
		return nil, err
	}
	services, err = mergeServiceFiles(services, files)
	if err != nil {
		return nil, err
	}
	return contract.ListResponse{Services: services}, nil
}

// action validates a service name before requesting one fixed systemctl verb.
func (m *manager) action(ctx context.Context, envelope messaging.Envelope, action string) (any, error) {
	request, err := decode(m, envelope, func(request contract.ActionRequest) string { return request.AccessToken })
	if err != nil {
		return nil, err
	}
	if !validServiceUnit(request.Unit) {
		return nil, fmt.Errorf("invalid service unit")
	}
	if _, err := m.submit(ctx, action, request.Unit); err != nil {
		return nil, err
	}
	return contract.ActionResponse{Completed: true}, nil
}

// submit applies backpressure and cancellation to the worker queue. Each RPC
// owns its buffered result channel; the worker sends once and never closes it.
func (m *manager) submit(ctx context.Context, action, unit string) ([]byte, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	job := commandJob{ctx: requestCtx, action: action, unit: unit, result: make(chan commandResult, 1)}
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
	if job.action == "list-units" {
		output, err = m.runner.Output(ctx, systemd.SystemctlBinary, "--system", "list-units", "--all", "--type=service", "--no-legend", "--plain", "--no-pager")
	} else if job.action == "list-unit-files" {
		output, err = m.runner.Output(ctx, systemd.SystemctlBinary, "--system", "list-unit-files", "--all", "--type=service", "--no-legend", "--plain", "--no-pager")
	} else {
		output, err = m.runner.Output(ctx, "/usr/bin/sudo", "-n", systemd.SystemctlBinary, "--system", job.action, "--", job.unit)
	}
	if err != nil {
		return commandResult{err: fmt.Errorf("systemctl %s failed: %w", job.action, err)}
	}
	if len(output) > 1024*1024 {
		return commandResult{err: fmt.Errorf("systemctl output too large")}
	}
	return commandResult{output: output}
}

// mergeServiceFiles includes disabled installed units and loaded transient units.
func mergeServiceFiles(loaded []contract.Service, output []byte) ([]contract.Service, error) {
	byName := make(map[string]contract.Service, len(loaded))
	for _, service := range loaded {
		byName[service.Name] = service
	}
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 2 || !validServiceUnit(fields[0]) {
			return nil, fmt.Errorf("invalid systemctl unit-file row")
		}
		service := byName[fields[0]]
		service.Name = fields[0]
		service.EnabledState = fields[1]
		if len(fields) > 2 {
			service.Preset = fields[2]
		}
		byName[service.Name] = service
	}
	result := make([]contract.Service, 0, len(byName))
	for _, service := range byName {
		result = append(result, service)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// parseServices accepts only service rows from systemctl's plain list output.
func parseServices(output []byte) ([]contract.Service, error) {
	services := make([]contract.Service, 0)
	for _, line := range strings.Split(string(output), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "●" {
			fields = fields[1:]
		}
		if len(fields) < 4 || !validServiceUnit(fields[0]) {
			return nil, fmt.Errorf("invalid systemctl service row")
		}
		services = append(services, contract.Service{Name: fields[0], LoadState: fields[1], ActiveState: fields[2], SubState: fields[3], Description: strings.Join(fields[4:], " ")})
	}
	return services, nil
}
