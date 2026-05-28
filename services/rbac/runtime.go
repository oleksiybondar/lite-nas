package main

import (
	"context"
	"encoding/json"

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
	rpcs := map[string]func(context.Context, messaging.Envelope) (any, error){
		rbaccontract.GetSubjectRolesRPCSubject: func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			var request rbaccontract.GetSubjectRolesRequest
			if !decodeRPCRequest(envelope, &request, requestValidator) {
				return rbaccontract.GetSubjectRolesResponse{}, nil
			}

			uid, groups, ok := service.GetSubjectRoles(ctx, request.Username)
			if !ok {
				return rbaccontract.GetSubjectRolesResponse{}, nil
			}

			return rbaccontract.GetSubjectRolesResponse{UID: uid, Groups: groups}, nil
		},
		rbaccontract.CanReadPathRPCSubject: func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			return handlePathCheckRPC(ctx, envelope, requestValidator, service.CanRead)
		},
		rbaccontract.CanWritePathRPCSubject: func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			return handlePathCheckRPC(ctx, envelope, requestValidator, service.CanWrite)
		},
		rbaccontract.CanExecPathRPCSubject: func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			return handlePathCheckRPC(ctx, envelope, requestValidator, service.CanExec)
		},
		rbaccontract.CanSudoExecRPCSubject: func(ctx context.Context, envelope messaging.Envelope) (any, error) {
			var request rbaccontract.CheckSudoExecRequest
			if !decodeRPCRequest(envelope, &request, requestValidator) {
				return rbaccontract.DecisionResponse{Allowed: false}, nil
			}
			return rbaccontract.DecisionResponse{
				Allowed: service.CanSudoExec(ctx, request.UID, request.Command),
			}, nil
		},
		rbaccontract.InvalidateCacheRPCSubject: func(_ context.Context, envelope messaging.Envelope) (any, error) {
			var request rbaccontract.InvalidateCacheRequest
			if !decodeRPCRequest(envelope, &request, requestValidator) {
				return rbaccontract.InvalidateCacheResponse{OK: false}, nil
			}
			service.InvalidateCache(request.UID)
			return rbaccontract.InvalidateCacheResponse{OK: true}, nil
		},
	}

	for subject, handler := range rpcs {
		if err := server.RegisterRPC(subject, handler); err != nil {
			return err
		}
	}
	return nil
}

func handlePathCheckRPC(
	ctx context.Context,
	envelope messaging.Envelope,
	requestValidator *validator.Validate,
	checkFn func(context.Context, string, string) bool,
) (rbaccontract.DecisionResponse, error) {
	var request rbaccontract.CheckPathRequest
	if !decodeRPCRequest(envelope, &request, requestValidator) {
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
