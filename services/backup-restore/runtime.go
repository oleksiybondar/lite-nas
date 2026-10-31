package backuprestore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"gopkg.in/ini.v1"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/config"
	sharedcontracts "lite-nas/shared/contracts"
	contract "lite-nas/shared/contracts/backuprestore"
	managercontract "lite-nas/shared/contracts/servicesmanager"
	"lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/roleauth"
)

// Runtime owns configured file targets and serialized host mutations.
type Runtime struct {
	specs    map[string]Spec
	store    Store
	verifier authtoken.Verifier
	validate *validator.Validate
	client   messaging.Client
	mutex    sync.Mutex
	session  *mergeSession
}

// Serve starts the on-demand NATS service and blocks until shutdown.
func Serve(ctx context.Context, configPath string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("backup restore service requires root")
	}
	reader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return err
	}
	configuration, err := config.LoadINI(reader)
	if err != nil {
		return err
	}
	sharedConfig, err := config.LoadSharedAuthTokenConfig(configuration)
	if err != nil {
		return err
	}
	specDir, storeRoot, err := runtimePaths(configuration)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(storeRoot, 0o700); err != nil {
		return err
	}
	if info, err := os.Lstat(storeRoot); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("backup store must be a directory")
	}
	if err := os.Chmod(storeRoot, 0o700); err != nil {
		return err
	}
	specs, err := LoadSpecs(specDir)
	if err != nil {
		return err
	}
	for _, spec := range specs {
		for _, selection := range spec.Selections {
			if selection.Path == storeRoot || strings.HasPrefix(selection.Path, storeRoot+"/") || selection.Directory && strings.HasPrefix(storeRoot, selection.Path+"/") {
				return fmt.Errorf("backup store overlaps application %s selection %s", spec.Application, selection.Path)
			}
		}
	}
	certificate, err := os.ReadFile(sharedConfig.AuthTokens.VerificationCert)
	if err != nil {
		return err
	}
	key, err := authtoken.ParseEd25519CertificatePublicKeyPEM(certificate)
	if err != nil {
		return err
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{
		Issuer: sharedConfig.AuthTokens.Issuer, Audience: sharedConfig.AuthTokens.Audience,
		ClockSkew: sharedConfig.AuthTokens.ClockSkew,
	}, key)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientServerInfra(sharedcontracts.ServiceBackupRestore, sharedConfig.Logging, sharedConfig.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	runtime := &Runtime{specs: specs, store: Store{Root: storeRoot}, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: infra.Client}
	if err := runtime.register(infra.Server); err != nil {
		return err
	}
	infra.Logger.Info("backup restore service started", "applications", len(specs))
	<-ctx.Done()
	return ctx.Err()
}

// runtimePaths reads the two service-owned directories from validated startup config.
func runtimePaths(configuration *ini.File) (string, string, error) {
	section := configuration.Section("backup")
	specDir := section.Key("spec_dir").String()
	storeRoot := section.Key("store_root").String()
	if !filepath.IsAbs(specDir) || !filepath.IsAbs(storeRoot) || specDir == "/" || storeRoot == "/" || filepath.Clean(specDir) != specDir || filepath.Clean(storeRoot) != storeRoot {
		return "", "", fmt.Errorf("backup directories must be clean absolute paths")
	}
	return specDir, storeRoot, nil
}

// decodeRequest enforces the complete JSON request shape and administrator token.
func decodeRequest[T any](runtime *Runtime, envelope messaging.Envelope) (T, error) {
	var request T
	limit := 16 * 1024
	if _, ok := any(request).(contract.MergeWriteRequest); ok {
		limit = 128 * 1024
	}
	if len(envelope.Payload) > limit {
		return request, fmt.Errorf("request too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return request, fmt.Errorf("unexpected trailing JSON")
	}
	if err := runtime.validate.Struct(request); err != nil {
		return request, err
	}
	token := ""
	switch value := any(request).(type) {
	case contract.ApplicationsRequest:
		token = value.AccessToken
	case contract.ApplicationRequest:
		token = value.AccessToken
	case contract.BackupRequest:
		token = value.AccessToken
	case contract.BackupAllRequest:
		token = value.AccessToken
	case contract.RestoreRequest:
		token = value.AccessToken
	case contract.VersionRequest:
		token = value.AccessToken
	case contract.MergeStartRequest:
		token = value.AccessToken
	case contract.MergeReadRequest:
		token = value.AccessToken
	case contract.MergeWriteRequest:
		token = value.AccessToken
	case contract.MergeActionRequest:
		token = value.AccessToken
	case contract.MergeSaveRequest:
		token = value.AccessToken
	}
	claims, err := runtime.verifier.Verify(token)
	if err != nil || !roleauth.MatchesRequirement(claims.Roles, roleauth.RequirementAdministrator) {
		return request, fmt.Errorf("administrator authorization required")
	}
	return request, nil
}

// register binds validated RPC operations to the shared messaging server.
func (runtime *Runtime) register(server messaging.Server) error {
	for _, operation := range []struct {
		subject string
		handler messaging.RPCHandler
	}{
		{contract.MergeStartSubject, runtime.mergeStart},
		{contract.MergeReadSubject, runtime.mergeRead},
		{contract.MergeStatusSubject, runtime.mergeStatus},
		{contract.MergeWriteSubject, runtime.mergeWrite},
		{contract.MergeApplySubject, runtime.mergeApply},
		{contract.MergeSaveSubject, runtime.mergeSave},
		{contract.MergeDiscardSubject, runtime.mergeDiscard},
	} {
		if err := server.RegisterRPC(operation.subject, operation.handler); err != nil {
			return err
		}
	}
	if err := server.RegisterRPC(contract.ApplicationsSubject, runtime.applications); err != nil {
		return err
	}
	if err := server.RegisterRPC(contract.BackedUpSubject, runtime.backedUp); err != nil {
		return err
	}
	if err := server.RegisterRPC(contract.VersionsSubject, runtime.versions); err != nil {
		return err
	}
	if err := server.RegisterRPC(contract.BackupSubject, runtime.backup); err != nil {
		return err
	}
	if err := server.RegisterRPC(contract.BackupAllSubject, runtime.backupAll); err != nil {
		return err
	}
	if err := server.RegisterRPC(contract.DiffSubject, runtime.diff); err != nil {
		return err
	}
	return server.RegisterRPC(contract.RestoreSubject, runtime.restore)
}

// backedUp returns applications with stored history, including removed specs.
func (runtime *Runtime) backedUp(_ context.Context, envelope messaging.Envelope) (any, error) {
	if _, err := decodeRequest[contract.ApplicationsRequest](runtime, envelope); err != nil {
		return nil, err
	}
	applications, err := runtime.store.BackedUpApplications()
	return contract.ApplicationsResponse{Applications: applications}, err
}

// diff compares a stored version with current files without returning content.
func (runtime *Runtime) diff(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.VersionRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	spec, ok := runtime.specs[request.Application]
	if !ok {
		return nil, fmt.Errorf("unknown application")
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	differences, err := runtime.store.Diff(spec, request.Version)
	if err != nil {
		return nil, err
	}
	response := contract.DiffResponse{Differences: make([]contract.Difference, 0, len(differences))}
	for _, difference := range differences {
		response.Differences = append(response.Differences, contract.Difference{Path: difference.Path, Status: difference.Status})
	}
	return response, nil
}

// applications returns the configured application inventory after authorization.
func (runtime *Runtime) applications(_ context.Context, envelope messaging.Envelope) (any, error) {
	if _, err := decodeRequest[contract.ApplicationsRequest](runtime, envelope); err != nil {
		return nil, err
	}
	applications := make([]string, 0, len(runtime.specs))
	for application := range runtime.specs {
		applications = append(applications, application)
	}
	sort.Strings(applications)
	return contract.ApplicationsResponse{Applications: applications}, nil
}

// versions lists completed versions for a configured application.
func (runtime *Runtime) versions(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.ApplicationRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	versions, err := runtime.store.Versions(request.Application)
	return contract.VersionsResponse{Versions: versions}, err
}

// backup serializes snapshot creation with restore operations.
func (runtime *Runtime) backup(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.BackupRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	spec, ok := runtime.specs[request.Application]
	if !ok {
		return nil, fmt.Errorf("unknown application")
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	claims, err := runtime.verifier.Verify(request.AccessToken)
	if err != nil {
		return nil, err
	}
	version, err := runtime.store.BackupWithMetadata(spec, request.Name, request.Description, claims.Login)
	return contract.BackupResponse{Version: version}, err
}

// backupAll visits configured applications one by one and reports every outcome.
func (runtime *Runtime) backupAll(_ context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.BackupAllRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	claims, err := runtime.verifier.Verify(request.AccessToken)
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	names := make([]string, 0, len(runtime.specs))
	for name := range runtime.specs {
		names = append(names, name)
	}
	sort.Strings(names)
	response := contract.BackupAllResponse{Results: make([]contract.BackupAllResult, 0, len(names))}
	for _, name := range names {
		spec := runtime.specs[name]
		result := contract.BackupAllResult{Application: name}
		present := false
		for _, selection := range spec.Selections {
			_, err := os.Lstat(selection.Path)
			if err == nil {
				present = true
				break
			}
			if !os.IsNotExist(err) {
				present = true
				break
			}
		}
		if !present {
			result.Status = "skipped"
			response.Results = append(response.Results, result)
			continue
		}
		version, err := runtime.store.BackupWithMetadata(spec, request.Name, request.Description, claims.Login)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
		} else {
			result.Status = "completed"
			result.Version = version
		}
		response.Results = append(response.Results, result)
	}
	return response, nil
}

// restore serializes privileged writes with backup creation.
func (runtime *Runtime) restore(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, err := decodeRequest[contract.RestoreRequest](runtime, envelope)
	if err != nil {
		return nil, err
	}
	spec, ok := runtime.specs[request.Application]
	if !ok {
		return nil, fmt.Errorf("unknown application")
	}
	mode := RestoreMode(request.Mode)
	if mode == "" {
		mode = ReplaceMode
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	if err := runtime.store.Restore(spec, request.Version, mode); err != nil {
		return nil, err
	}
	response := contract.RestoreResponse{Restored: true}
	for _, unit := range spec.RestartUnits {
		var result managercontract.ActionResponse
		err := runtime.client.Request(ctx, managercontract.RestartSubject, managercontract.ActionRequest{AccessToken: request.AccessToken, Unit: unit}, &result)
		if err != nil || !result.Completed {
			if err == nil {
				err = fmt.Errorf("restart was not confirmed")
			}
			response.Warnings = append(response.Warnings, fmt.Sprintf("%s: %v", unit, err))
		}
	}
	return response, nil
}
