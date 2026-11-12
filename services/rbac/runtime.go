package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"

	rbaccache "lite-nas/services/rbac/cache"
	rbacchecks "lite-nas/services/rbac/checks"
	rbacmodules "lite-nas/services/rbac/modules"
	sharedcontracts "lite-nas/shared/contracts"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/messaging"
	sharedworkers "lite-nas/shared/workers"
)

const packagedConfigPath = "/etc/lite-nas/rbac-service.conf"

var (
	newInfraModule      = rbacmodules.NewInfraModule
	newPollingTimerFunc = sharedworkers.NewPollingTimerWorker
)

func run(ctx context.Context) error {
	infra, err := newInfraModule(packagedConfigPath, sharedcontracts.ServiceRBAC)
	if err != nil {
		return err
	}
	defer infra.Close()

	invalidateCh := make(chan struct{}, 1)
	cacheStore := rbaccache.NewStore(invalidateCh)
	service := newDecisionService(
		cacheStore,
		rbacchecks.ExecRunner{},
		infra.Config.Cache.RealUserTTL,
		infra.Config.Cache.NonInteractiveUserTTL,
	)
	requestValidator := validator.New(validator.WithRequiredStructEnabled())

	invalidateTimer, invalidateTicks, err := newPollingTimerFunc(infra.Config.Cache.InvalidateInterval, 1)
	if err != nil {
		return err
	}
	invalidateTimer.Start(ctx)
	go forwardInvalidateTicks(ctx, invalidateTicks, invalidateCh)
	go func() {
		_ = cacheStore.RunInvalidationWorker(ctx)
	}()

	if err = registerRPCHandlers(infra.Server, requestValidator, service); err != nil {
		return err
	}

	infra.Logger.Info("rbac service started", "config", packagedConfigPath)
	<-ctx.Done()
	infra.Logger.Info("rbac service stopping")
	return ctx.Err()
}

func forwardInvalidateTicks(ctx context.Context, ticks <-chan struct{}, invalidateCh chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			select {
			case invalidateCh <- struct{}{}:
			default:
			}
		}
	}
}

func registerRPCHandlers(server messaging.Server, requestValidator *validator.Validate, service *decisionService) error {
	if err := registerRPCEndpoint(server, rbaccontract.GetSubjectRolesRPCSubject, newGetSubjectRolesHandler(requestValidator, service)); err != nil {
		return err
	}
	if err := registerRPCEndpoint(server, rbaccontract.CanReadPathRPCSubject, newPathCheckHandler(requestValidator, service.CanRead)); err != nil {
		return err
	}
	if err := registerRPCEndpoint(server, rbaccontract.CanWritePathRPCSubject, newPathCheckHandler(requestValidator, service.CanWrite)); err != nil {
		return err
	}
	if err := registerRPCEndpoint(server, rbaccontract.CanExecPathRPCSubject, newPathCheckHandler(requestValidator, service.CanExec)); err != nil {
		return err
	}
	if err := registerRPCEndpoint(server, rbaccontract.CheckPathOperationRPCSubject, newPathOperationHandler(requestValidator, service)); err != nil {
		return err
	}
	if err := registerRPCEndpoint(server, rbaccontract.CanSudoExecRPCSubject, newCanSudoExecHandler(requestValidator, service)); err != nil {
		return err
	}
	return registerRPCEndpoint(server, rbaccontract.InvalidateCacheRPCSubject, newInvalidateCacheHandler(requestValidator, service))
}

func registerRPCEndpoint(
	server messaging.Server,
	subject string,
	handler func(context.Context, messaging.Envelope) (any, error),
) error {
	return server.RegisterRPC(subject, handler)
}

func newGetSubjectRolesHandler(
	requestValidator *validator.Validate,
	service *decisionService,
) func(context.Context, messaging.Envelope) (any, error) {
	return func(ctx context.Context, envelope messaging.Envelope) (any, error) {
		var request rbaccontract.GetSubjectRolesRequest
		if !decodeRPCRequest(envelope, &request, requestValidator) ||
			rbacchecks.ValidateIdentityOperand(request.Username) != nil {
			return rbaccontract.GetSubjectRolesResponse{}, nil
		}

		uid, groups, ok := service.GetSubjectRoles(ctx, request.Username)
		if !ok {
			return rbaccontract.GetSubjectRolesResponse{}, nil
		}

		return rbaccontract.GetSubjectRolesResponse{UID: uid, Groups: groups}, nil
	}
}

func newPathCheckHandler(
	requestValidator *validator.Validate,
	checkFn func(context.Context, string, string) bool,
) func(context.Context, messaging.Envelope) (any, error) {
	return func(ctx context.Context, envelope messaging.Envelope) (any, error) {
		return handlePathCheckRPC(ctx, envelope, requestValidator, checkFn)
	}
}

// newPathOperationHandler validates an operation request before invoking the decision service.
func newPathOperationHandler(
	requestValidator *validator.Validate,
	service *decisionService,
) func(context.Context, messaging.Envelope) (any, error) {
	return func(ctx context.Context, envelope messaging.Envelope) (any, error) {
		var request rbaccontract.CheckPathOperationRequest
		if !decodeRPCRequest(envelope, &request, requestValidator) || !validPathRequest(request.UID, request.Path) {
			return rbaccontract.DecisionResponse{Allowed: false}, nil
		}
		return rbaccontract.DecisionResponse{
			Allowed: service.CheckPathOperation(ctx, request.UID, request.Path, string(request.Operation)),
		}, nil
	}
}

func newCanSudoExecHandler(
	requestValidator *validator.Validate,
	service *decisionService,
) func(context.Context, messaging.Envelope) (any, error) {
	return func(ctx context.Context, envelope messaging.Envelope) (any, error) {
		var request rbaccontract.CheckSudoExecRequest
		if !decodeRPCRequest(envelope, &request, requestValidator) ||
			rbacchecks.ValidateSudoArguments(request.UID, request.Command) != nil {
			return rbaccontract.DecisionResponse{Allowed: false}, nil
		}
		return rbaccontract.DecisionResponse{
			Allowed: service.CanSudoExec(ctx, request.UID, request.Command),
		}, nil
	}
}

func newInvalidateCacheHandler(
	requestValidator *validator.Validate,
	service *decisionService,
) func(context.Context, messaging.Envelope) (any, error) {
	return func(_ context.Context, envelope messaging.Envelope) (any, error) {
		var request rbaccontract.InvalidateCacheRequest
		if !decodeRPCRequest(envelope, &request, requestValidator) {
			return rbaccontract.InvalidateCacheResponse{OK: false}, nil
		}
		service.InvalidateCache(request.UID)
		return rbaccontract.InvalidateCacheResponse{OK: true}, nil
	}
}

func handlePathCheckRPC(
	ctx context.Context,
	envelope messaging.Envelope,
	requestValidator *validator.Validate,
	checkFn func(context.Context, string, string) bool,
) (rbaccontract.DecisionResponse, error) {
	var request rbaccontract.CheckPathRequest
	if !decodeRPCRequest(envelope, &request, requestValidator) || !validPathRequest(request.UID, request.Path) {
		return rbaccontract.DecisionResponse{Allowed: false}, nil
	}

	return rbaccontract.DecisionResponse{
		Allowed: checkFn(ctx, request.UID, request.Path),
	}, nil
}

func decodeRPCRequest(envelope messaging.Envelope, request any, requestValidator *validator.Validate) bool {
	if err := json.Unmarshal(envelope.Payload, request); err != nil {
		return false
	}

	return requestValidator.Struct(request) == nil
}

// validPathRequest rejects ambiguous paths and identities before a path decision runs.
func validPathRequest(uid string, path string) bool {
	numericUID, err := strconv.ParseUint(uid, 10, 32)
	return err == nil && numericUID != 0 && filepath.IsAbs(path) &&
		filepath.Clean(path) == path && !strings.ContainsRune(path, 0)
}
