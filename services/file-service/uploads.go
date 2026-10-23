package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"

	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

const (
	chunkBudget      = 60 << 20
	minChunkSize     = 1 << 20
	maxChunkSize     = 16 << 20
	maxBrowserUpload = 10 << 30
)

// chunkWindow accounts for the maximum buffered bytes of active PATCH requests.
type chunkWindow struct {
	mu   sync.Mutex
	used int64
}

// reserve admits one bounded chunk before its body is read.
func (window *chunkWindow) reserve(length int64) bool {
	window.mu.Lock()
	defer window.mu.Unlock()
	if length < 0 || window.used+length > chunkBudget {
		return false
	}
	window.used += length
	return true
}

// release returns a completed request's buffer reservation.
func (window *chunkWindow) release(length int64) {
	window.mu.Lock()
	window.used -= length
	window.mu.Unlock()
}

// available reports current capacity for the upload-window resource.
func (window *chunkWindow) available() int64 {
	window.mu.Lock()
	defer window.mu.Unlock()
	return chunkBudget - window.used
}

// createInput is the validated JSON metadata for one resumable upload.
type createInput struct {
	Destination string `json:"destination" minLength:"1" maxLength:"4096"`
	Filename    string `json:"filename" minLength:"1" maxLength:"255"`
	Size        int64  `json:"size" minimum:"0"`
	SHA256      string `json:"sha256" pattern:"^[0-9a-fA-F]{64}$"`
	ContentType string `json:"content_type,omitempty" maxLength:"255"`
	ChunkSize   *int64 `json:"chunk_size,omitempty" minimum:"1048576" maximum:"16777216"`
}

// validateCreateDTO applies the repository's Huma schema at the JSON boundary.
func validateCreateDTO(body []byte, input *createInput) bool {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return false
	}
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)
	schema := registry.Schema(reflect.TypeOf(createInput{}), true, "CreateUploadInput")
	result := &huma.ValidateResult{}
	huma.Validate(registry, schema, huma.NewPathBuffer(nil, 0), huma.ModeWriteToServer, raw, result)
	if len(result.Errors) != 0 {
		return false
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(input) == nil && decoder.Decode(new(any)) == io.EOF && validateCreateInput(*input)
}

// validateCreateInput rejects missing and noncanonical metadata at the HTTP boundary.
func validateCreateInput(input createInput) bool {
	if fileaccess.ValidateHostPath(input.Destination) != nil || input.Filename == "" ||
		filepath.Base(input.Filename) != input.Filename || strings.ContainsAny(input.Filename, "\r\n\x00") ||
		input.Size < 0 || len(input.SHA256) != 64 || len(input.ContentType) > 255 ||
		(input.ChunkSize != nil && (*input.ChunkSize < minChunkSize || *input.ChunkSize > maxChunkSize)) {
		return false
	}
	checksum, err := hex.DecodeString(input.SHA256)
	return err == nil && len(checksum) == sha256.Size
}

// uploadPermission resolves and authorizes the current destination state.
func (s *fileService) uploadPermission(ctx context.Context, subject fileaccess.Subject, requested string) (string, bool, uint64, uint64, error) {
	target, err := fileaccess.ResolveNewFile(requested)
	if err != nil {
		return "", false, 0, 0, err
	}
	if err := s.policy.AllowMutation(target); err != nil {
		return "", false, 0, 0, err
	}
	info, err := os.Lstat(target)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", false, 0, 0, err
	}
	operation := rbaccontract.PathOperationCreate
	var device, inode uint64
	if err == nil {
		if !info.Mode().IsRegular() {
			return "", false, 0, 0, fileaccess.ErrInvalidPath
		}
		operation = rbaccontract.PathOperationReplace
		device, inode = fileIdentity(info)
	}
	if err := s.auth.Operation(ctx, subject, target, operation); err != nil {
		return "", false, 0, 0, err
	}
	return target, operation == rbaccontract.PathOperationReplace, device, inode, nil
}

// createUpload accepts metadata only after auth, policy, RBAC, and target reservation.
func (s *fileService) createUpload(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	key := request.Header.Get("Idempotency-Key")
	if len(key) < 8 || len(key) > 128 || strings.ContainsAny(key, "\r\n\x00") {
		writeError(writer, http.StatusBadRequest, "invalid_idempotency_key")
		return
	}
	var input createInput
	request.Body = http.MaxBytesReader(writer, request.Body, 4096)
	body, err := io.ReadAll(request.Body)
	if err != nil || !validateCreateDTO(body, &input) {
		writeError(writer, http.StatusBadRequest, "invalid_upload_metadata")
		return
	}
	target, exists, device, inode, err := s.uploadPermission(request.Context(), subject, input.Destination)
	if err != nil {
		decisionError(writer, err)
		return
	}
	requestedChunkSize := int64(0)
	effectiveChunkSize := s.cfg.ChunkSize
	if input.ChunkSize != nil {
		requestedChunkSize = *input.ChunkSize
		effectiveChunkSize = requestedChunkSize
	}
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	if id := s.store.keys[subject.Username+"\x00"+key]; id != "" {
		record := s.store.records[id]
		if record != nil && record.Target == target && record.Filename == input.Filename &&
			record.Size == input.Size && record.SHA256 == input.SHA256 && record.ContentType == input.ContentType &&
			record.RequestedChunkSize == requestedChunkSize {
			writer.Header().Set("Location", "/f/proprietary/upload/"+id)
			writeJSON(writer, http.StatusOK, progress(record))
			return
		}
		writeError(writer, http.StatusConflict, "idempotency_conflict")
		return
	}
	if s.store.targets[target] != "" {
		writeError(writer, http.StatusConflict, "target_busy")
		return
	}
	if len(s.store.records)*4096 >= maxMetadataBytes {
		writeError(writer, http.StatusServiceUnavailable, "upload_metadata_full")
		return
	}
	id, err := newUploadID()
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "upload_unavailable")
		return
	}
	record := &uploadRecord{
		ID: id, Owner: subject.Username, UID: subject.UID,
		Target: target, Filename: input.Filename, Size: input.Size,
		SHA256: strings.ToLower(input.SHA256), ContentType: input.ContentType,
		ChunkSize: effectiveChunkSize, RequestedChunkSize: requestedChunkSize,
		IdempotencyKey: key, State: "active",
		Expires: time.Now().Add(s.cfg.UploadTTL), DestinationExists: exists,
		DestinationDevice: device, DestinationInode: inode,
	}
	result, err := s.writerCommand(request.Context(), subject, "create", record, nil)
	if err != nil {
		writeError(writer, http.StatusForbidden, "stage_create_failed")
		return
	}
	record.StageDevice, record.StageInode, record.StageUID = result.Device, result.Inode, result.UID
	if strconv.FormatUint(uint64(result.UID), 10) != subject.UID || s.store.persist(record) != nil {
		_, _ = s.writerCommand(request.Context(), subject, "cleanup", record, nil)
		writeError(writer, http.StatusServiceUnavailable, "upload_state_failed")
		return
	}
	s.store.add(record)
	writer.Header().Set("Location", "/f/proprietary/upload/"+id)
	writeJSON(writer, http.StatusCreated, progress(record))
}

// progress returns only client-facing upload fields.
func progress(record *uploadRecord) map[string]any {
	return map[string]any{
		"upload_id": record.ID, "state": record.State,
		"total_size": record.Size, "acknowledged_bytes": record.Offset,
		"next_chunk_index": record.NextIndex, "chunk_size": record.ChunkSize,
		"maximum_chunk_size": maxChunkSize,
	}
}

// uploadResource routes an authenticated upload ID operation.
func (s *fileService) uploadResource(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	suffix := strings.TrimPrefix(request.URL.Path, "/proprietary/upload/")
	completion := strings.HasSuffix(suffix, "/completion")
	id := strings.TrimSuffix(suffix, "/completion")
	if len(id) != 48 || strings.Trim(id, "0123456789abcdef") != "" || strings.Contains(id, "/") {
		writeError(writer, http.StatusNotFound, "upload_not_found")
		return
	}
	record, err := s.store.find(id)
	if err != nil || record.Owner != subject.Username || record.UID != subject.UID {
		writeError(writer, http.StatusNotFound, "upload_not_found")
		return
	}
	record.mu.Lock()
	defer record.mu.Unlock()
	if time.Now().After(record.Expires) && record.State != "completed" {
		writeError(writer, http.StatusGone, "upload_expired")
		return
	}
	if record.State == "completed" {
		if err := s.auth.Operation(request.Context(), subject, record.Target, rbaccontract.PathOperationReplace); err != nil {
			decisionError(writer, err)
			return
		}
		switch {
		case completion && request.Method == http.MethodPut:
			writeJSON(writer, http.StatusOK, map[string]any{"upload_id": record.ID, "state": "completed", "sha256": record.SHA256, "resource": "/f" + record.Target})
		case !completion && request.Method == http.MethodGet:
			writeJSON(writer, http.StatusOK, progress(record))
		case !completion && request.Method == http.MethodHead:
			writer.Header().Set("Upload-Length", strconv.FormatInt(record.Size, 10))
			writer.Header().Set("Upload-Offset", strconv.FormatInt(record.Offset, 10))
			writer.Header().Set("Upload-Next-Chunk-Index", strconv.FormatInt(record.NextIndex, 10))
			writer.WriteHeader(http.StatusOK)
		default:
			writeError(writer, http.StatusConflict, "already_completed")
		}
		return
	}
	target, exists, device, inode, err := s.uploadPermission(request.Context(), subject, record.Target)
	if err != nil || target != record.Target || exists != record.DestinationExists ||
		(exists && (device != record.DestinationDevice || inode != record.DestinationInode)) {
		if err != nil {
			decisionError(writer, err)
		} else {
			writeError(writer, http.StatusConflict, "destination_changed")
		}
		return
	}
	switch {
	case completion && request.Method == http.MethodPut:
		s.completeUpload(writer, request, subject, record)
	case completion:
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
	case request.Method == http.MethodGet:
		writeJSON(writer, http.StatusOK, progress(record))
	case request.Method == http.MethodHead:
		writer.Header().Set("Upload-Length", strconv.FormatInt(record.Size, 10))
		writer.Header().Set("Upload-Offset", strconv.FormatInt(record.Offset, 10))
		writer.Header().Set("Upload-Next-Chunk-Index", strconv.FormatInt(record.NextIndex, 10))
		writer.Header().Set("Upload-Chunk-Size", strconv.FormatInt(record.ChunkSize, 10))
		writer.WriteHeader(http.StatusOK)
	case request.Method == http.MethodPatch:
		s.patchUpload(writer, request, subject, record)
	case request.Method == http.MethodDelete:
		s.cancelUpload(writer, request, subject, record)
	default:
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

// patchUpload commits exactly one bounded next chunk or acknowledges its replay.
func (s *fileService) patchUpload(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, record *uploadRecord) {
	if record.State != "active" {
		writeError(writer, http.StatusConflict, "upload_not_active")
		return
	}
	offset, err := strconv.ParseInt(request.Header.Get("Upload-Offset"), 10, 64)
	if err != nil || offset < 0 || request.ContentLength <= 0 || request.ContentLength > record.ChunkSize {
		writeError(writer, http.StatusBadRequest, "invalid_chunk")
		return
	}
	if offset != record.Offset && offset != record.Offset-record.LastChunkSize {
		writeChunkConflict(writer, record)
		return
	}
	if request.ContentLength > record.Size-offset ||
		(request.ContentLength != record.ChunkSize && offset+request.ContentLength != record.Size) {
		writeError(writer, http.StatusBadRequest, "invalid_chunk_size")
		return
	}
	if !s.window.reserve(record.ChunkSize) {
		writer.Header().Set("Retry-After", "2")
		writeError(writer, http.StatusServiceUnavailable, "upload_window_full")
		return
	}
	defer s.window.release(record.ChunkSize)
	data := make([]byte, request.ContentLength)
	if _, err := io.ReadFull(request.Body, data); err != nil {
		writeError(writer, http.StatusBadRequest, "incomplete_chunk")
		return
	}
	var extra [1]byte
	if n, _ := request.Body.Read(extra[:]); n != 0 {
		writeError(writer, http.StatusBadRequest, "invalid_chunk_size")
		return
	}
	checksum := sha256.Sum256(data)
	if offset < record.Offset {
		if record.LastChunkSize == int64(len(data)) && record.LastChunkSHA256 == hex.EncodeToString(checksum[:]) {
			writeChunkSuccess(writer, record)
		} else {
			writeChunkConflict(writer, record)
		}
		return
	}
	result, err := s.writerCommand(request.Context(), subject, "write", record, data)
	if err != nil || result.NextOffset != offset+int64(len(data)) {
		writeError(writer, http.StatusServiceUnavailable, "stage_write_failed")
		return
	}
	previousOffset, previousIndex := record.Offset, record.NextIndex
	previousSize, previousSHA, previousExpiry := record.LastChunkSize, record.LastChunkSHA256, record.Expires
	record.Offset = result.NextOffset
	record.NextIndex++
	record.LastChunkSize = int64(len(data))
	record.LastChunkSHA256 = hex.EncodeToString(checksum[:])
	record.Expires = time.Now().Add(s.cfg.UploadTTL)
	if err := s.store.persist(record); err != nil {
		record.Offset, record.NextIndex, record.LastChunkSize = previousOffset, previousIndex, previousSize
		record.LastChunkSHA256, record.Expires = previousSHA, previousExpiry
		writeError(writer, http.StatusServiceUnavailable, "upload_state_failed")
		return
	}
	writeChunkSuccess(writer, record)
}

// conventionalUpload streams a destination-first multipart file in bounded chunks.
func (s *fileService) conventionalUpload(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	multipartReader, err := request.MultipartReader()
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_multipart")
		return
	}
	destinationPart, err := multipartReader.NextPart()
	if err != nil || destinationPart.FormName() != "destination" {
		writeError(writer, http.StatusBadRequest, "destination_required_first")
		return
	}
	destinationBytes, err := io.ReadAll(io.LimitReader(destinationPart, 4097))
	if err != nil || len(destinationBytes) == 0 || len(destinationBytes) > 4096 {
		writeError(writer, http.StatusBadRequest, "invalid_destination")
		return
	}
	target, exists, device, inode, err := s.uploadPermission(request.Context(), subject, string(destinationBytes))
	if err != nil {
		decisionError(writer, err)
		return
	}
	s.store.mu.Lock()
	if s.store.targets[target] != "" {
		s.store.mu.Unlock()
		writeError(writer, http.StatusConflict, "target_busy")
		return
	}
	id, err := newUploadID()
	if err != nil {
		s.store.mu.Unlock()
		writeError(writer, http.StatusServiceUnavailable, "upload_unavailable")
		return
	}
	record := &uploadRecord{
		ID: id, Owner: subject.Username, UID: subject.UID, Target: target,
		ChunkSize: s.cfg.ChunkSize, State: "active", Expires: time.Now().Add(s.cfg.UploadTTL),
		DestinationExists: exists, DestinationDevice: device, DestinationInode: inode,
		IdempotencyKey: id,
	}
	result, err := s.writerCommand(request.Context(), subject, "create", record, nil)
	if err != nil {
		s.store.mu.Unlock()
		writeError(writer, http.StatusForbidden, "stage_create_failed")
		return
	}
	record.StageDevice, record.StageInode, record.StageUID = result.Device, result.Inode, result.UID
	if s.store.persist(record) != nil {
		s.store.mu.Unlock()
		_, _ = s.writerCommand(context.Background(), subject, "cleanup", record, nil)
		writeError(writer, http.StatusServiceUnavailable, "upload_state_failed")
		return
	}
	s.store.add(record)
	s.store.mu.Unlock()
	defer func() {
		if record.State != "completed" {
			if _, err := s.writerCommand(context.Background(), subject, "cleanup", record, nil); err != nil {
				return // Keep durable metadata and the target reservation for the expiry worker.
			}
		}
		_ = s.store.remove(record)
	}()
	filePart, err := multipartReader.NextPart()
	if err != nil || filePart.FormName() != "file" {
		writeError(writer, http.StatusBadRequest, "file_part_required")
		return
	}
	if err := s.writeMultipartFile(request.Context(), subject, record, filePart); err != nil {
		if errors.Is(err, errUploadTooLarge) {
			writeError(writer, http.StatusRequestEntityTooLarge, "upload_too_large")
		} else {
			writeError(writer, http.StatusServiceUnavailable, "stage_write_failed")
		}
		return
	}
	if _, err := multipartReader.NextPart(); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "unexpected_multipart_part")
		return
	}
	if _, currentExists, currentDevice, currentInode, err := s.uploadPermission(request.Context(), subject, record.Target); err != nil || currentExists != record.DestinationExists ||
		(currentExists && (currentDevice != record.DestinationDevice || currentInode != record.DestinationInode)) {
		writeError(writer, http.StatusConflict, "destination_changed")
		return
	}
	checksum, err := s.writerCommand(request.Context(), subject, "checksum", record, nil)
	if err != nil || checksum.Size != record.Offset {
		writeError(writer, http.StatusServiceUnavailable, "checksum_failed")
		return
	}
	if _, err := s.writerCommand(request.Context(), subject, "publish", record, nil); err != nil {
		writeError(writer, http.StatusConflict, "publish_failed")
		return
	}
	record.State = "completed"
	writeJSON(writer, http.StatusCreated, map[string]any{
		"resource": "/f" + record.Target,
		"size":     record.Offset, "sha256": checksum.SHA256,
	})
}

var errUploadTooLarge = errors.New("conventional upload exceeds 10 GiB")

// writeMultipartFile forwards each bounded part chunk to the user-scoped writer.
func (s *fileService) writeMultipartFile(ctx context.Context, subject fileaccess.Subject, record *uploadRecord, part *multipart.Part) error {
	buffer := make([]byte, s.cfg.ChunkSize)
	for {
		count, err := io.ReadFull(part, buffer)
		if count > 0 {
			if record.Offset+int64(count) > maxBrowserUpload {
				return errUploadTooLarge
			}
			result, writeErr := s.writerCommand(ctx, subject, "write", record, buffer[:count])
			if writeErr != nil || result.NextOffset != record.Offset+int64(count) {
				return errors.New("writer failed")
			}
			record.Offset = result.NextOffset
			record.Expires = time.Now().Add(s.cfg.UploadTTL)
			if saveErr := s.store.persist(record); saveErr != nil {
				return saveErr
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// writeChunkSuccess acknowledges a durable chunk without a response body.
func writeChunkSuccess(writer http.ResponseWriter, record *uploadRecord) {
	writer.Header().Set("Upload-Offset", strconv.FormatInt(record.Offset, 10))
	writer.Header().Set("Upload-Next-Chunk-Index", strconv.FormatInt(record.NextIndex, 10))
	writer.WriteHeader(http.StatusNoContent)
}

// writeChunkConflict reports the authoritative committed position.
func writeChunkConflict(writer http.ResponseWriter, record *uploadRecord) {
	writeJSON(writer, http.StatusConflict, map[string]any{"error": map[string]any{
		"code": "chunk_conflict", "upload_offset": record.Offset,
		"next_chunk_index": record.NextIndex,
	}})
}

// completeUpload verifies the stage before the writer's same-directory rename.
func (s *fileService) completeUpload(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, record *uploadRecord) {
	if record.State == "completed" {
		writeJSON(writer, http.StatusOK, map[string]any{"upload_id": record.ID, "state": "completed", "sha256": record.SHA256, "resource": "/f" + record.Target})
		return
	}
	if record.Offset != record.Size {
		writeError(writer, http.StatusConflict, "upload_incomplete")
		return
	}
	checksum, err := s.writerCommand(request.Context(), subject, "checksum", record, nil)
	if err != nil || checksum.Size != record.Size || checksum.SHA256 != record.SHA256 {
		writeError(writer, http.StatusConflict, "checksum_mismatch")
		return
	}
	if _, err := s.writerCommand(request.Context(), subject, "publish", record, nil); err != nil {
		_, _ = s.writerCommand(request.Context(), subject, "cleanup", record, nil)
		writeError(writer, http.StatusConflict, "publish_failed")
		return
	}
	record.State = "completed"
	record.Expires = time.Now().Add(s.cfg.UploadTTL)
	if err := s.store.persist(record); err != nil {
		writeError(writer, http.StatusServiceUnavailable, "upload_state_failed")
		return
	}
	s.store.mu.Lock()
	delete(s.store.targets, record.Target)
	s.store.mu.Unlock()
	writeJSON(writer, http.StatusOK, map[string]any{"upload_id": record.ID, "state": "completed", "sha256": checksum.SHA256, "resource": "/f" + record.Target})
}

// cancelUpload removes only the recorded stage and then its metadata.
func (s *fileService) cancelUpload(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, record *uploadRecord) {
	if record.State == "completed" {
		writeError(writer, http.StatusConflict, "already_completed")
		return
	}
	if _, err := s.writerCommand(request.Context(), subject, "cleanup", record, nil); err != nil {
		writeError(writer, http.StatusServiceUnavailable, "cleanup_failed")
		return
	}
	if err := s.store.remove(record); err != nil {
		writeError(writer, http.StatusServiceUnavailable, "upload_state_failed")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

// uploadWindow reports current weighted chunk admission capacity.
func (s *fileService) uploadWindow(writer http.ResponseWriter) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"available_buffer_bytes": s.window.available(),
		"maximum_chunk_size":     maxChunkSize,
	})
}

// cleanupLoop triggers expiry work on startup and every three hours.
func (s *fileService) cleanupLoop(ctx context.Context) {
	s.cleanupExpired(ctx)
	ticker := time.NewTicker(s.cfg.CleanupEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpired(ctx)
		}
	}
}

// cleanupExpired removes verified stale stages without consuming upload buffers.
func (s *fileService) cleanupExpired(ctx context.Context) {
	s.store.mu.Lock()
	records := make([]*uploadRecord, 0, len(s.store.records))
	for _, record := range s.store.records {
		records = append(records, record)
	}
	s.store.mu.Unlock()
	for _, record := range records {
		record.mu.Lock()
		if time.Now().Before(record.Expires) {
			record.mu.Unlock()
			continue
		}
		if record.State == "completed" {
			_ = s.store.remove(record)
		} else {
			subject := fileaccess.Subject{Username: record.Owner, UID: record.UID}
			if _, err := s.writerCommand(ctx, subject, "cleanup", record, nil); err == nil {
				_ = s.store.remove(record)
			}
		}
		record.mu.Unlock()
	}
}
