package main

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	servicemodules "lite-nas/services/litenas-integrity/modules"
	authcontract "lite-nas/shared/contracts/auth"
	litenasintegritycontract "lite-nas/shared/contracts/litenasintegrity"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	rbaccontract "lite-nas/shared/contracts/rbac"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"
	"lite-nas/shared/servicetoken"
)

// integrityRunnerResult describes one expected command and its observed result.
type integrityRunnerResult struct {
	name   string
	args   []string
	output string
	err    error
}

// integrityRunner records the host commands issued by a service test.
type integrityRunner struct {
	results []integrityRunnerResult
	index   int
}

// CombinedOutput checks one command invocation and returns its scripted result.
func (runner *integrityRunner) CombinedOutput(_ context.Context, name string, args ...string) ([]byte, error) {
	if runner.index >= len(runner.results) {
		return nil, errors.New("unexpected host command")
	}
	result := runner.results[runner.index]
	runner.index++
	if name != result.name || strings.Join(args, "\x00") != strings.Join(result.args, "\x00") {
		return nil, errors.New("wrong host command")
	}
	return []byte(result.output), result.err
}

// integrityClient supplies the requested outbound messaging outcomes.
type integrityClient struct {
	request func(string, any, any) error
	publish func(string, any) error
}

// Request dispatches one scripted RPC response.
func (client integrityClient) Request(_ context.Context, subject string, request any, response any) error {
	if client.request == nil {
		return errors.New("unexpected RPC")
	}
	return client.request(subject, request, response)
}

// Publish dispatches one scripted event publication.
func (client integrityClient) Publish(_ context.Context, subject string, payload any) error {
	if client.publish == nil {
		return errors.New("unexpected publish")
	}
	return client.publish(subject, payload)
}

// Drain satisfies the messaging client lifecycle in tests.
func (integrityClient) Drain() error { return nil }

// Close satisfies the messaging client lifecycle in tests.
func (integrityClient) Close() {}

// newTestIntegrityService builds the command and messaging boundaries for one test.
func newTestIntegrityService(runner *integrityRunner, client integrityClient) *integrityService {
	return &integrityService{
		client: client, logger: sharedlogger.NewNop(), validator: validator.New(),
		runner: runner, clock: func() time.Time { return time.Unix(1700000000, 0) },
	}
}

// mustAuthRefreshInfra builds a token manager for one refresh-loop test.
func mustAuthRefreshInfra(t *testing.T, client integrityClient) (servicemodules.Infra, *servicetoken.Manager) {
	t.Helper()
	manager, err := servicetoken.NewManager(client, servicetoken.Options{Service: "integrity"})
	if err != nil {
		t.Fatal(err)
	}
	infra := servicemodules.Infra{}
	infra.AuthTokenManager = manager
	infra.Logger = sharedlogger.NewNop()
	return infra, manager
}

// aideResult supplies one expected AIDE operation to the scripted runner.
func aideResult(option string, output string, err error) integrityRunnerResult {
	return integrityRunnerResult{name: sudoBinary, args: []string{aideBinary, option}, output: output, err: err}
}

// moveResult supplies one expected database promotion to the scripted runner.
func moveResult(compressed bool, err error) integrityRunnerResult {
	if compressed {
		return integrityRunnerResult{name: sudoBinary, args: []string{moveBinary, aideDatabaseNewGZ, aideDatabaseGZ}, err: err}
	}
	return integrityRunnerResult{name: sudoBinary, args: []string{moveBinary, aideDatabaseNewPath, aideDatabasePath}, err: err}
}

// TestRunIntegrityCheckClassifiesCommandResults covers normal, drift, and execution errors.
func TestRunIntegrityCheckClassifiesCommandResults(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want checkStatus
	}{
		{name: "normal", want: checkStatusNormal},
		{name: "changes", err: &exec.ExitError{}, want: checkStatusFailure},
		{name: "failed invocation", err: errors.New("unavailable"), want: checkStatusExecutionError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--check", " drift \n", test.err)}}
			result := newTestIntegrityService(runner, integrityClient{}).runIntegrityCheck(t.Context())
			if result.Status != test.want {
				t.Fatalf("runIntegrityCheck() status = %v, want %v", result.Status, test.want)
			}
		})
	}
}

// TestReplaceActiveDatabaseTriesCompressedFallback covers both database names.
func TestReplaceActiveDatabaseTriesCompressedFallback(t *testing.T) {
	for _, secondErr := range []error{nil, errors.New("compressed missing")} {
		runner := &integrityRunner{results: []integrityRunnerResult{
			moveResult(false, errors.New("uncompressed missing")), moveResult(true, secondErr),
		}}
		err := newTestIntegrityService(runner, integrityClient{}).replaceActiveDatabase(t.Context())
		if (err == nil) != (secondErr == nil) {
			t.Fatalf("replaceActiveDatabase() error = %v", err)
		}
	}
}

// TestRegenerateBaselineRunsInitAndPromotion covers a completed baseline refresh.
func TestRegenerateBaselineRunsInitAndPromotion(t *testing.T) {
	runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--init", "", nil), moveResult(false, nil)}}
	service := newTestIntegrityService(runner, integrityClient{})
	if err := service.regenerateBaseline(t.Context()); err != nil {
		t.Fatalf("regenerateBaseline() = %v", err)
	}
}

// TestRegenerateBaselineStopsAfterInitFailure covers command failure propagation.
func TestRegenerateBaselineStopsAfterInitFailure(t *testing.T) {
	runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--init", "denied", errors.New("failed"))}}
	if err := newTestIntegrityService(runner, integrityClient{}).regenerateBaseline(t.Context()); err == nil {
		t.Fatal("regenerateBaseline() accepted failed init")
	}
}

// TestRegenerateBaselineReportsPromotionFailure covers failed database replacement.
func TestRegenerateBaselineReportsPromotionFailure(t *testing.T) {
	runner := &integrityRunner{results: []integrityRunnerResult{
		aideResult("--init", "", nil),
		moveResult(false, errors.New("missing")),
		moveResult(true, errors.New("missing")),
	}}
	if err := newTestIntegrityService(runner, integrityClient{}).regenerateBaseline(t.Context()); err == nil {
		t.Fatal("regenerateBaseline() accepted failed promotion")
	}
}

// TestRecordIntegrityFailureOpensIncident covers initial alert state.
func TestRecordIntegrityFailureOpensIncident(t *testing.T) {
	var subjects []string
	client := integrityClient{publish: func(subject string, _ any) error {
		subjects = append(subjects, subject)
		return nil
	}}
	service := newTestIntegrityService(&integrityRunner{}, client)
	if err := service.recordIntegrityFailure(t.Context(), "change"); err != nil {
		t.Fatal(err)
	}
	if service.active == nil {
		t.Fatal("recordIntegrityFailure() left no active incident")
	}
}

// TestRecordIntegrityFailurePublishesOccurrence covers repeated alert delivery.
func TestRecordIntegrityFailurePublishesOccurrence(t *testing.T) {
	var subjects []string
	client := integrityClient{publish: func(subject string, _ any) error { subjects = append(subjects, subject); return nil }}
	service := newTestIntegrityService(&integrityRunner{}, client)
	if err := service.recordIntegrityFailure(t.Context(), "change"); err != nil {
		t.Fatal(err)
	}
	if err := service.recordIntegrityFailure(t.Context(), "again"); err != nil {
		t.Fatal(err)
	}
	if len(subjects) != 2 {
		t.Fatalf("published subjects = %v", subjects)
	}
}

// TestRecordIntegrityFailureKeepsStateAfterPublishError covers alert delivery failure.
func TestRecordIntegrityFailureKeepsStateAfterPublishError(t *testing.T) {
	client := integrityClient{publish: func(string, any) error { return errors.New("offline") }}
	service := newTestIntegrityService(&integrityRunner{}, client)
	_ = service.recordIntegrityFailure(t.Context(), "change")
	if service.active != nil {
		t.Fatalf("recordIntegrityFailure() active = %+v", service.active)
	}
}

// TestNormalizeActiveIncidentRequiresAcknowledgment covers state clearing after RPC success.
func TestNormalizeActiveIncidentRequiresAcknowledgment(t *testing.T) {
	for _, accepted := range []bool{false, true} {
		client := integrityClient{request: func(_ string, _ any, response any) error {
			response.(*loggingmanagercontract.OKResponse).OK = accepted
			return nil
		}}
		service := newTestIntegrityService(&integrityRunner{}, client)
		service.active = &activeIncident{EventID: "integrity00000001"}
		_ = service.normalizeActiveIncident(t.Context(), "normal")
		if (service.active == nil) != accepted {
			t.Fatalf("normalizeActiveIncident() active = %+v", service.active)
		}
	}
}

// mustAuthorizeBaselineRequest verifies fixture authorization and returns the caller.
func mustAuthorizeBaselineRequest(t *testing.T) string {
	t.Helper()
	username, err := newTestIntegrityService(&integrityRunner{}, authorizedIntegrityClient()).authorizeBaselineRequest(t.Context(), "token")
	if err != nil {
		t.Fatal(err)
	}
	return username
}

// TestAuthorizeBaselineRequestChecksTokenAndRole covers the two authorization RPCs.
func TestAuthorizeBaselineRequestChecksTokenAndRole(t *testing.T) {
	username := mustAuthorizeBaselineRequest(t)
	if username != "alice" {
		t.Fatalf("authorizeBaselineRequest() = %q", username)
	}
}

// TestAuthorizeBaselineRequestRejectsInvalidToken covers early denial.
func TestAuthorizeBaselineRequestRejectsInvalidToken(t *testing.T) {
	client := integrityClient{request: func(_ string, _ any, _ any) error { return nil }}
	if _, err := newTestIntegrityService(&integrityRunner{}, client).authorizeBaselineRequest(t.Context(), "token"); !errors.Is(err, errUnauthorized) {
		t.Fatalf("authorizeBaselineRequest() error = %v", err)
	}
}

// TestAuthorizeBaselineRequestRejectsInsufficientRole covers RBAC denial after valid Auth.
func TestAuthorizeBaselineRequestRejectsInsufficientRole(t *testing.T) {
	client := integrityClient{request: func(subject string, _ any, response any) error {
		if subject == authcontract.ValidateAccessTokenRPCSubject {
			*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{
				Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice",
			}
		}
		return nil
	}}
	if _, err := newTestIntegrityService(&integrityRunner{}, client).authorizeBaselineRequest(t.Context(), "token"); !errors.Is(err, errUnauthorized) {
		t.Fatalf("authorizeBaselineRequest() error = %v", err)
	}
}

// TestAuthorizeBaselineRequestReportsAuthFailure covers failed token verification RPC.
func TestAuthorizeBaselineRequestReportsAuthFailure(t *testing.T) {
	want := errors.New("auth unavailable")
	client := integrityClient{request: func(string, any, any) error { return want }}
	if _, err := newTestIntegrityService(&integrityRunner{}, client).authorizeBaselineRequest(t.Context(), "token"); !errors.Is(err, want) {
		t.Fatalf("authorizeBaselineRequest() error = %v", err)
	}
}

// mustDecodeBaselineRPCResponse invokes the validation boundary for one payload.
func mustDecodeBaselineRPCResponse(t *testing.T, service *integrityService, payload []byte) litenasintegritycontract.RegenerateBaselineResponse {
	t.Helper()
	response, err := service.handleRegenerateBaselineRPC(t.Context(), messaging.Envelope{Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	return response.(litenasintegritycontract.RegenerateBaselineResponse)
}

// TestRegenerateBaselineRPCRejectsMalformedInput covers validation before host commands.
func TestRegenerateBaselineRPCRejectsMalformedInput(t *testing.T) {
	service := newTestIntegrityService(&integrityRunner{}, integrityClient{})
	for _, payload := range [][]byte{[]byte("{"), []byte(`{"access_token":""}`)} {
		response := mustDecodeBaselineRPCResponse(t, service, payload)
		if response.OK {
			t.Fatalf("handleRegenerateBaselineRPC() = %+v", response)
		}
	}
}

// mustInvokeBaselineRPC sends a valid authorized request through the RPC boundary.
func mustInvokeBaselineRPC(t *testing.T, runner *integrityRunner) litenasintegritycontract.RegenerateBaselineResponse {
	t.Helper()
	service := newTestIntegrityService(runner, authorizedIntegrityClient())
	response, err := service.handleRegenerateBaselineRPC(t.Context(), messaging.Envelope{
		Payload: []byte(`{"access_token":"token"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	return response.(litenasintegritycontract.RegenerateBaselineResponse)
}

// TestRegenerateBaselineRPCAuthorizesAndRunsFixedCommands covers the privileged flow.
func TestRegenerateBaselineRPCAuthorizesAndRunsFixedCommands(t *testing.T) {
	runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--init", "", nil), moveResult(false, nil)}}
	response := mustInvokeBaselineRPC(t, runner)
	if !response.OK {
		t.Fatalf("handleRegenerateBaselineRPC() = %+v", response)
	}
}

// TestRegenerateBaselineRPCFailsClosedAfterCommandError covers a failed AIDE init.
func TestRegenerateBaselineRPCFailsClosedAfterCommandError(t *testing.T) {
	runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--init", "denied", errors.New("failed"))}}
	response := mustInvokeBaselineRPC(t, runner)
	if response.OK {
		t.Fatalf("handleRegenerateBaselineRPC() = %+v", response)
	}
}

// TestScheduledCheckAppliesEachCommandOutcome covers normal, drift, and invocation failure.
func TestScheduledCheckAppliesEachCommandOutcome(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		active bool
	}{
		{name: "normal"},
		{name: "drift", err: &exec.ExitError{}, active: true},
		{name: "execution error", err: errors.New("offline")},
	} {
		t.Run(test.name, func(t *testing.T) {
			runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--check", "change", test.err)}}
			client := integrityClient{publish: func(string, any) error { return nil }}
			service := newTestIntegrityService(runner, client)
			service.handleScheduledCheck(t.Context())
			if (service.active != nil) != test.active {
				t.Fatalf("handleScheduledCheck() active = %+v", service.active)
			}
		})
	}
}

// authorizedIntegrityClient returns successful Auth and RBAC responses for one human.
func authorizedIntegrityClient() integrityClient {
	return integrityClient{request: func(subject string, _ any, response any) error {
		switch subject {
		case authcontract.ValidateAccessTokenRPCSubject:
			*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{
				Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice",
			}
		case rbaccontract.GetSubjectRolesRPCSubject:
			response.(*rbaccontract.GetSubjectRolesResponse).Groups = []string{roleauth.RoleSecurity}
		default:
			return errors.New("unexpected RPC")
		}
		return nil
	}}
}

// TestNextEventIDWrapsCounter covers stable alert identifiers.
func TestNextEventIDWrapsCounter(t *testing.T) {
	service := &integrityService{counter: maxEventCounter}
	if id := service.nextEventIDLocked(); id != "integrity00000001" {
		t.Fatalf("nextEventIDLocked() = %q", id)
	}
}

// TestInitialEventCounterClampsNegativeTime covers event counter seeds.
func TestInitialEventCounterClampsNegativeTime(t *testing.T) {
	if got := initialEventCounter(time.Unix(-1, 0)); got != 0 {
		t.Fatalf("counter = %d", got)
	}
}

// TestInitialEventCounterUsesUnixTime covers event counter seeds.
func TestInitialEventCounterUsesUnixTime(t *testing.T) {
	if got := initialEventCounter(time.Unix(100, 0)); got != 100 {
		t.Fatalf("counter = %d", got)
	}
}

// TestTrimTextUsesFallback covers blank alert output.
func TestTrimTextUsesFallback(t *testing.T) {
	if got := trimText(" ", 4, "base"); got != "base" {
		t.Fatalf("text = %q", got)
	}
}

// TestTrimTextBoundsOutput covers long alert output.
func TestTrimTextBoundsOutput(t *testing.T) {
	if got := trimText("abcdef", 4, "base"); got != "abcd" {
		t.Fatalf("text = %q", got)
	}
}

// TestSanitizeCommandOutputTrimsWhitespace covers command output normalization.
func TestSanitizeCommandOutputTrimsWhitespace(t *testing.T) {
	if got := sanitizeCommandOutput([]byte(" line \n")); got != "line" {
		t.Fatalf("output = %q", got)
	}
}

// TestStringPointerPreservesText covers optional text values.
func TestStringPointerPreservesText(t *testing.T) {
	if got := *stringPointer("text"); got != "text" {
		t.Fatalf("text = %q", got)
	}
}

// TestDecodeRPCRequestAcceptsValidToken covers RPC validation.
func TestDecodeRPCRequestAcceptsValidToken(t *testing.T) {
	var request litenasintegritycontract.RegenerateBaselineRequest
	if decodeRPCRequest(messaging.Envelope{Payload: json.RawMessage(`{"access_token":"yes"}`)}, &request, validator.New()) != true {
		t.Fatal("valid RPC request rejected")
	}
}

// TestNewIntegrityServiceWiresRuntimeDependencies covers runtime construction.
func TestNewIntegrityServiceWiresRuntimeDependencies(t *testing.T) {
	service := newIntegrityService(integrityClient{}, sharedlogger.NewNop())
	for _, test := range []struct {
		name      string
		available bool
	}{
		{name: "validator", available: service.validator != nil},
		{name: "runner", available: service.runner != nil},
		{name: "clock", available: service.clock != nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			if !test.available {
				t.Fatalf("newIntegrityService() omitted %s", test.name)
			}
		})
	}
}

// TestRunWithDependenciesReturnsInfraFailure covers startup before command execution.
func TestRunWithDependenciesReturnsInfraFailure(t *testing.T) {
	want := errors.New("infra unavailable")
	err := runWithDependencies(t.Context(), "/tmp/config", "integrity", func(string, string) (servicemodules.Infra, error) {
		return servicemodules.Infra{}, want
	}, nil)
	if !errors.Is(err, want) {
		t.Fatalf("runWithDependencies() error = %v", err)
	}
}

// TestRunAuthRefreshLoopStopsOnCancellation covers the background loop lifetime.
func TestRunAuthRefreshLoopStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	runAuthRefreshLoop(ctx, servicemodules.Infra{})
}

// TestHandleAuthRefreshTickLogsInWhenRefreshUnavailable covers token reacquisition.
func TestHandleAuthRefreshTickLogsInWhenRefreshUnavailable(t *testing.T) {
	client := integrityClient{request: func(subject string, _ any, response any) error {
		if subject != authcontract.ServiceTokenLoginRPCSubject {
			return errors.New("unexpected auth subject")
		}
		response.(*authcontract.ServiceTokenLoginResponse).AccessToken = "new token"
		return nil
	}}
	infra, manager := mustAuthRefreshInfra(t, client)
	handleAuthRefreshTick(t.Context(), infra)
	if token, _, err := manager.Token(); err != nil {
		t.Fatal(err)
	} else if token != "new token" {
		t.Fatalf("handleAuthRefreshTick() token = %q", token)
	}
}

// TestHandleAuthRefreshTickHandlesLoginFailure covers unavailable Auth service.
func TestHandleAuthRefreshTickHandlesLoginFailure(t *testing.T) {
	client := integrityClient{request: func(string, any, any) error { return errors.New("auth offline") }}
	infra, manager := mustAuthRefreshInfra(t, client)
	handleAuthRefreshTick(t.Context(), infra)
	if _, _, err := manager.Token(); !errors.Is(err, servicetoken.ErrTokenUnavailable) {
		t.Fatalf("handleAuthRefreshTick() token error = %v", err)
	}
}
