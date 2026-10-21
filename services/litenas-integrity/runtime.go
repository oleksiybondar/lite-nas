package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"

	servicemodules "lite-nas/services/litenas-integrity/modules"
	sharedcontracts "lite-nas/shared/contracts"
	authcontract "lite-nas/shared/contracts/auth"
	litenasintegritycontract "lite-nas/shared/contracts/litenasintegrity"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	rbaccontract "lite-nas/shared/contracts/rbac"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	sharedlogger "lite-nas/shared/logger"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"
	sharedworkers "lite-nas/shared/workers"
)

const (
	packagedConfigPath  = "/etc/lite-nas/litenas-integrity.conf"
	serviceName         = sharedcontracts.ServiceLiteNASIntegrity
	eventCategory       = "integrity"
	eventIDPrefix       = "integrity"
	aideAlertMessage    = "AIDE integrity check reported changes"
	aideNormalMessage   = "AIDE integrity state returned to normal"
	aideBaselineMessage = "AIDE baseline regenerated"
	maxEventCounter     = uint64(99_999_999)
	maxAlertMessageLen  = 256
	maxTriggerValueLen  = 512
	sudoBinary          = "/usr/bin/sudo"
	aideBinary          = "/usr/bin/aide"
	moveBinary          = "/bin/mv"
	aideDatabaseNewPath = "/var/lib/aide/aide.db.new"
	aideDatabasePath    = "/var/lib/aide/aide.db"
	aideDatabaseNewGZ   = "/var/lib/aide/aide.db.new.gz"
	aideDatabaseGZ      = "/var/lib/aide/aide.db.gz"
	defaultTimerBufSize = 1
)

var (
	errUnauthorized = errors.New("unauthorized baseline regeneration request")

	newInfraModule      = servicemodules.NewInfraModule
	newPollingTimerFunc = sharedworkers.NewPollingTimerWorker
)

// commandRunner executes external host commands needed by the integrity
// service.
type commandRunner interface {
	CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error)
}

// execCommandRunner invokes host commands through the local operating system.
type execCommandRunner struct{}

// CombinedOutput runs one command and returns combined stdout/stderr output.
func (execCommandRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// activeIncident stores the single in-memory integrity incident tracked by the
// runtime.
type activeIncident struct {
	EventID string
}

// checkStatus identifies one interpreted AIDE check outcome.
type checkStatus int

const (
	checkStatusNormal checkStatus = iota
	checkStatusFailure
	checkStatusExecutionError
)

// checkResult stores one interpreted AIDE check execution.
type checkResult struct {
	Status checkStatus
	Output string
}

// integrityService owns the scheduled AIDE runtime loop, active alert cache,
// and privileged baseline-regeneration behavior.
type integrityService struct {
	client    messaging.Client
	logger    sharedlogger.Logger
	validator *validator.Validate
	runner    commandRunner
	clock     func() time.Time

	mu      sync.Mutex
	active  *activeIncident
	counter uint64
}

// run constructs runtime dependencies, registers the privileged RPC, and
// serves scheduled AIDE checks until shutdown.
func run(ctx context.Context) error {
	return runWithDependencies(ctx, packagedConfigPath, serviceName, newInfraModule, newPollingTimerFunc)
}

// runWithDependencies executes runtime flow with injectable constructors for
// testability.
func runWithDependencies(
	ctx context.Context,
	configPath string,
	appServiceName string,
	infraFactory func(string, string) (servicemodules.Infra, error),
	timerFactory func(time.Duration, int) (sharedworkers.TimerWorker, <-chan struct{}, error),
) error {
	infra, err := infraFactory(configPath, appServiceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	service := newIntegrityService(infra.Client, infra.Logger)
	if err := registerRPCHandlers(infra.Server, service); err != nil {
		return err
	}

	checkTimer, checkTicks, err := timerFactory(infra.Config.AIDE.Schedule, defaultTimerBufSize)
	if err != nil {
		return err
	}

	infra.AuthRefreshTimer.Start(ctx)
	checkTimer.Start(ctx)
	go runAuthRefreshLoop(ctx, infra)

	infra.Logger.Info("litenas integrity service started", "config", configPath)

	for {
		select {
		case <-ctx.Done():
			infra.Logger.Info("litenas integrity service stopping")
			return ctx.Err()
		case <-checkTicks:
			service.handleScheduledCheck(ctx)
		}
	}
}

// newIntegrityService constructs the core runtime state for scheduled checks
// and privileged RPC handling.
func newIntegrityService(client messaging.Client, log sharedlogger.Logger) *integrityService {
	return &integrityService{
		client:    client,
		logger:    log,
		validator: validator.New(validator.WithRequiredStructEnabled()),
		runner:    execCommandRunner{},
		clock:     time.Now,
		counter:   initialEventCounter(time.Now()),
	}
}

// registerRPCHandlers wires the baseline-regeneration RPC onto the messaging
// server.
func registerRPCHandlers(server messaging.Server, service *integrityService) error {
	return server.RegisterRPC(litenasintegritycontract.RegenerateBaselineRPCSubject, service.handleRegenerateBaselineRPC)
}

// runAuthRefreshLoop keeps the background service token fresh for alert
// publishing and logging-manager mutations.
func runAuthRefreshLoop(ctx context.Context, infra servicemodules.Infra) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-infra.AuthRefreshTicks:
			handleAuthRefreshTick(ctx, infra)
		}
	}
}

// handleAuthRefreshTick rotates or reacquires the background service token.
func handleAuthRefreshTick(ctx context.Context, infra servicemodules.Infra) {
	refreshErr := infra.AuthTokenManager.Refresh(ctx)
	if refreshErr == nil {
		infra.Logger.Debug("integrity auth token refreshed")
		return
	}

	if err := infra.AuthTokenManager.Login(ctx); err != nil {
		infra.Logger.Warn("integrity auth token login failed", "error", err)
		return
	}

	infra.Logger.Info("integrity auth token login succeeded")
}

// handleScheduledCheck runs one AIDE check and applies alert lifecycle
// behavior based on the interpreted result.
func (service *integrityService) handleScheduledCheck(ctx context.Context) {
	result := service.runIntegrityCheck(ctx)

	switch result.Status {
	case checkStatusNormal:
		if err := service.normalizeActiveIncident(ctx, aideNormalMessage); err != nil {
			service.logger.Warn("failed to normalize integrity alert", "error", err)
		}
	case checkStatusFailure:
		if err := service.recordIntegrityFailure(ctx, result.Output); err != nil {
			service.logger.Warn("failed to record integrity failure", "error", err)
		}
	case checkStatusExecutionError:
		service.logger.Warn("failed to execute AIDE integrity check", "output", result.Output)
	}
}

// handleRegenerateBaselineRPC validates one privileged caller and regenerates
// the AIDE baseline when authorization succeeds.
func (service *integrityService) handleRegenerateBaselineRPC(ctx context.Context, envelope messaging.Envelope) (any, error) {
	var request litenasintegritycontract.RegenerateBaselineRequest
	if !decodeRPCRequest(envelope, &request, service.validator) {
		return litenasintegritycontract.RegenerateBaselineResponse{OK: false}, nil
	}

	if _, err := service.authorizeBaselineRequest(ctx, request.AccessToken); err != nil {
		service.logger.Warn("baseline regeneration denied", "error", err)
		return litenasintegritycontract.RegenerateBaselineResponse{OK: false}, nil
	}

	if err := service.regenerateBaseline(ctx); err != nil {
		service.logger.Warn("baseline regeneration failed", "error", err)
		return litenasintegritycontract.RegenerateBaselineResponse{OK: false}, nil
	}

	return litenasintegritycontract.RegenerateBaselineResponse{OK: true}, nil
}

// authorizeBaselineRequest validates the submitted human access token and
// resolves current caller roles through RBAC.
func (service *integrityService) authorizeBaselineRequest(ctx context.Context, accessToken string) (string, error) {
	var validateResponse authcontract.ValidateAccessTokenResponse
	if err := service.client.Request(
		ctx,
		authcontract.ValidateAccessTokenRPCSubject,
		authcontract.ValidateAccessTokenRequest{AccessToken: accessToken},
		&validateResponse,
	); err != nil {
		return "", err
	}

	if !validateResponse.Valid || validateResponse.Status != authcontract.StatusAuthenticated || strings.TrimSpace(validateResponse.Username) == "" {
		return "", errUnauthorized
	}

	var rolesResponse rbaccontract.GetSubjectRolesResponse
	if err := service.client.Request(
		ctx,
		rbaccontract.GetSubjectRolesRPCSubject,
		rbaccontract.GetSubjectRolesRequest{Username: validateResponse.Username},
		&rolesResponse,
	); err != nil {
		return "", err
	}

	if !roleauth.HasAnyRole(rolesResponse.Groups, roleauth.AllowedRoles(roleauth.RequirementSecurity)) {
		return "", errUnauthorized
	}

	return validateResponse.Username, nil
}

// regenerateBaseline runs `aide --init`, replaces the active database with the
// generated baseline, and normalizes any active integrity alert.
func (service *integrityService) regenerateBaseline(ctx context.Context) error {
	output, err := service.runAIDECommand(ctx, "--init")
	if err != nil {
		return fmt.Errorf("aide init failed: %w: %s", err, sanitizeCommandOutput(output))
	}

	if err := service.replaceActiveDatabase(ctx); err != nil {
		return err
	}

	if err := service.normalizeActiveIncident(ctx, aideBaselineMessage); err != nil {
		return err
	}

	return nil
}

// runIntegrityCheck executes `aide --check` and classifies the outcome for the
// scheduled runtime loop.
func (service *integrityService) runIntegrityCheck(ctx context.Context) checkResult {
	output, err := service.runAIDECommand(ctx, "--check")
	text := sanitizeCommandOutput(output)
	if err == nil {
		return checkResult{Status: checkStatusNormal, Output: text}
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return checkResult{Status: checkStatusFailure, Output: text}
	}

	return checkResult{Status: checkStatusExecutionError, Output: text}
}

// runAIDECommand invokes AIDE through passwordless sudo under the service's
// dedicated host identity.
func (service *integrityService) runAIDECommand(ctx context.Context, subcommand string) ([]byte, error) {
	return service.runner.CombinedOutput(ctx, sudoBinary, aideBinary, subcommand)
}

// replaceActiveDatabase promotes the generated `.new` AIDE database into the
// active database path using the same privileged execution path as AIDE itself.
func (service *integrityService) replaceActiveDatabase(ctx context.Context) error {
	if _, err := service.runner.CombinedOutput(ctx, sudoBinary, moveBinary, aideDatabaseNewPath, aideDatabasePath); err == nil {
		return nil
	}

	output, gzErr := service.runner.CombinedOutput(ctx, sudoBinary, moveBinary, aideDatabaseNewGZ, aideDatabaseGZ)
	if gzErr == nil {
		return nil
	}

	return fmt.Errorf("failed to replace active AIDE database: %w: %s", gzErr, sanitizeCommandOutput(output))
}

// recordIntegrityFailure creates a new critical alert on the first active
// failure or appends one occurrence to the cached alert on later failures.
func (service *integrityService) recordIntegrityFailure(ctx context.Context, output string) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	if service.active == nil {
		eventID := service.nextEventIDLocked()
		payload := loggingmanagercontract.AlertPayload{
			EventID:      eventID,
			Category:     eventCategory,
			Severity:     sharedloggingenum.SeverityCritical,
			Source:       serviceName,
			Message:      trimText(output, maxAlertMessageLen, aideAlertMessage),
			TriggerValue: trimText(output, maxTriggerValueLen, output),
			CreatedAt:    service.clock().UTC().Format(time.RFC3339),
		}
		if err := service.client.Publish(ctx, securityloggingmanagercontract.AlertSubject, payload); err != nil {
			return err
		}

		service.active = &activeIncident{EventID: eventID}
		return nil
	}

	value := trimText(output, maxTriggerValueLen, output)
	occurrence := loggingmanagercontract.AlertOccurrencePayload{
		EventID:   service.active.EventID,
		Timestamp: service.clock().UTC().Format(time.RFC3339),
		ValueType: sharedloggingenum.ValueTypeText,
		ValueText: stringPointer(value),
	}
	return service.client.Publish(ctx, securityloggingmanagercontract.AlertOccurrenceSubject, occurrence)
}

// normalizeActiveIncident sets the cached integrity alert state to `normal`
// and clears the in-memory event identity when one is active.
func (service *integrityService) normalizeActiveIncident(ctx context.Context, message string) error {
	service.mu.Lock()
	defer service.mu.Unlock()

	if service.active == nil {
		return nil
	}

	var response loggingmanagercontract.OKResponse
	request := loggingmanagercontract.UpdateAlertStateInput{
		EventID: service.active.EventID,
		Status:  sharedloggingenum.StatusNormal,
		Message: stringPointer(trimText(message, maxAlertMessageLen, message)),
	}
	if err := service.client.Request(ctx, securityloggingmanagercontract.UpdateAlertStateRPCSubject, request, &response); err != nil {
		return err
	}
	if !response.OK {
		return errors.New("security logging manager rejected alert normalization")
	}

	service.active = nil
	return nil
}

// nextEventIDLocked increments the local counter and formats the next business
// event ID. The caller must hold `service.mu`.
func (service *integrityService) nextEventIDLocked() string {
	service.counter++
	if service.counter > maxEventCounter {
		service.counter = 1
	}

	return fmt.Sprintf("%s%08d", eventIDPrefix, service.counter)
}

// decodeRPCRequest unmarshals and validates one RPC request payload.
func decodeRPCRequest(envelope messaging.Envelope, request any, requestValidator *validator.Validate) bool {
	if err := json.Unmarshal(envelope.Payload, request); err != nil {
		return false
	}

	return requestValidator.Struct(request) == nil
}

// initialEventCounter derives a non-negative startup seed from current time.
func initialEventCounter(now time.Time) uint64 {
	seconds := now.Unix()
	if seconds <= 0 {
		return 0
	}

	return uint64(seconds) % maxEventCounter
}

// sanitizeCommandOutput trims command output into a stable one-line log or
// alert payload representation.
func sanitizeCommandOutput(output []byte) string {
	return strings.TrimSpace(string(output))
}

// trimText bounds a text field to the repository contract limit while
// preserving a fallback message when the candidate value is empty.
func trimText(text string, limit int, fallback string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		trimmed = fallback
	}
	if len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit]
}

// stringPointer returns a pointer to one string value.
func stringPointer(value string) *string {
	return &value
}
