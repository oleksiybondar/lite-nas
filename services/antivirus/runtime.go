package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"

	backgroundtasks "lite-nas/shared/backgroundtasks"
	sharedcontracts "lite-nas/shared/contracts"
	antiviruscontract "lite-nas/shared/contracts/antivirus"
	authcontract "lite-nas/shared/contracts/auth"
	backgroundtaskscontract "lite-nas/shared/contracts/backgroundtasks"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	mountcontract "lite-nas/shared/contracts/mountmonitor"
	rbaccontract "lite-nas/shared/contracts/rbac"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/eventmanager"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanagerauth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/mountinfo"
	"lite-nas/shared/roleauth"
	"lite-nas/shared/workers"
)

const (
	packagedAntivirusConfig = "/etc/lite-nas/antivirus.conf"
	maxMountEventBytes      = 16 * 1024
	maxScanQueue            = 64
	maxScanEffectQueue      = 64
	maxSignatureUpdateQueue = 1
	signatureUpdateTimeout  = 10 * time.Minute
	sudoBinary              = "/usr/bin/sudo"
	systemctlBinary         = "/usr/bin/systemctl"
	updateUnitName          = "lite-nas-antivirus-updater.service"
)

// scanJob carries an independently cancelable scan for one verified mount.
type scanJob struct {
	mount mountinfo.Record
	ctx   context.Context
}

// run composes the validated NATS bridge, coordinator, scanner and outbound
// reporting worker within one service process.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedAntivirusConfig)
	if err != nil {
		return err
	}
	cfg, err := loadAntivirusConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceAntivirus, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	server, err := messaging.NewServer(cfg.Messaging, infra.Logger, messaging.NewJSONCodec())
	if err != nil {
		return err
	}
	defer server.Close()
	tasks := backgroundtasks.NewManager(ctx)
	defer tasks.Close()
	workers.NewBaseWorker(tasks.Run).Start(ctx)
	<-tasks.Ready()
	operatorClient := loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken)
	// Coordinator and the task-result dispatcher produce effects; reportScans consumes them.
	// Runtime closes this bounded fan-in queue after both producers have stopped.
	effects := make(chan scanResult, maxScanEffectQueue)
	updates := make(chan struct{}, maxSignatureUpdateQueue)
	taskService := &antivirusTaskService{client: infra.Client, logger: infra.Logger, manager: tasks, cfg: cfg, validator: validator.New(validator.WithRequiredStructEnabled()), effects: effects, targets: make(map[string]scanTarget), updates: updates}
	updateDone := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(updateDone)
		runSignatureUpdateWorker(workerCtx, updates, taskService)
	}).Start(ctx)
	taskResultsDone := make(chan struct{})
	workers.NewBaseWorker(func(context.Context) {
		defer close(taskResultsDone)
		forwardScanTaskResults(tasks.Results(), taskService)
	}).Start(ctx)
	if err := registerTaskRPCs(server, taskService); err != nil {
		return err
	}
	infra.AuthRefreshTimer.Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		runAuthRefreshLoop(workerCtx, infra)
	}).Start(ctx)
	// NATS callbacks produce incoming lifecycle events; the coordinator consumes.
	// Runtime closes this queue only after draining the subscription.
	incoming := make(chan mountcontract.LifecycleEvent, maxScanQueue)
	// The coordinator produces scan jobs and closes jobs after its last schedule.
	// The scanner worker consumes the bounded queue.
	jobs := make(chan scanJob, maxScanQueue)
	// The scanner worker produces outcomes and owns closure; the coordinator consumes.
	results := make(chan scanResult, maxScanQueue)
	validator := validator.New(validator.WithRequiredStructEnabled())
	if err := server.Subscribe(mountcontract.LifecycleSubject, func(handlerCtx context.Context, envelope messaging.Envelope) error {
		event, err := decodeMountEvent(envelope, validator)
		if err != nil {
			infra.Logger.Warn("invalid mount notification", "error", err)
			return err
		}
		select {
		case incoming <- event:
			return nil
		case <-handlerCtx.Done():
			return handlerCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return err
	}
	coordinatorDone := make(chan struct{})
	reporterDone := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) { runScanWorker(workerCtx, jobs, results, cfg.ScanTimeout) }).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		coordinateScans(workerCtx, cfg, incoming, jobs, results, effects, infra.Logger)
		close(coordinatorDone)
	}).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		reportScans(workerCtx, effects, operatorClient, infra.Logger)
		close(reporterDone)
	}).Start(ctx)
	<-ctx.Done()
	if err := server.Drain(); err != nil {
		infra.Logger.Error("NATS subscription drain failed", "error", err)
	}
	close(updates) // Runtime owns closure after the drained RPC server can no longer enqueue updates.
	<-updateDone
	tasks.Close()
	close(incoming) // The drained NATS server can no longer send.
	<-coordinatorDone
	tasks.Wait()
	<-taskResultsDone
	close(effects) // Runtime owns closure after the coordinator and task-result dispatcher finish.
	<-reporterDone
	return ctx.Err()
}

// runAuthRefreshLoop keeps the antivirus service token valid for alert publishing.
func runAuthRefreshLoop(ctx context.Context, infra modules.CoreClientAuthInfra) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-infra.AuthRefreshTicks:
			handleAuthRefreshTick(ctx, infra)
		}
	}
}

// handleAuthRefreshTick rotates an existing token or acquires the initial service token.
func handleAuthRefreshTick(ctx context.Context, infra modules.CoreClientAuthInfra) {
	if err := infra.AuthTokenManager.Refresh(ctx); err == nil {
		infra.Logger.Debug("antivirus auth token refreshed")
		return
	}
	if err := infra.AuthTokenManager.Login(ctx); err != nil {
		infra.Logger.Warn("antivirus auth token login failed", "error", err)
		return
	}
	infra.Logger.Info("antivirus auth token login succeeded")
}

// antivirusTaskService owns authorized on-demand scan RPCs for this process.
type antivirusTaskService struct {
	client        messaging.Client
	logger        sharedlogger.Logger
	manager       *backgroundtasks.Manager
	cfg           antivirusConfig
	validator     *validator.Validate
	effects       chan<- scanResult
	mu            sync.Mutex
	targets       map[string]scanTarget
	updateState   string
	updateDetails string
	updates       chan<- struct{}
}

// scanTarget retains the verified folder identity needed to report one task result.
type scanTarget struct {
	mount  mountinfo.Record
	folder string
}

// registerTaskRPCs binds the distinct antivirus task subjects to their handlers.
func registerTaskRPCs(server messaging.Server, service *antivirusTaskService) error {
	for subject, handler := range map[string]messaging.RPCHandler{
		antiviruscontract.StartScanSubject:        service.start,
		antiviruscontract.RetryScanSubject:        service.retry,
		antiviruscontract.ScanStatusSubject:       service.status,
		antiviruscontract.InfoSubject:             service.info,
		antiviruscontract.UpdateSignaturesSubject: service.updateSignatures,
	} {
		if err := server.RegisterRPC(subject, handler); err != nil {
			return err
		}
	}
	return nil
}

// info returns signature database freshness and the ClamAV engine version to an authorized caller.
func (service *antivirusTaskService) info(ctx context.Context, envelope messaging.Envelope) (any, error) {
	var request antiviruscontract.InfoRequest
	if err := decodeRPCRequestError(envelope, &request, service.validator); err != nil {
		return nil, errors.New("invalid antivirus info request")
	}
	if err := service.authorize(ctx, request.AccessToken); err != nil {
		return antiviruscontract.InfoResponse{}, err
	}
	response := antiviruscontract.InfoResponse{State: "missing_signatures", UpdateState: "idle"}
	service.mu.Lock()
	if service.updateState != "" {
		response.UpdateState = service.updateState
	}
	response.UpdateDetails = service.updateDetails
	service.mu.Unlock()
	entries, err := os.ReadDir("/var/lib/clamav")
	if err != nil {
		response.Details = err.Error()
	} else {
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() || !(strings.HasSuffix(name, ".cvd") || strings.HasSuffix(name, ".cld")) {
				continue
			}
			info, statErr := entry.Info()
			if statErr != nil {
				continue
			}
			response.SignatureDatabaseCount++
			updated := info.ModTime().UTC()
			if response.SignatureUpdatedAt == nil || updated.After(*response.SignatureUpdatedAt) {
				response.SignatureUpdatedAt = &updated
			}
		}
		if response.SignatureDatabaseCount > 0 {
			response.State = "ready"
		}
	}
	versionCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	output, versionErr := exec.CommandContext(versionCtx, clamscanBinary, "--version").CombinedOutput()
	response.EngineVersion = strings.TrimSpace(string(output))
	if len(response.EngineVersion) > 256 {
		response.EngineVersion = response.EngineVersion[:256]
	}
	if versionErr != nil {
		response.Details = strings.TrimSpace(response.Details + " " + versionErr.Error())
	}
	return response, nil
}

// updateSignatures authorizes and queues at most one manual FreshClam update.
func (service *antivirusTaskService) updateSignatures(ctx context.Context, envelope messaging.Envelope) (any, error) {
	var request antiviruscontract.UpdateSignaturesRequest
	if err := decodeRPCRequestError(envelope, &request, service.validator); err != nil {
		return nil, fmt.Errorf("invalid antivirus signature update request: %w", err)
	}
	if err := service.authorize(ctx, request.AccessToken); err != nil {
		return nil, err
	}
	service.mu.Lock()
	if service.updateState == "queued" || service.updateState == "updating" {
		service.mu.Unlock()
		return antiviruscontract.UpdateSignaturesResponse{State: "already_updating"}, nil
	}
	select {
	case service.updates <- struct{}{}:
		service.updateState = "queued"
		service.updateDetails = ""
		service.mu.Unlock()
		return antiviruscontract.UpdateSignaturesResponse{State: "queued"}, nil
	default:
		service.mu.Unlock()
		return nil, errors.New("antivirus signature update queue is full")
	}
}

// runSignatureUpdateWorker starts and monitors the isolated FreshClam unit.
func runSignatureUpdateWorker(ctx context.Context, updates <-chan struct{}, service *antivirusTaskService) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, open := <-updates:
			if !open {
				return
			}
			service.mu.Lock()
			service.updateState = "updating"
			service.mu.Unlock()
			commandCtx, cancel := context.WithTimeout(ctx, signatureUpdateTimeout)
			output, err := exec.CommandContext(commandCtx, sudoBinary, systemctlBinary, "start", "--no-block", updateUnitName).CombinedOutput()
			if err != nil {
				err = fmt.Errorf("start signature updater: %w: %s", err, strings.TrimSpace(string(output)))
			}
			if err == nil {
				err = waitForSignatureUpdate(commandCtx)
			}
			cancel()
			service.mu.Lock()
			if err != nil {
				service.updateState = "failed"
				service.updateDetails = "FreshClam update failed; see the updater service journal."
				service.logger.Error("antivirus signature update failed", "error", err)
			} else {
				service.updateState = "idle"
				service.updateDetails = ""
			}
			service.mu.Unlock()
		}
	}
}

// waitForSignatureUpdate polls the fixed systemd unit until FreshClam exits.
func waitForSignatureUpdate(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		output, err := exec.CommandContext(ctx, systemctlBinary, "show", "--property=ActiveState", "--value", updateUnitName).Output()
		if err != nil {
			return err
		}
		state := strings.TrimSpace(string(output))
		if state == "failed" {
			return errors.New("FreshClam updater unit failed")
		}
		if state == "inactive" {
			result, resultErr := exec.CommandContext(ctx, systemctlBinary, "show", "--property=Result", "--value", updateUnitName).Output()
			if resultErr != nil {
				return resultErr
			}
			if strings.TrimSpace(string(result)) != "success" {
				return errors.New("FreshClam updater unit exited unsuccessfully")
			}
			return nil
		}
		if state != "activating" && state != "active" {
			return fmt.Errorf("unexpected FreshClam updater state %q", state)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// start validates authorization and folder scope before queuing clamscan.
func (service *antivirusTaskService) start(ctx context.Context, envelope messaging.Envelope) (any, error) {
	var request antiviruscontract.StartScanRequest
	if err := decodeRPCRequestError(envelope, &request, service.validator); err != nil {
		service.logger.Warn("antivirus scan request rejected", "reason", err)
		return nil, fmt.Errorf("invalid antivirus scan request: %w", err)
	}
	if err := service.authorize(ctx, request.AccessToken); err != nil {
		return antiviruscontract.StartScanResponse{}, err
	}
	mount, folder, err := resolveScanFolder(request.Folder)
	if err != nil {
		return antiviruscontract.StartScanResponse{}, err
	}
	spec := execTaskSpec(mount, folder, service.cfg.ScanTimeout)
	service.mu.Lock()
	snapshot, err := service.manager.Start(spec)
	if err != nil {
		service.mu.Unlock()
		return antiviruscontract.StartScanResponse{}, err
	}
	service.targets[snapshot.ID] = scanTarget{mount: mount, folder: folder}
	service.mu.Unlock()
	return antiviruscontract.StartScanResponse{TaskID: snapshot.ID, State: string(snapshot.State)}, nil
}

// queueScanResult maps one completed command snapshot to the reporter's bounded result channel.
func queueScanResult(effects chan<- scanResult, mount mountinfo.Record, folder string, snapshot backgroundtasks.Snapshot) {
	resultMount := mount
	resultMount.Point = folder
	result := scanResult{mount: resultMount, evidence: snapshot.Output}
	switch {
	case snapshot.State == backgroundtasks.StateCancelled:
		result.status = scanCancelled
	case snapshot.ExitCode != nil && *snapshot.ExitCode == 1:
		result.status = scanInfected
	case snapshot.State == backgroundtasks.StateSucceeded:
		result.status = scanClean
	default:
		result.status = scanIncomplete
		result.err = errors.New(snapshot.Error)
	}
	if result.status != scanCancelled {
		result.eventID = newScanEventID()
	}
	// Block when the report queue is full; shutdown keeps the reporter draining
	// until this result has been accounted for.
	effects <- result
}

// forwardScanTaskResults consumes manager snapshots and feeds the antivirus effect pipeline.

func forwardScanTaskResults(snapshots <-chan backgroundtasks.Snapshot, service *antivirusTaskService) {
	for snapshot := range snapshots {
		service.mu.Lock()
		target, exists := service.targets[snapshot.ID]
		service.mu.Unlock()
		if !exists {
			service.logger.Error("on-demand antivirus result has no validated scan target", "task_id", snapshot.ID)
			continue
		}
		queueScanResult(service.effects, target.mount, target.folder, snapshot)
	}
}

// retry validates authorization and starts a fresh attempt for a finished scan.
func (service *antivirusTaskService) retry(ctx context.Context, envelope messaging.Envelope) (any, error) {
	var request antiviruscontract.TaskRequest
	if !decodeRPCRequest(envelope, &request, service.validator) {
		return nil, errors.New("invalid antivirus task request")
	}
	if err := service.authorize(ctx, request.AccessToken); err != nil {
		return antiviruscontract.RetryResponse{}, err
	}
	service.mu.Lock()
	snapshot, err := service.manager.Retry(request.TaskID)
	if err != nil {
		service.mu.Unlock()
		return antiviruscontract.RetryResponse{}, err
	}
	if target, exists := service.targets[request.TaskID]; exists {
		service.targets[snapshot.ID] = target
	}
	service.mu.Unlock()
	return antiviruscontract.RetryResponse{PreviousTaskID: request.TaskID, TaskID: snapshot.ID, State: string(snapshot.State)}, nil
}

// status returns one authenticated scan task snapshot.
func (service *antivirusTaskService) status(ctx context.Context, envelope messaging.Envelope) (any, error) {
	var request antiviruscontract.TaskRequest
	if !decodeRPCRequest(envelope, &request, service.validator) {
		return nil, errors.New("invalid antivirus task request")
	}
	if err := service.authorize(ctx, request.AccessToken); err != nil {
		return antiviruscontract.StatusResponse{}, err
	}
	snapshot, err := service.manager.Status(request.TaskID)
	if err != nil {
		return antiviruscontract.StatusResponse{}, err
	}
	return antiviruscontract.StatusResponse{TaskID: snapshot.ID, State: string(snapshot.State), StartedAt: snapshot.StartedAt, UpdatedAt: snapshot.UpdatedAt, FinishedAt: snapshot.FinishedAt, ExitCode: snapshot.ExitCode, Error: snapshot.Error, Output: snapshot.Output, OutputTruncated: snapshot.OutputTruncated, Progress: progressContract(snapshot.Progress)}, nil
}

// authorize requires a current authenticated identity with a security role.
func (service *antivirusTaskService) authorize(ctx context.Context, token string) error {
	var validated authcontract.ValidateAccessTokenResponse
	if err := service.client.Request(ctx, authcontract.ValidateAccessTokenRPCSubject, authcontract.ValidateAccessTokenRequest{AccessToken: token}, &validated); err != nil {
		return err
	}
	if !validated.Valid || validated.Status != authcontract.StatusAuthenticated || strings.TrimSpace(validated.Username) == "" {
		return errors.New("unauthorized task request")
	}
	var roles rbaccontract.GetSubjectRolesResponse
	if err := service.client.Request(ctx, rbaccontract.GetSubjectRolesRPCSubject, rbaccontract.GetSubjectRolesRequest{Username: validated.Username}, &roles); err != nil {
		return err
	}
	if !roleauth.HasAnyRole(roles.Groups, roleauth.AllowedRoles(roleauth.RequirementSecurity)) {
		return errors.New("unauthorized task request")
	}
	return nil
}

// resolveScanFolder validates any administrator-selected directory and records
// its current mount identity so execution can reject a changed mount.
func resolveScanFolder(value string) (mountinfo.Record, string, error) {
	if !filepath.IsAbs(value) || filepath.Clean(value) != value {
		return mountinfo.Record{}, "", errors.New("scan folder must be a clean absolute path")
	}
	resolved, err := filepath.EvalSymlinks(value)
	if err != nil || filepath.Clean(resolved) != value {
		return mountinfo.Record{}, "", errors.New("scan folder must exist and contain no symlinks")
	}
	info, err := os.Stat(value)
	if err != nil || !info.IsDir() {
		return mountinfo.Record{}, "", errors.New("scan target must be a directory")
	}
	mounts, err := mountinfo.ReadCurrent()
	if err != nil {
		return mountinfo.Record{}, "", err
	}
	var selected mountinfo.Record
	for _, mount := range mounts {
		rel, relErr := filepath.Rel(mount.Point, value)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if len(mount.Point) > len(selected.Point) {
			selected = mount
		}
	}
	if selected.Point == "" {
		return mountinfo.Record{}, "", errors.New("scan folder has no current mount")
	}
	if err := liveMount(selected); err != nil {
		return mountinfo.Record{}, "", err
	}
	return selected, value, nil
}

// execTaskSpec maps a verified folder to fixed ClamAV options and bounded task output.
func execTaskSpec(mount mountinfo.Record, folder string, timeout time.Duration) backgroundtasks.CommandSpec {
	validateFolder := func() error {
		current, currentFolder, err := resolveScanFolder(folder)
		if err != nil {
			return err
		}
		if current != mount || currentFolder != folder {
			return errors.New("scan folder mount identity changed")
		}
		return nil
	}
	return backgroundtasks.CommandSpec{
		Name: sudoBinary, Args: []string{"-n", clamscanBinary, "--recursive=yes", "--cross-fs=no", "--follow-dir-symlinks=0", "--follow-file-symlinks=0", "--suppress-ok-results", "--", folder}, Env: []string{"PATH=/usr/bin:/bin", "LANG=C"}, Timeout: timeout, OutputLimit: maxScanOutput,
		BeforeRun: validateFolder, AfterRun: validateFolder,
	}
}

// progressContract converts the runner's optional progress into the shared NATS DTO.
func progressContract(progress *backgroundtasks.Progress) *backgroundtaskscontract.StatusCheckProgress {
	if progress == nil {
		return nil
	}
	return &backgroundtaskscontract.StatusCheckProgress{Percent: progress.Percent, Message: progress.Message}
}

// decodeRPCRequest rejects oversized, trailing, unknown-field, or invalid RPC payloads.
func decodeRPCRequest(envelope messaging.Envelope, output any, validate *validator.Validate) bool {
	return decodeRPCRequestError(envelope, output, validate) == nil
}

// decodeRPCRequestError validates one bounded JSON request without exposing its values.
func decodeRPCRequestError(envelope messaging.Envelope, output any, validate *validator.Validate) error {
	if len(envelope.Payload) == 0 || len(envelope.Payload) > 16*1024 {
		return errors.New("request size is empty or exceeds 16 KiB")
	}
	decoder := json.NewDecoder(strings.NewReader(string(envelope.Payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(output); err != nil {
		return fmt.Errorf("invalid JSON request: %w", err)
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return errors.New("request contains trailing JSON data")
	}
	if err := validate.Struct(output); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}
	return nil
}

// decodeMountEvent enforces the versioned transport schema and bounded payload
// before the coordinator performs any eligibility or filesystem work.
func decodeMountEvent(envelope messaging.Envelope, validate *validator.Validate) (mountcontract.LifecycleEvent, error) {
	var event mountcontract.LifecycleEvent
	if len(envelope.Payload) == 0 || len(envelope.Payload) > maxMountEventBytes {
		return event, errors.New("mount event payload size invalid")
	}
	decoder := json.NewDecoder(strings.NewReader(string(envelope.Payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return event, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return event, errors.New("mount event contains trailing JSON")
	}
	if err := validate.Struct(event); err != nil {
		return event, err
	}
	if event.MountPoint == "" || len(event.MountPoint) > 4096 {
		return event, errors.New("invalid mount point")
	}
	return event, nil
}

// coordinateScans reconciles existing mounts once, then consumes lifecycle
// events. It keeps at most one active scan per mount ID and cancels old scans
// on unmount or identity change. Queue saturation is logged explicitly.
func coordinateScans(ctx context.Context, cfg antivirusConfig, incoming <-chan mountcontract.LifecycleEvent,
	jobs chan<- scanJob, results <-chan scanResult, effects chan<- scanResult, log sharedlogger.Logger,
) {
	defer close(jobs)
	active := make(map[uint64]context.CancelFunc)
	activeMounts := make(map[uint64]mountinfo.Record)
	completed := make(map[uint64]mountinfo.Record)
	pending := make(map[uint64]mountinfo.Record)
	schedule := func(mount mountinfo.Record) {
		if prefix := excludedMountPrefix(cfg, mount.Point); prefix != "" {
			log.Warn("mount scan skipped by configured exclusion", "mount_id", mount.ID,
				"mount_point", mount.Point, "excluded_prefix", prefix, "filesystem_type", mount.FSType)
			return
		}
		if !eligibleMount(cfg, mount) || activeMounts[mount.ID] == mount || completed[mount.ID] == mount {
			return
		}
		if cancel := active[mount.ID]; cancel != nil {
			cancel()
			pending[mount.ID] = mount
			return
		}
		jobCtx, cancel := context.WithCancel(ctx)
		select {
		case jobs <- scanJob{mount: mount, ctx: jobCtx}:
			active[mount.ID] = cancel
			activeMounts[mount.ID] = mount
		case <-ctx.Done():
			cancel()
			log.Warn("scan not queued during shutdown", "mount_id", mount.ID)
		default:
			cancel()
			log.Error("antivirus scan queue full", "mount_id", mount.ID)
		}
	}
	if mounts, err := mountinfo.ReadCurrent(); err != nil {
		log.Error("startup mount reconciliation failed", "error", err)
	} else {
		for _, mount := range mounts {
			schedule(mount)
		}
	}
	for incoming != nil || len(active) > 0 {
		select {
		case event, ok := <-incoming:
			if !ok {
				incoming = nil
				continue
			}
			if event.Action == mountcontract.Unmounted {
				for localID, mount := range activeMounts {
					if sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
						if cancel := active[localID]; cancel != nil {
							cancel()
						}
						delete(pending, localID)
						delete(completed, localID)
					}
				}
				for localID, mount := range pending {
					if sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
						delete(pending, localID)
					}
				}
				for localID, mount := range completed {
					if sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
						delete(completed, localID)
					}
				}
				continue
			}
			mounts, err := mountinfo.ReadCurrent()
			if err != nil {
				log.Error("mount recheck failed", "error", err)
				continue
			}
			mount, exists := findMountByIdentity(mounts, event)
			if !exists {
				log.Warn("stale mount notification", "event_id", event.EventID,
					"mount_id", event.MountID, "mount_point", event.MountPoint,
					"source", event.Source, "filesystem_type", event.FSType)
				continue
			}
			schedule(mount)
		case result, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			if cancel := active[result.mount.ID]; cancel != nil {
				cancel()
			}
			delete(active, result.mount.ID)
			delete(activeMounts, result.mount.ID)
			if result.status == scanClean || result.status == scanInfected {
				completed[result.mount.ID] = result.mount
			}
			select {
			case effects <- result:
			case <-ctx.Done():
				log.Error("scan result not queued during shutdown", "mount_id", result.mount.ID, "status", result.status)
			}
			if next, exists := pending[result.mount.ID]; exists {
				delete(pending, result.mount.ID)
				schedule(next)
			}
		}
	}
}

// findMountByIdentity resolves a lifecycle notification in the receiver's
// mount namespace, where kernel mount IDs may differ from the producer's IDs.
func findMountByIdentity(mounts map[uint64]mountinfo.Record, event mountcontract.LifecycleEvent) (mountinfo.Record, bool) {
	if mount, exists := mounts[event.MountID]; exists && sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
		return mount, true
	}
	var match mountinfo.Record
	found := false
	for _, mount := range mounts {
		if !sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
			continue
		}
		if found {
			return mountinfo.Record{}, false
		}
		match = mount
		found = true
	}
	return match, found
}

// sameMountIdentity compares mount metadata shared across mount namespaces.
func sameMountIdentity(mount mountinfo.Record, point, source, fsType string) bool {
	return mount.Point == point && mount.Source == source && mount.FSType == fsType
}

// runScanWorker serially executes bounded scans and drains queued jobs as
// cancelled results after process cancellation.
func runScanWorker(_ context.Context, jobs <-chan scanJob, results chan<- scanResult, timeout time.Duration) {
	defer close(results)
	for job := range jobs {
		results <- scanMount(job.ctx, job.mount, timeout)
	}
}

// reportScans emits malware through the security logging manager's existing
// alert path, which the security email notifier already consumes.
func reportScans(ctx context.Context, effects <-chan scanResult, client messaging.Client, log sharedlogger.Logger) {
	for result := range effects {
		if result.status == scanCancelled {
			log.Warn("antivirus scan cancelled", "mount_id", result.mount.ID)
			continue
		}
		if result.status == scanIncomplete {
			log.Warn("antivirus scan incomplete", "mount_id", result.mount.ID,
				"mount_point", result.mount.Point, "error", result.err)
		}
		if result.eventID == "" {
			id := newScanEventID()
			result.eventID = id
		}
		alertSent := false
		for {
			attemptCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := publishScanResult(attemptCtx, result, &alertSent, client)
			if cancel != nil {
				cancel()
			}
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				log.Error("scan result undelivered at shutdown", "mount_id", result.mount.ID, "error", err)
				break
			}
			log.Error("scan result publication failed; retrying", "mount_id", result.mount.ID, "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// publishScanResult distinguishes clean scans, malware and incomplete scans. The scan
// result owns a unique event ID so retries update only that scan's event.
func publishScanResult(ctx context.Context, result scanResult, alertSent *bool, client messaging.Client) error {
	if result.status == scanInfected {
		message := "Malware detected on " + result.mount.Point
		if len(message) > 256 {
			message = message[:256]
		}
		evidence := result.evidence
		if len(evidence) > 512 {
			evidence = evidence[:512]
		}
		email := true
		if !*alertSent {
			if err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
				EventID: result.eventID, Category: "Malware detection", Severity: enum.SeverityError,
				CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
				Message: message, TriggerValue: evidence, Email: &email,
				Mitigate: "Review the scan evidence and affected files. Quarantine or remove confirmed malware, then scan the affected storage again.",
			}); err != nil {
				return err
			}
			// The manager handles event creation asynchronously; do not publish
			// another active create while retrying its immediate resolution.
			*alertSent = true
		}
		response := loggingcontract.OKResponse{}
		findingCount := strings.Count(result.evidence, " FOUND")
		resolved := fmt.Sprintf("Malware detection recorded for scan on %s (%d findings)", result.mount.Point, findingCount)
		if len(resolved) > 256 {
			resolved = resolved[:256]
		}
		if err := client.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: result.eventID, Status: enum.StatusNormal, Message: &resolved}, &response); err != nil {
			return err
		}
		if !response.OK {
			return errors.New("security logging manager rejected malware event resolution")
		}
		return nil
	}
	if result.status == scanClean {
		message := "Antivirus scan clean on " + result.mount.Point
		if len(message) > 256 {
			message = message[:256]
		}
		if !*alertSent {
			if err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
				EventID: result.eventID, Category: "Antivirus scan", Severity: enum.SeverityInfo,
				CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
				Message: message,
			}); err != nil {
				return err
			}
			*alertSent = true
		}
		response := loggingcontract.OKResponse{}
		if err := client.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: result.eventID, Status: enum.StatusNormal}, &response); err != nil {
			return err
		}
		if !response.OK {
			return errors.New("security logging manager rejected clean scan resolution")
		}
		return nil
	}
	message := "Antivirus scan incomplete on " + result.mount.Point
	if result.err != nil {
		message += ": " + result.err.Error()
	}
	if len(message) > 256 {
		message = message[:256]
	}
	if *alertSent {
		return nil
	}
	evidence := result.evidence
	if len(evidence) > 512 {
		evidence = evidence[:512]
	}
	err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
		EventID: result.eventID, Category: "Antivirus scan", Severity: enum.SeverityWarning,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
		Message: message, TriggerValue: evidence,
		Mitigate: "Review the scanner error and evidence, resolve the scan failure, then rerun the scan before relying on its result.",
	})
	if err != nil {
		return err
	}
	*alertSent = true
	return nil
}

// monitorEventIDs shares the resources monitor's time-seeded counter convention.
var monitorEventIDs = eventmanager.NewManager(uint64(time.Now().Unix()) % 99_999_999)

// newScanEventID allocates a readable ID for one new event.
func newScanEventID() string {
	return monitorEventIDs.NextEventID("secav")
}
