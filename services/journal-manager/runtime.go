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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/config"
	sharedcontracts "lite-nas/shared/contracts"
	contract "lite-nas/shared/contracts/journalmanager"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/eventmanager"
	"lite-nas/shared/fileio"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/roleauth"
	"lite-nas/shared/workers"
)

const (
	packagedConfigPath = "/etc/lite-nas/journal-manager.conf"
	journalctlBinary   = "/usr/bin/journalctl"
	vacuumHelper       = "/usr/libexec/lite-nas/journal-manager"
	maxOutputBytes     = 4 << 20
)

var (
	unitPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.@:-]{0,247}\.service$`)
	cursorPattern = regexp.MustCompile(`^[A-Za-z0-9;=_.:-]+$`)
	alertIDs      = eventmanager.NewManager(uint64(time.Now().Unix()) % 99_999_999)
)

// commandRunner is the injected, bounded host command boundary.
type commandRunner interface {
	Output(context.Context, string, ...string) ([]byte, error)
}

// execRunner executes a fixed binary without a shell and caps its output.
type execRunner struct{}

// Output runs one validated argument vector with bounded standard output.
func (execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &limitedWriter{writer: &output, remaining: maxOutputBytes}
	var stderr bytes.Buffer
	cmd.Stderr = &limitedWriter{writer: &stderr, remaining: 4096}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return output.Bytes(), nil
}

// limitedWriter fails when a command exceeds its output allowance.
type limitedWriter struct {
	writer    io.Writer
	remaining int
}

// Write enforces the output limit before retaining any excess bytes.
func (w *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		return 0, fmt.Errorf("command output too large")
	}
	n, err := w.writer.Write(p)
	w.remaining -= n
	return n, err
}

// commandJob carries one validated journal operation to the worker.
type commandJob struct {
	ctx    context.Context
	query  *contract.QueryRequest
	units  *contract.UnitsRequest
	days   int
	result chan commandResult
}

// commandResult delivers one worker result to its requesting RPC callback.
type commandResult struct {
	output []byte
	err    error
}

// manager owns authentication and a bounded single-consumer command queue.
// RPC callbacks produce jobs (capacity 32); one worker consumes them. The
// channel stays open while callbacks may still send during NATS drain.
type manager struct {
	verifier  authtoken.Verifier
	validator *validator.Validate
	runner    commandRunner
	client    messaging.Client
	jobs      chan commandJob
	ctx       context.Context
}

// run starts the packaged journal service until cancellation.
func run(ctx context.Context) error {
	return runWithDependencies(ctx, packagedConfigPath, modules.NewCoreClientServerInfra, execRunner{})
}

// runWithDependencies keeps configuration, messaging, and command execution explicit for lifecycle tests.
func runWithDependencies(ctx context.Context, configPath string, infraFactory modules.CoreInfraConstructor, runner commandRunner) error {
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	ini, err := config.LoadINI(reader)
	if err != nil {
		return err
	}
	cfg, err := config.LoadSharedAuthTokenConfig(ini)
	if err != nil {
		return err
	}
	verifier, err := loadVerifier(cfg.AuthTokens)
	if err != nil {
		return err
	}
	infra, err := infraFactory(sharedcontracts.ServiceJournalManager, cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	service := &manager{verifier: verifier, validator: validator.New(validator.WithRequiredStructEnabled()), runner: runner, client: infra.Client, jobs: make(chan commandJob, 32), ctx: ctx}
	if err := service.register(infra.Server); err != nil {
		return err
	}
	done := make(chan struct{})
	workers.NewBaseWorker(func(ctx context.Context) { defer close(done); service.work(ctx) }).Start(ctx)
	infra.Logger.Info("journal manager started")
	<-ctx.Done()
	if err := infra.Server.Drain(); err != nil {
		infra.Logger.Error("journal manager drain failed", "error", err)
	}
	<-done
	return ctx.Err()
}

// loadVerifier constructs the access-token verifier from packaged certificate settings.
func loadVerifier(settings config.AuthTokenConfig) (authtoken.Verifier, error) {
	certificate, err := os.ReadFile(settings.VerificationCert)
	if err != nil {
		return authtoken.Verifier{}, err
	}
	key, err := authtoken.ParseEd25519CertificatePublicKeyPEM(certificate)
	if err != nil {
		return authtoken.Verifier{}, err
	}
	return authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: settings.Issuer, Audience: settings.Audience, ClockSkew: settings.ClockSkew}, key)
}

// register binds the fixed RPC subjects to authenticated handlers.
func (m *manager) register(server messaging.Server) error {
	if err := server.RegisterRPC(contract.QuerySubject, m.query); err != nil {
		return err
	}
	if err := server.RegisterRPC(contract.UnitsSubject, m.units); err != nil {
		return err
	}
	return server.RegisterRPC(contract.VacuumSubject, m.vacuum)
}

// decode checks complete JSON, schema, and security authorization before work.
func decode[T any](ctx context.Context, m *manager, envelope messaging.Envelope, token func(T) string) (T, string, error) {
	var request T
	if len(envelope.Payload) > 16*1024 {
		return request, "", fmt.Errorf("request too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, "", err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return request, "", fmt.Errorf("unexpected trailing JSON")
	}
	if err := m.validator.Struct(request); err != nil {
		return request, "", err
	}
	claims, err := authtoken.VerifyLive(ctx, m.client, m.verifier, token(request))
	if err != nil || !roleauth.MatchesRequirement(claims.Roles, roleauth.RequirementSecurity) {
		return request, "", fmt.Errorf("security authorization required")
	}
	return request, claims.Login, nil
}

// query validates filters, then returns one newest-first journal chunk.
func (m *manager) query(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, _, err := decode(ctx, m, envelope, func(v contract.QueryRequest) string { return v.AccessToken })
	if err != nil {
		return nil, err
	}
	if err := validateQuery(request); err != nil {
		return nil, err
	}
	result, err := m.submit(ctx, &request, 0)
	if err != nil {
		return nil, err
	}
	return parseEntries(result, request.Limit)
}

// units lists service units whose journal entries satisfy the non-unit filters.
func (m *manager) units(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, _, err := decode(ctx, m, envelope, func(v contract.UnitsRequest) string { return v.AccessToken })
	if err != nil {
		return nil, err
	}
	if err := validateQuery(contract.QueryRequest{Since: request.Since, Until: request.Until, Message: request.Message}); err != nil {
		return nil, err
	}
	result, err := m.submitJob(ctx, commandJob{units: &request})
	if err != nil {
		return nil, err
	}
	return parseUnits(result, request.Message != "" || request.Since != "" || request.Until != "" || request.Priority != nil)
}

// validateQuery rejects malformed or option-shaped filter values.
func validateQuery(request contract.QueryRequest) error {
	if request.Unit != "" && !unitPattern.MatchString(request.Unit) {
		return fmt.Errorf("invalid unit")
	}
	if request.Cursor != "" && (strings.HasPrefix(request.Cursor, "-") || !cursorPattern.MatchString(request.Cursor)) {
		return fmt.Errorf("invalid cursor")
	}
	if strings.ContainsAny(request.Message, "\x00\r\n") {
		return fmt.Errorf("invalid message search")
	}
	if request.Since != "" && request.Until != "" {
		since, _ := time.Parse(time.RFC3339, request.Since)
		until, _ := time.Parse(time.RFC3339, request.Until)
		if since.After(until) {
			return fmt.Errorf("since must precede until")
		}
	}
	return nil
}

// vacuum rotates active files, applies retention, and emits a security alert.
func (m *manager) vacuum(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, actor, err := decode(ctx, m, envelope, func(v contract.VacuumRequest) string { return v.AccessToken })
	if err != nil {
		return nil, err
	}
	if _, err := m.submit(ctx, nil, request.Days); err != nil {
		return nil, err
	}
	payload := loggingcontract.AlertPayload{AccessToken: request.AccessToken, EventID: alertIDs.NextEventID("secjrn"), Category: "journal_vacuum", Severity: enum.SeverityWarning, Source: sharedcontracts.ServiceJournalManager, Message: fmt.Sprintf("Journal vacuum retained %d days, requested by %s", request.Days, actor), TriggerValue: fmt.Sprintf("days=%d actor=%s", request.Days, actor), Mitigate: "Confirm the journal vacuum was authorized. Review logging policy and available backups if it was unexpected."}
	if err := m.client.Publish(ctx, securitycontract.AlertSubject, payload); err != nil {
		return nil, fmt.Errorf("journal vacuum completed but security alert publication failed: %w", err)
	}
	return contract.VacuumResponse{Completed: true}, nil
}

// submit applies backpressure and cancellation at the worker boundary.
func (m *manager) submit(ctx context.Context, query *contract.QueryRequest, days int) ([]byte, error) {
	return m.submitJob(ctx, commandJob{query: query, days: days})
}

// submitJob sends one validated command to the bounded worker queue.
func (m *manager) submitJob(ctx context.Context, job commandJob) ([]byte, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	job.ctx = requestCtx
	job.result = make(chan commandResult, 1)
	select {
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	case m.jobs <- job:
	}
	return waitForCommand(requestCtx, m.ctx, job.result)
}

// waitForCommand returns the worker result or the first cancellation reason.
func waitForCommand(requestCtx, serviceCtx context.Context, result <-chan commandResult) ([]byte, error) {
	select {
	case value := <-result:
		return value.output, value.err
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	case <-serviceCtx.Done():
		return nil, serviceCtx.Err()
	}
}

// work runs one command at a time and accounts for queued work on shutdown.
func (m *manager) work(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			m.failPending(ctx.Err())
			return
		case job := <-m.jobs:
			if err := ctx.Err(); err != nil {
				job.result <- commandResult{err: err}
				continue
			}
			job.result <- m.execute(job)
		}
	}
}

// failPending accounts for every queued request after producers stop on cancellation.
func (m *manager) failPending(reason error) {
	for {
		select {
		case job := <-m.jobs:
			job.result <- commandResult{err: reason}
		default:
			return
		}
	}
}

// execute uses fixed journalctl commands or the validated vacuum helper.
func (m *manager) execute(job commandJob) commandResult {
	ctx, cancel := context.WithTimeout(job.ctx, 20*time.Second)
	defer cancel()
	if job.query == nil && job.units == nil {
		output, err := m.runner.Output(ctx, "/usr/bin/sudo", "-n", vacuumHelper, "vacuum", strconv.Itoa(job.days))
		return commandResult{output: output, err: err}
	}
	if job.units != nil {
		q := job.units
		if q.Message == "" && q.Since == "" && q.Until == "" && q.Priority == nil {
			output, err := m.runner.Output(ctx, journalctlBinary, "--no-pager", "--quiet", "--field=_SYSTEMD_UNIT")
			return commandResult{output: output, err: err}
		}
		args := []string{"--no-pager", "--quiet", "--output=json", "--output-fields=_SYSTEMD_UNIT"}
		if q.Since != "" {
			args = append(args, "--since="+q.Since)
		}
		if q.Until != "" {
			args = append(args, "--until="+q.Until)
		}
		if q.Priority != nil {
			args = append(args, "--priority="+strconv.Itoa(*q.Priority))
		}
		if q.Message != "" {
			args = append(args, "--grep="+regexp.QuoteMeta(q.Message), "--case-sensitive=no")
		}
		output, err := m.journalOutput(ctx, args, q.Message != "")
		return commandResult{output: output, err: err}
	}
	q := job.query
	args := []string{"--no-pager", "--quiet", "--utc", "--output=json", "--reverse", "--lines=" + strconv.Itoa(q.Limit+1)}
	if q.Since != "" {
		args = append(args, "--since="+q.Since)
	}
	if q.Until != "" {
		args = append(args, "--until="+q.Until)
	}
	if q.Unit != "" {
		args = append(args, "--unit="+q.Unit)
	}
	if q.Priority != nil {
		args = append(args, "--priority="+strconv.Itoa(*q.Priority))
	}
	if q.Message != "" {
		args = append(args, "--grep="+regexp.QuoteMeta(q.Message), "--case-sensitive=no")
	}
	if q.Cursor != "" {
		args = append(args, "--after-cursor="+q.Cursor)
	}
	output, err := m.journalOutput(ctx, args, q.Message != "")
	return commandResult{output: output, err: err}
}

// journalOutput distinguishes no grep matches from journalctl access failures.
func (m *manager) journalOutput(ctx context.Context, args []string, hasGrep bool) ([]byte, error) {
	output, err := m.runner.Output(ctx, journalctlBinary, args...)
	if err == nil || !hasGrep {
		return output, err
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
		return nil, err
	}
	probe := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--grep=") && !strings.HasPrefix(arg, "--case-sensitive=") && !strings.HasPrefix(arg, "--lines=") {
			probe = append(probe, arg)
		}
	}
	probe = append(probe, "--lines=1")
	if _, probeErr := m.runner.Output(ctx, journalctlBinary, probe...); probeErr != nil {
		return nil, err
	}
	return []byte{}, nil
}

// parseUnits returns sorted service units from a journal field list or JSON records.
func parseUnits(output []byte, filtered bool) (contract.UnitsResponse, error) {
	seen := make(map[string]struct{})
	for _, line := range bytes.Split(bytes.TrimSpace(output), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		unit := string(line)
		if filtered {
			var raw struct {
				Unit string `json:"_SYSTEMD_UNIT"`
			}
			if err := json.Unmarshal(line, &raw); err != nil {
				return contract.UnitsResponse{}, fmt.Errorf("decode journal unit: %w", err)
			}
			unit = raw.Unit
		}
		if unitPattern.MatchString(unit) {
			seen[unit] = struct{}{}
		}
	}
	units := make([]string, 0, len(seen))
	for unit := range seen {
		units = append(units, unit)
	}
	sort.Strings(units)
	return contract.UnitsResponse{Units: units}, nil
}

// parseEntries maps journal JSON lines to text fields and a continuation cursor.
func parseEntries(output []byte, limit int) (contract.QueryResponse, error) {
	response := contract.QueryResponse{Entries: make([]contract.Entry, 0, limit)}
	for _, line := range bytes.Split(bytes.TrimSpace(output), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			return response, fmt.Errorf("decode journal entry: %w", err)
		}
		field := func(key string) string { var value string; _ = json.Unmarshal(raw[key], &value); return value }
		entry := contract.Entry{Cursor: field("__CURSOR"), Timestamp: field("__REALTIME_TIMESTAMP"), Message: field("MESSAGE"), Unit: field("_SYSTEMD_UNIT"), Priority: field("PRIORITY"), Identifier: field("SYSLOG_IDENTIFIER"), PID: field("_PID")}
		if entry.Cursor == "" {
			return response, fmt.Errorf("journal entry has no cursor")
		}
		if len(response.Entries) == limit {
			response.HasMore = true
			break
		}
		response.Entries = append(response.Entries, entry)
		response.NextCursor = entry.Cursor
	}
	return response, nil
}
