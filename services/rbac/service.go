package main

import (
	"context"
	"time"

	rbaccache "lite-nas/services/rbac/cache"
	rbacchecks "lite-nas/services/rbac/checks"
)

type decisionService struct {
	cache                 *rbaccache.Store
	runner                rbacchecks.Runner
	realUserTTL           time.Duration
	nonInteractiveUserTTL time.Duration
}

func newDecisionService(
	cache *rbaccache.Store,
	runner rbacchecks.Runner,
	realUserTTL time.Duration,
	nonInteractiveUserTTL time.Duration,
) *decisionService {
	return &decisionService{
		cache:                 cache,
		runner:                runner,
		realUserTTL:           realUserTTL,
		nonInteractiveUserTTL: nonInteractiveUserTTL,
	}
}

func (service *decisionService) GetSubjectRoles(ctx context.Context, username string) (string, []string, bool) {
	identity, err := rbacchecks.ResolveIdentityByUID(ctx, service.runner, username)
	if err != nil {
		return "", nil, false
	}
	return identity.UID, identity.Groups, true
}

func (service *decisionService) CanRead(ctx context.Context, uid string, path string) bool {
	allowed, err := rbacchecks.CanRead(ctx, service.runner, uid, path)
	return err == nil && allowed
}

func (service *decisionService) CanWrite(ctx context.Context, uid string, path string) bool {
	allowed, err := rbacchecks.CanWrite(ctx, service.runner, uid, path)
	return err == nil && allowed
}

func (service *decisionService) CanExec(ctx context.Context, uid string, path string) bool {
	allowed, err := rbacchecks.CanExec(ctx, service.runner, uid, path)
	return err == nil && allowed
}

func (service *decisionService) CanSudoExec(ctx context.Context, uid string, command string) bool {
	if entry, ok := service.cache.Get(uid, command); ok && entry.IsValid(time.Now()) {
		return entry.Allowed
	}

	allowed, err := rbacchecks.CanSudoExec(ctx, service.runner, uid, command)
	if err != nil {
		return false
	}

	cacheTTL := service.nonInteractiveUserTTL
	isInteractive, interactiveErr := rbacchecks.IsInteractiveUser(ctx, service.runner, uid)
	if interactiveErr == nil && isInteractive {
		cacheTTL = service.realUserTTL
	}
	service.cache.Set(uid, command, rbaccache.Entry{
		Allowed:   allowed,
		ExpiresAt: time.Now().Add(cacheTTL),
	})

	return allowed
}

func (service *decisionService) InvalidateCache(uid string) {
	if uid == "" {
		service.cache.InvalidateAll()
		return
	}
	service.cache.InvalidateUID(uid)
}
