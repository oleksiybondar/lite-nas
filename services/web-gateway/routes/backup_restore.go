package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
)

// mountBackupRestoreRouter registers the authenticated backup and restore API.
func mountBackupRestoreRouter(api huma.API, controllers modules.Controllers, authentication middlewares.AuthenticationOptions) {
	group := huma.NewGroup(api, "/backup-restore")
	group.UseSimpleModifier(func(op *huma.Operation) { op.Tags = []string{"backup-restore"} })
	group.UseMiddleware(middlewares.RequireAuthentication(api, authentication))
	group.UseMiddleware(middlewares.RequireAdministrator(api))
	c := controllers.BackupRestore
	huma.Register(group, huma.Operation{OperationID: "backup-restore-applications", Method: http.MethodGet, Path: "/applications", Summary: "List configured backup applications", DefaultStatus: http.StatusOK}, c.Applications)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-backed-up", Method: http.MethodGet, Path: "/backed-up", Summary: "List applications with backups", DefaultStatus: http.StatusOK}, c.BackedUp)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-versions", Method: http.MethodGet, Path: "/applications/{application}/versions", Summary: "List backup versions", DefaultStatus: http.StatusOK}, c.Versions)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-metadata", Method: http.MethodGet, Path: "/applications/{application}/versions/{version}", Summary: "Read backup metadata", DefaultStatus: http.StatusOK}, c.Metadata)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-delete", Method: http.MethodDelete, Path: "/applications/{application}/versions/{version}", Summary: "Delete a backup version", DefaultStatus: http.StatusOK}, c.Delete)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-backup", Method: http.MethodPost, Path: "/applications/{application}/backups", Summary: "Create an application backup", DefaultStatus: http.StatusOK}, c.Backup)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-backup-all", Method: http.MethodPost, Path: "/backups", Summary: "Back up all configured applications", DefaultStatus: http.StatusOK}, c.BackupAll)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-diff", Method: http.MethodGet, Path: "/applications/{application}/versions/{version}/diff", Summary: "Compare a backup with live files", DefaultStatus: http.StatusOK}, c.Diff)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-restore", Method: http.MethodPost, Path: "/applications/{application}/versions/{version}/restore", Summary: "Restore an application backup", DefaultStatus: http.StatusOK}, c.Restore)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-start", Method: http.MethodPost, Path: "/applications/{application}/versions/{version}/merge", Summary: "Start a merge session", DefaultStatus: http.StatusOK}, c.MergeStart)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-status", Method: http.MethodGet, Path: "/merge-sessions/{session_id}", Summary: "Read a merge session", DefaultStatus: http.StatusOK}, c.MergeStatus)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-read", Method: http.MethodGet, Path: "/merge-sessions/{session_id}/content", Summary: "Read merge content", DefaultStatus: http.StatusOK}, c.MergeRead)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-write", Method: http.MethodPut, Path: "/merge-sessions/{session_id}/content", Summary: "Write resolved merge content", DefaultStatus: http.StatusOK}, c.MergeWrite)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-apply", Method: http.MethodPost, Path: "/merge-sessions/{session_id}/apply", Summary: "Apply a merge session", DefaultStatus: http.StatusOK}, c.MergeApply)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-save", Method: http.MethodPost, Path: "/merge-sessions/{session_id}/save", Summary: "Save merged content as a backup", DefaultStatus: http.StatusOK}, c.MergeSave)
	huma.Register(group, huma.Operation{OperationID: "backup-restore-merge-discard", Method: http.MethodDelete, Path: "/merge-sessions/{session_id}", Summary: "Discard a merge session", DefaultStatus: http.StatusOK}, c.MergeDiscard)
}
