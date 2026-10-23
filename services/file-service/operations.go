package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/file-service/filesystem"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

// pathInput is the validated body for creating one directory.
type pathInput struct {
	Path string `json:"path" minLength:"1" maxLength:"4096"`
}

// transferInput names an existing source and the complete destination path.
type transferInput struct {
	Source      string `json:"source" minLength:"1" maxLength:"4096"`
	Destination string `json:"destination" minLength:"1" maxLength:"4096"`
}

// decodeOperationDTO validates JSON with Huma before filesystem or business work.
func decodeOperationDTO(writer http.ResponseWriter, request *http.Request, value any) bool {
	request.Body = http.MaxBytesReader(writer, request.Body, 8192)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return false
	}
	var raw map[string]any
	if json.Unmarshal(body, &raw) != nil || raw == nil {
		return false
	}
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)
	schema := registry.Schema(reflect.TypeOf(value).Elem(), true, "FileOperationInput")
	result := &huma.ValidateResult{}
	huma.Validate(registry, schema, huma.NewPathBuffer(nil, 0), huma.ModeWriteToServer, raw, result)
	if len(result.Errors) != 0 {
		return false
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value) == nil && decoder.Decode(new(any)) == io.EOF
}

// operationJob is an in-memory operation that remains available across browser reloads.
type operationJob struct {
	ID             string `json:"job_id"`
	State          string `json:"state"`
	Percent        int    `json:"percent"`
	CompletedBytes int64  `json:"completed_bytes"`
	TotalBytes     int64  `json:"total_bytes"`
	CompletedItems int    `json:"completed_items"`
	TotalItems     int    `json:"total_items"`
	ErrorCode      string `json:"error_code,omitempty"`
	owner          string
	uid            string
	kind           string
	source         string
	destination    string
	subject        fileaccess.Subject
}

// operationUpdate carries worker progress to the manager's output channel.
type operationUpdate struct {
	id         string
	state      string
	bytes      int64
	items      int
	totalBytes int64
	totalItems int
	errorCode  string
}

// operationManager serializes mutation jobs using explicit input and output channels.
type operationManager struct {
	mu      sync.Mutex
	jobs    map[string]*operationJob
	input   chan *operationJob
	output  chan operationUpdate
	service *fileService
}

// newOperationManager starts the queue worker and progress collector.
func newOperationManager(service *fileService) *operationManager {
	manager := &operationManager{jobs: make(map[string]*operationJob), input: make(chan *operationJob, 128), output: make(chan operationUpdate, 128), service: service}
	go manager.collect()
	go manager.work()
	return manager
}

// collect applies worker progress without exposing mutable job records.
func (manager *operationManager) collect() {
	for update := range manager.output {
		manager.mu.Lock()
		job := manager.jobs[update.id]
		if job != nil {
			if update.state != "" {
				job.State = update.state
			}
			job.CompletedBytes += update.bytes
			job.CompletedItems += update.items
			if update.totalItems > 0 {
				job.TotalItems = update.totalItems
			}
			if update.totalBytes > 0 {
				job.TotalBytes = update.totalBytes
			}
			job.ErrorCode = update.errorCode
			if job.State == "done" {
				job.Percent = 100
			} else if job.TotalItems > 0 {
				// Bytes show large-file progress; item units advance empty directories.
				const itemUnit = int64(1 << 20)
				job.Percent = int(100 * (job.CompletedBytes + int64(job.CompletedItems)*itemUnit) / (job.TotalBytes + int64(job.TotalItems)*itemUnit))
				if job.Percent > 99 {
					job.Percent = 99
				}
			}
		}
		manager.mu.Unlock()
	}
}

// work executes queued jobs sequentially inside the File Service process.
func (manager *operationManager) work() {
	for job := range manager.input {
		manager.output <- operationUpdate{id: job.ID, state: "running"}
		if err := manager.service.runOperation(context.Background(), job, manager.output); err != nil {
			manager.output <- operationUpdate{id: job.ID, state: "error", errorCode: "operation_failed"}
			continue
		}
		manager.output <- operationUpdate{id: job.ID, state: "done"}
	}
}

// enqueue reserves one job ID and adds it to the serialized input channel.
func (manager *operationManager) enqueue(subject fileaccess.Subject, kind, source, destination string) (*operationJob, error) {
	id, err := newUploadID()
	if err != nil {
		return nil, err
	}
	job := &operationJob{ID: id, State: "pending", owner: subject.Username, uid: subject.UID, kind: kind, source: source, destination: destination, subject: subject}
	manager.mu.Lock()
	if len(manager.input) == cap(manager.input) {
		manager.mu.Unlock()
		return nil, errors.New("queue full")
	}
	manager.jobs[id] = job
	manager.input <- job
	snapshot := *job
	manager.mu.Unlock()
	return &snapshot, nil
}

// getJob exposes only the authenticated owner's current job snapshot.
func (s *fileService) getJob(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	id := strings.TrimPrefix(request.URL.Path, "/jobs/")
	if len(id) != 48 || strings.Trim(id, "0123456789abcdef") != "" {
		writeError(writer, http.StatusBadRequest, "invalid_job_id")
		return
	}
	s.jobs.mu.Lock()
	job := s.jobs.jobs[id]
	if job == nil || job.owner != subject.Username || job.uid != subject.UID {
		s.jobs.mu.Unlock()
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	snapshot := *job
	s.jobs.mu.Unlock()
	writeJSON(writer, http.StatusOK, snapshot)
}

// createDirectory makes a single authorized directory as the verified Unix user.
func (s *fileService) createDirectory(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	var input pathInput
	if !decodeOperationDTO(writer, request, &input) || fileaccess.ValidateHostPath(input.Path) != nil || input.Path == "/" {
		writeError(writer, http.StatusBadRequest, "invalid_path")
		return
	}
	path, err := resolveNewDirectory(input.Path)
	if err != nil {
		writeError(writer, http.StatusConflict, "invalid_destination")
		return
	}
	if err := s.allowDestination(request.Context(), subject, path, true); err != nil {
		decisionError(writer, err)
		return
	}
	if err := runUserCommand(request.Context(), subject, "/bin/mkdir", path); err != nil {
		writeError(writer, http.StatusConflict, "mkdir_failed")
		return
	}
	writeJSON(writer, http.StatusCreated, map[string]string{"path": path})
}

// createTransfer queues a recursive copy or move after validating both paths.
func (s *fileService) createTransfer(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	var input transferInput
	if !decodeOperationDTO(writer, request, &input) || fileaccess.ValidateHostPath(input.Source) != nil || fileaccess.ValidateHostPath(input.Destination) != nil || input.Source == "/" || input.Destination == "/" {
		writeError(writer, http.StatusBadRequest, "invalid_path")
		return
	}
	requestedInfo, err := os.Lstat(input.Source)
	if err == nil && requestedInfo.Mode()&os.ModeSymlink != 0 {
		writeError(writer, http.StatusBadRequest, "unsupported_target")
		return
	}
	source, err := fileaccess.ResolveExisting(input.Source)
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	info, err := os.Lstat(source)
	if err != nil || (!info.IsDir() && !info.Mode().IsRegular()) {
		writeError(writer, http.StatusBadRequest, "unsupported_target")
		return
	}
	var destination string
	if info.IsDir() {
		destination, err = resolveNewDirectory(input.Destination)
	} else {
		destination, err = fileaccess.ResolveNewFile(input.Destination)
	}
	if err != nil {
		writeError(writer, http.StatusConflict, "invalid_destination")
		return
	}
	if source == destination || info.IsDir() && strings.HasPrefix(destination, source+"/") {
		writeError(writer, http.StatusConflict, "recursive_destination")
		return
	}
	if err := s.auth.Read(request.Context(), subject, source); err != nil {
		decisionError(writer, err)
		return
	}
	scope, err := s.policy.ReadScope(subject.Groups)
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "authorization_unavailable")
		return
	}
	if info.IsDir() && !scope.AllowsDirectory(source) || info.Mode().IsRegular() && !scope.AllowsFile(source) {
		decisionError(writer, fileaccess.ErrDenied)
		return
	}
	if err := s.allowDestination(request.Context(), subject, destination, info.IsDir()); err != nil {
		decisionError(writer, err)
		return
	}
	if request.URL.Path == "/move" {
		if err := s.policy.AllowMutation(source); err != nil {
			decisionError(writer, err)
			return
		}
		if err := s.auth.Operation(request.Context(), subject, source, rbaccontract.PathOperationDelete); err != nil {
			decisionError(writer, err)
			return
		}
	}
	job, err := s.jobs.enqueue(subject, strings.TrimPrefix(request.URL.Path, "/"), source, destination)
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "queue_unavailable")
		return
	}
	writeJSON(writer, http.StatusAccepted, job)
}

// resolveNewDirectory canonicalizes a directory target without following its final entry.
func resolveNewDirectory(path string) (string, error) {
	if err := fileaccess.ValidateHostPath(path); err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(parent, filepath.Base(path))
	info, err := os.Lstat(resolved)
	if errors.Is(err, os.ErrNotExist) {
		return resolved, nil
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fileaccess.ErrInvalidPath
	}
	return resolved, nil
}

// allowDestination checks policy and the parent mutation operation for a target.
func (s *fileService) allowDestination(ctx context.Context, subject fileaccess.Subject, path string, directory bool) error {
	if err := s.policy.AllowMutation(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil {
		if directory && info.IsDir() {
			return s.auth.Read(ctx, subject, path)
		}
		if !directory && info.Mode().IsRegular() {
			return s.auth.Operation(ctx, subject, path, rbaccontract.PathOperationReplace)
		}
		return fileaccess.ErrInvalidPath
	}
	return s.auth.Operation(ctx, subject, path, rbaccontract.PathOperationCreate)
}

// runUserCommand invokes only a fixed binary as the verified non-root Unix user.
func runUserCommand(ctx context.Context, subject fileaccess.Subject, binary string, paths ...string) error {
	username, err := filesystem.VerifyUnixIdentity(subject)
	if err != nil {
		return err
	}
	arguments := append([]string{"-n", "-u", username, "--", binary, "--"}, paths...)
	return exec.CommandContext(ctx, "/usr/bin/sudo", arguments...).Run()
}

// operationEntry records a source tree member and its relative destination.
type operationEntry struct {
	path      string
	relative  string
	directory bool
	size      int64
}

// runOperation walks an authorized source tree, then copies and optionally removes it.
func (s *fileService) runOperation(ctx context.Context, job *operationJob, updates chan<- operationUpdate) error {
	entries := make([]operationEntry, 0)
	var scope filesystem.ReadScope
	if job.kind != "delete" {
		var err error
		scope, err = s.policy.ReadScope(job.subject.Groups)
		if err != nil {
			return err
		}
	}
	err := filepath.WalkDir(job.source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 || (!entry.IsDir() && !entry.Type().IsRegular()) {
			return fileaccess.ErrInvalidPath
		}
		if job.kind != "delete" {
			if entry.IsDir() && !scope.AllowsDirectory(path) || entry.Type().IsRegular() && !scope.AllowsFile(path) {
				return fileaccess.ErrDenied
			}
			if err := s.auth.Read(ctx, job.subject, path); err != nil {
				return err
			}
		}
		if job.kind == "move" || job.kind == "delete" {
			if err := s.policy.AllowMutation(path); err != nil {
				return err
			}
			if err := s.auth.Operation(ctx, job.subject, path, rbaccontract.PathOperationDelete); err != nil {
				return err
			}
		}
		relative, err := filepath.Rel(job.source, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		entries = append(entries, operationEntry{path: path, relative: relative, directory: entry.IsDir(), size: info.Size()})
		return nil
	})
	if err != nil {
		return err
	}
	var totalBytes int64
	if job.kind != "delete" {
		for _, entry := range entries {
			if !entry.directory {
				totalBytes += entry.size
			}
		}
	}
	totalItems := len(entries)
	if job.kind == "move" {
		totalItems *= 2
	}
	updates <- operationUpdate{id: job.ID, totalBytes: totalBytes, totalItems: totalItems}
	if job.kind != "delete" {
		for _, entry := range entries {
			target := job.destination
			if entry.relative != "." {
				target = filepath.Join(target, entry.relative)
			}
			if entry.directory {
				resolved, err := resolveNewDirectory(target)
				if err != nil {
					return err
				}
				if err := s.allowDestination(ctx, job.subject, resolved, true); err != nil {
					return err
				}
				if _, err := os.Lstat(resolved); errors.Is(err, os.ErrNotExist) {
					if err := runUserCommand(ctx, job.subject, "/bin/mkdir", resolved); err != nil {
						return err
					}
				}
			} else {
				resolved, err := fileaccess.ResolveNewFile(target)
				if err != nil {
					return err
				}
				if err := s.allowDestination(ctx, job.subject, resolved, false); err != nil {
					return err
				}
				if err := s.copyFile(ctx, job, entry.path, resolved, updates); err != nil {
					return err
				}
			}
			updates <- operationUpdate{id: job.ID, items: 1}
		}
	}
	if job.kind == "move" || job.kind == "delete" {
		for index := len(entries) - 1; index >= 0; index-- {
			entry := entries[index]
			if err := s.policy.AllowMutation(entry.path); err != nil {
				return err
			}
			if err := s.auth.Operation(ctx, job.subject, entry.path, rbaccontract.PathOperationDelete); err != nil {
				return err
			}
			binary := "/bin/rm"
			if entry.directory {
				binary = "/bin/rmdir"
			}
			if err := runUserCommand(ctx, job.subject, binary, entry.path); err != nil {
				return err
			}
			updates <- operationUpdate{id: job.ID, items: 1}
		}
	}
	return nil
}

// copyFile streams source bytes through the existing user-scoped staged writer.
func (s *fileService) copyFile(ctx context.Context, job *operationJob, source, destination string, updates chan<- operationUpdate) error {
	s.store.mu.Lock()
	if s.store.targets[destination] != "" {
		s.store.mu.Unlock()
		return errors.New("destination has an active upload")
	}
	s.store.targets[destination] = job.ID
	s.store.mu.Unlock()
	defer func() {
		s.store.mu.Lock()
		if s.store.targets[destination] == job.ID {
			delete(s.store.targets, destination)
		}
		s.store.mu.Unlock()
	}()
	info, err := os.Lstat(destination)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	record := &uploadRecord{Target: destination}
	if err == nil {
		if !info.Mode().IsRegular() {
			return fileaccess.ErrInvalidPath
		}
		record.DestinationExists = true
		record.DestinationDevice, record.DestinationInode = fileIdentity(info)
	}
	created, err := invokeWriter(ctx, job.subject, "create", record, nil)
	if err != nil {
		return err
	}
	record.StageDevice, record.StageInode, record.StageUID = created.Device, created.Inode, created.UID
	defer func() { _, _ = invokeWriter(ctx, job.subject, "cleanup", record, nil) }()
	reader, err := openCat(ctx, job.subject, source)
	if err != nil {
		return err
	}
	defer reader.Close()
	hash := sha256.New()
	buffer := make([]byte, 4<<20)
	for {
		count, readErr := reader.Read(buffer)
		if count > 0 {
			chunk := buffer[:count]
			if _, err := hash.Write(chunk); err != nil {
				return err
			}
			result, err := invokeWriter(ctx, job.subject, "write", record, chunk)
			if err != nil || result.BytesWritten != int64(count) {
				return errors.New("copy write failed")
			}
			record.Offset = result.NextOffset
			updates <- operationUpdate{id: job.ID, bytes: int64(count)}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	staged, err := invokeWriter(ctx, job.subject, "checksum", record, nil)
	if err != nil || staged.SHA256 != hex.EncodeToString(hash.Sum(nil)) {
		return errors.New("copy checksum failed")
	}
	if _, err := invokeWriter(ctx, job.subject, "publish", record, nil); err != nil {
		return err
	}
	return nil
}
