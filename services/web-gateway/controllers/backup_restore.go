package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	brdto "lite-nas/services/web-gateway/dto/backup_restore"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
	contract "lite-nas/shared/contracts/backuprestore"
)

// BackupRestoreController adapts authenticated browser requests to backup RPC operations.
type BackupRestoreController struct{ service services.BackupRestoreService }

// NewBackupRestoreController constructs the backup and restore controller.
func NewBackupRestoreController(service services.BackupRestoreService) BackupRestoreController {
	return BackupRestoreController{service: service}
}

func (c BackupRestoreController) token(ctx context.Context) (string, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return "", huma.Error401Unauthorized("missing or invalid access token")
	}
	return token, nil
}

func (c BackupRestoreController) Applications(ctx context.Context, _ *brdto.EmptyInput) (*brdto.Output[[]string], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.Applications(ctx, t)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to list backup applications")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) BackedUp(ctx context.Context, _ *brdto.EmptyInput) (*brdto.Output[[]string], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.BackedUp(ctx, t)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to list backed up applications")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) Versions(ctx context.Context, i *brdto.ApplicationInput) (*brdto.Output[[]string], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.Versions(ctx, t, i.Application)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to list backup versions")
	}
	return brdto.NewOutput(v), nil
}

// Metadata returns the descriptive fields embedded in one stored archive.
func (c BackupRestoreController) Metadata(ctx context.Context, i *brdto.VersionInput) (*brdto.Output[contract.BackupMetadataResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.Metadata(ctx, t, i.Application, i.Version)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to read backup metadata")
	}
	return brdto.NewOutput(v), nil
}

// Delete removes one stored archive after the backup service confirms its removal.
func (c BackupRestoreController) Delete(ctx context.Context, i *brdto.VersionInput) (*brdto.Output[contract.DeleteResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	if e := c.service.Delete(ctx, t, i.Application, i.Version); e != nil {
		return nil, huma.Error502BadGateway("failed to delete backup")
	}
	return brdto.NewOutput(contract.DeleteResponse{Deleted: true}), nil
}

func (c BackupRestoreController) Backup(ctx context.Context, i *brdto.BackupInput) (*brdto.Output[map[string]string], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.Backup(ctx, t, i.Application, i.Name, i.Description)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to create backup")
	}
	return brdto.NewOutput(map[string]string{"version": v}), nil
}

func (c BackupRestoreController) BackupAll(ctx context.Context, i *brdto.BackupAllInput) (*brdto.Output[[]contract.BackupAllResult], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.BackupAll(ctx, t, i.Name, i.Description)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to create backups")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) Diff(ctx context.Context, i *brdto.VersionInput) (*brdto.Output[[]contract.Difference], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.Diff(ctx, t, i.Application, i.Version)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to compare backup")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) Restore(ctx context.Context, i *brdto.RestoreInput) (*brdto.Output[contract.RestoreResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.Restore(ctx, t, i.Application, i.Version, i.Mode)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to restore backup")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) MergeStart(ctx context.Context, i *brdto.MergeStartInput) (*brdto.Output[contract.MergeStartResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.MergeStart(ctx, t, i.Application, i.Version)
	if e != nil {
		return nil, huma.Error502BadGateway("failed to start merge session")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) MergeStatus(ctx context.Context, i *brdto.SessionInput) (*brdto.Output[contract.MergeStartResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.MergeStatus(ctx, contract.MergeActionRequest{AccessToken: t, SessionID: i.SessionID})
	if e != nil {
		return nil, huma.Error502BadGateway("failed to read merge session")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) MergeRead(ctx context.Context, i *brdto.MergeReadInput) (*brdto.Output[contract.MergeReadResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.MergeRead(ctx, contract.MergeReadRequest{AccessToken: t, SessionID: i.SessionID, Path: i.Path, Side: i.Side, Offset: i.Offset})
	if e != nil {
		return nil, huma.Error502BadGateway("failed to read merge content")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) MergeWrite(ctx context.Context, i *brdto.MergeWriteInput) (*brdto.Output[contract.MergeWriteResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.MergeWrite(ctx, contract.MergeWriteRequest{AccessToken: t, SessionID: i.SessionID, Path: i.Path, Offset: i.Offset, Data: i.Data, Final: i.Final, SHA256: i.SHA256})
	if e != nil {
		return nil, huma.Error502BadGateway("failed to write merge content")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) MergeApply(ctx context.Context, i *brdto.SessionInput) (*brdto.Output[contract.MergeActionResponse], error) {
	return c.mergeAction(ctx, i, false)
}

func (c BackupRestoreController) MergeDiscard(ctx context.Context, i *brdto.SessionInput) (*brdto.Output[contract.MergeActionResponse], error) {
	return c.mergeAction(ctx, i, true)
}

func (c BackupRestoreController) mergeAction(ctx context.Context, i *brdto.SessionInput, discard bool) (*brdto.Output[contract.MergeActionResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	req := contract.MergeActionRequest{AccessToken: t, SessionID: i.SessionID}
	var v contract.MergeActionResponse
	if discard {
		v, e = c.service.MergeDiscard(ctx, req)
	} else {
		v, e = c.service.MergeApply(ctx, req)
	}
	if e != nil {
		return nil, huma.Error502BadGateway("failed to complete merge session action")
	}
	return brdto.NewOutput(v), nil
}

func (c BackupRestoreController) MergeSave(ctx context.Context, i *brdto.MergeSaveInput) (*brdto.Output[contract.BackupResponse], error) {
	t, e := c.token(ctx)
	if e != nil {
		return nil, e
	}
	v, e := c.service.MergeSave(ctx, contract.MergeSaveRequest{AccessToken: t, SessionID: i.SessionID, Name: i.Name, Description: i.Description})
	if e != nil {
		return nil, huma.Error502BadGateway("failed to save merged backup")
	}
	return brdto.NewOutput(v), nil
}
