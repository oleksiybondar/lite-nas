package backuprestore

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	contract "lite-nas/shared/contracts/backuprestore"
	managercontract "lite-nas/shared/contracts/servicesmanager"
	"lite-nas/shared/messaging"
)

const (
	mergeFileLimit  = 1 << 20
	mergeTotalLimit = 32 << 20
	mergeChunkSize  = 64 << 10
)

// mergeFile retains bounded inputs and the client's staged result.
type mergeFile struct {
	header                            *tar.Header
	snapshot, current, result, upload []byte
	uploading, dirty, present         bool
}

// mergeSession is the sole process-local merge, owned by one authenticated subject.
type mergeSession struct {
	id, owner, version string
	spec               Spec
	files              map[string]*mergeFile
	applyFailed        bool
}

// mergeStart validates an archive and captures its selected regular files and live inputs.
func (runtime *Runtime) mergeStart(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeStartRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	claims, err := runtime.verifier.Verify(request.AccessToken)
	if err != nil {
		return nil, err
	}
	spec, ok := runtime.specs[request.Application]
	if !ok {
		return nil, fmt.Errorf("unknown application")
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	if runtime.session != nil {
		return nil, fmt.Errorf("merge session already active")
	}
	archive, err := runtime.store.OpenVersion(spec.Application, request.Version)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	stage, err := os.MkdirTemp(runtime.store.Root, ".merge-start-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage)
	entries, err := stageArchive(archive, stage)
	if err != nil {
		return nil, err
	}
	if err := validateRestoreEntries(entries, spec); err != nil {
		return nil, err
	}
	files := make(map[string]*mergeFile)
	response := contract.MergeStartResponse{Files: []contract.MergeFile{}, ApplyAvailable: true}
	var total int
	for _, entry := range entries {
		if entry.header.Typeflag != tar.TypeReg || entry.path == spec.SourcePath || !selectedPath(spec, entry.path) {
			continue
		}
		if entry.header.Size > mergeFileLimit || len(files) >= 32 {
			return nil, fmt.Errorf("merge file limit exceeded")
		}
		if err := rejectSymlinkParents(entry.path); err != nil {
			return nil, err
		}
		snapshot, err := os.ReadFile(entry.stage)
		if err != nil {
			return nil, err
		}
		current, err := boundedLiveFile(entry.path)
		if err != nil {
			return nil, err
		}
		if !mergeText(snapshot) || !mergeText(current) {
			return nil, fmt.Errorf("merge requires UTF-8 text: %s", entry.path)
		}
		_, statErr := os.Lstat(entry.path)
		if statErr != nil && !os.IsNotExist(statErr) {
			return nil, statErr
		}
		total += len(snapshot) + 2*len(current)
		if total > mergeTotalLimit {
			return nil, fmt.Errorf("merge session size limit exceeded")
		}
		files[entry.path] = &mergeFile{header: entry.header, snapshot: snapshot, current: current, result: append([]byte(nil), current...), present: statErr == nil, dirty: true}
		response.Files = append(response.Files, mergeFileInfo(entry.path, files[entry.path]))
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("snapshot has no editable regular files")
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, err
	}
	sort.Slice(response.Files, func(i, j int) bool { return response.Files[i].Path < response.Files[j].Path })
	response.SessionID = hex.EncodeToString(id[:])
	runtime.session = &mergeSession{id: response.SessionID, owner: claims.Subject, version: request.Version, spec: spec, files: files}
	return response, nil
}

// boundedLiveFile reads only a regular file within the merge size limit.
func boundedLiveFile(path string) ([]byte, error) {
	if err := rejectSymlinkParents(path); err != nil {
		return nil, err
	}
	pathInfo, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !pathInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("live merge path is not regular: %s", path)
	}
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > mergeFileLimit {
		return nil, fmt.Errorf("live merge file is not a bounded regular file: %s", path)
	}
	content, err := io.ReadAll(io.LimitReader(file, mergeFileLimit+1))
	if err != nil {
		return nil, err
	}
	if len(content) > mergeFileLimit {
		return nil, fmt.Errorf("live merge file too large")
	}
	return content, nil
}

// mergeText rejects binary data from the text merge boundary.
func mergeText(content []byte) bool {
	return utf8.Valid(content) && !strings.ContainsRune(string(content), 0)
}

// mergeFileInfo reports the comparison available to either client.
func mergeFileInfo(path string, file *mergeFile) contract.MergeFile {
	status := "changed"
	if !file.present {
		status = "snapshot-only"
	} else if sha256.Sum256(file.snapshot) == sha256.Sum256(file.current) {
		status = "same"
	}
	return contract.MergeFile{Path: path, SnapshotSize: len(file.snapshot), CurrentSize: len(file.current), Status: status}
}

// ownedSession checks the session ID and verified caller identity under runtime.mutex.
func (runtime *Runtime) ownedSession(id, token string) (*mergeSession, error) {
	claims, err := runtime.verifier.Verify(token)
	if err != nil {
		return nil, err
	}
	if runtime.session == nil || runtime.session.id != id || runtime.session.owner != claims.Subject {
		return nil, fmt.Errorf("merge session unavailable")
	}
	return runtime.session, nil
}

// mergeRead returns at most one chunk and the full input checksum.
func (runtime *Runtime) mergeRead(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeReadRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	session, err := runtime.ownedSession(request.SessionID, request.AccessToken)
	if err != nil {
		return nil, err
	}
	file := session.files[request.Path]
	if file == nil {
		return nil, fmt.Errorf("path is not selected for merge")
	}
	var content []byte
	switch request.Side {
	case "snapshot":
		content = file.snapshot
	case "current":
		content = file.current
	case "result":
		content = file.result
	default:
		return nil, fmt.Errorf("invalid merge side")
	}
	if request.Offset > len(content) {
		return nil, fmt.Errorf("invalid merge offset")
	}
	end := min(request.Offset+mergeChunkSize, len(content))
	checksum := sha256.Sum256(content)
	return contract.MergeReadResponse{Data: base64.StdEncoding.EncodeToString(content[request.Offset:end]), NextOffset: end, Done: end == len(content), SHA256: hex.EncodeToString(checksum[:])}, nil
}

// mergeStatus lets the owner resume a process-local session after the CLI exits.
func (runtime *Runtime) mergeStatus(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeActionRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	session, err := runtime.ownedSession(request.SessionID, request.AccessToken)
	if err != nil {
		return nil, err
	}
	response := contract.MergeStartResponse{SessionID: session.id, Files: make([]contract.MergeFile, 0, len(session.files)), ApplyAvailable: !session.applyFailed}
	for _, path := range sortedMergePaths(session.files) {
		file := session.files[path]
		response.Files = append(response.Files, mergeFileInfo(path, file))
	}
	return response, nil
}

// mergeWrite accepts sequential chunks and verifies the completed result.
func (runtime *Runtime) mergeWrite(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeWriteRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(request.Data)
	if err != nil || len(data) > mergeChunkSize || len(data) == 0 && !request.Final {
		return nil, fmt.Errorf("invalid merge chunk")
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	session, err := runtime.ownedSession(request.SessionID, request.AccessToken)
	if err != nil {
		return nil, err
	}
	file := session.files[request.Path]
	if file == nil {
		return nil, fmt.Errorf("path is not selected for merge")
	}
	if request.Offset == 0 {
		file.upload = nil
		file.uploading = true
	}
	if !file.uploading || request.Offset != len(file.upload) || len(file.upload)+len(data) > mergeFileLimit {
		return nil, fmt.Errorf("invalid merge upload offset or size")
	}
	file.upload = append(file.upload, data...)
	if request.Final {
		if !mergeText(file.upload) {
			return nil, fmt.Errorf("merged result must be UTF-8 text")
		}
		checksum := sha256.Sum256(file.upload)
		if request.SHA256 == "" || request.SHA256 != hex.EncodeToString(checksum[:]) {
			return nil, fmt.Errorf("merge checksum mismatch")
		}
		file.result = append([]byte(nil), file.upload...)
		file.upload = nil
		file.uploading = false
		file.dirty = true
	}
	return contract.MergeWriteResponse{NextOffset: request.Offset + len(data)}, nil
}

// mergeApply checks all live inputs before writing any staged result.
func (runtime *Runtime) mergeApply(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeActionRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	session, err := runtime.ownedSession(request.SessionID, request.AccessToken)
	if err != nil {
		return nil, err
	}
	if session.applyFailed {
		return nil, fmt.Errorf("previous apply failed; save or discard session")
	}
	for path, file := range session.files {
		if file.uploading {
			return nil, fmt.Errorf("unfinished merge upload")
		}
		current, err := boundedLiveFile(path)
		if err != nil {
			session.applyFailed = true
			return nil, err
		}
		_, statErr := os.Lstat(path)
		if statErr != nil && !os.IsNotExist(statErr) {
			session.applyFailed = true
			return nil, statErr
		}
		if (statErr == nil) != file.present {
			session.applyFailed = true
			return nil, fmt.Errorf("live file changed since merge start: %s", path)
		}
		if sha256.Sum256(current) != sha256.Sum256(file.current) {
			session.applyFailed = true
			return nil, fmt.Errorf("live file changed since merge start: %s", path)
		}
	}
	paths := sortedMergePaths(session.files)
	for _, path := range paths {
		file := session.files[path]
		if !file.dirty {
			continue
		}
		stage, err := os.CreateTemp(runtime.store.Root, ".merge-apply-*")
		if err != nil {
			session.applyFailed = true
			return nil, fmt.Errorf("partial merge apply: %w", err)
		}
		_, writeErr := stage.Write(file.result)
		closeErr := stage.Close()
		if writeErr == nil {
			writeErr = closeErr
		}
		if writeErr == nil {
			writeErr = rejectSymlinkParents(path)
		}
		if writeErr == nil {
			writeErr = restoreRegular(stagedEntry{header: file.header, path: path, stage: stage.Name()})
		}
		_ = os.Remove(stage.Name())
		if writeErr != nil {
			session.applyFailed = true
			return nil, fmt.Errorf("partial merge apply: %w", writeErr)
		}
	}
	response := contract.MergeActionResponse{Completed: true}
	for _, unit := range session.spec.RestartUnits {
		var result managercontract.ActionResponse
		err := runtime.client.Request(ctx, managercontract.RestartSubject, managercontract.ActionRequest{AccessToken: request.AccessToken, Unit: unit}, &result)
		if err != nil || !result.Completed {
			if err == nil {
				err = fmt.Errorf("restart was not confirmed")
			}
			response.Warnings = append(response.Warnings, fmt.Sprintf("%s: %v", unit, err))
		}
	}
	runtime.session = nil
	return response, nil
}

// sortedMergePaths makes multi-file apply deterministic.
func sortedMergePaths(files map[string]*mergeFile) []string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// mergeSave creates a new archive from the selected snapshot and staged bytes.
func (runtime *Runtime) mergeSave(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeSaveRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	session, err := runtime.ownedSession(request.SessionID, request.AccessToken)
	if err != nil {
		return nil, err
	}
	for _, file := range session.files {
		if file.uploading {
			return nil, fmt.Errorf("unfinished merge upload")
		}
	}
	claims, err := runtime.verifier.Verify(request.AccessToken)
	if err != nil {
		return nil, err
	}
	version, err := runtime.store.SaveMerged(session.spec, session.version, session.files, request.Name, request.Description, claims.Login)
	if err != nil {
		return nil, err
	}
	runtime.session = nil
	return contract.BackupResponse{Version: version}, nil
}

// mergeDiscard drops the sole active merge session.
func (runtime *Runtime) mergeDiscard(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.MergeActionRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	if _, err := runtime.ownedSession(request.SessionID, request.AccessToken); err != nil {
		return nil, err
	}
	runtime.session = nil
	return contract.MergeActionResponse{Completed: true}, nil
}

// SaveMerged publishes a new snapshot with staged content and the selected spec.
func (store Store) SaveMerged(spec Spec, sourceVersion string, files map[string]*mergeFile, name, description, createdBy string) (string, error) {
	directory := filepath.Join(store.Root, spec.Application)
	created := time.Now().UTC()
	base := created.Format("20060102T150405.000000000Z")
	for suffix := 0; suffix < 100; suffix++ {
		version := base
		if suffix != 0 {
			version = fmt.Sprintf("%s-%d", base, suffix)
		}
		temporary, err := os.CreateTemp(directory, ".merge-snapshot-*")
		if err != nil {
			return "", err
		}
		metadata := Metadata{Application: spec.Application, Version: version, RestartUnits: spec.RestartUnits, Name: name, Description: description, CreatedAt: created.Format(time.RFC3339Nano), CreatedBy: createdBy}
		err = store.writeMergedArchive(temporary, spec.Application, sourceVersion, metadata, files)
		if err == nil {
			err = temporary.Sync()
		}
		closeErr := temporary.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			_ = os.Remove(temporary.Name())
			return "", err
		}
		err = os.Link(temporary.Name(), filepath.Join(directory, version+".tar.gz"))
		_ = os.Remove(temporary.Name())
		if os.IsExist(err) {
			continue
		}
		return version, err
	}
	return "", fmt.Errorf("could not allocate merge version")
}

// writeMergedArchive copies the selected archive, replacing only staged regular members.
func (store Store) writeMergedArchive(destination io.Writer, application, version string, metadata Metadata, files map[string]*mergeFile) error {
	source, err := store.OpenVersion(application, version)
	if err != nil {
		return err
	}
	defer source.Close()
	compressed, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	output := gzip.NewWriter(destination)
	writer := tar.NewWriter(output)
	content, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	if err := writer.WriteHeader(&tar.Header{Name: "metadata.json", Mode: 0o600, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
		return err
	}
	if _, err := writer.Write(content); err != nil {
		return err
	}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Name == "metadata.json" {
			continue
		}
		file := files["/"+header.Name]
		if file != nil && file.dirty {
			header.Size = int64(len(file.result))
			header.ModTime = time.Now().UTC()
		}
		if err := writer.WriteHeader(header); err != nil {
			return err
		}
		if file != nil && file.dirty {
			_, err = writer.Write(file.result)
		} else {
			_, err = io.Copy(writer, reader)
		}
		if err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return output.Close()
}
