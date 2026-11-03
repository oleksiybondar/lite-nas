package main

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"lite-nas/shared/fileaccess"
)

// mustCurrentProfileSubject binds the running Unix account to profile tests.
func mustCurrentProfileSubject(t *testing.T) fileaccess.Subject {
	t.Helper()
	account, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	if account.Uid == "0" || account.HomeDir != filepath.Join("/home", account.Username) {
		t.Skip("profile paths require a nonroot /home account")
	}
	return fileaccess.Subject{Username: account.Username, UID: account.Uid}
}

// mustPNGAvatar returns one fully encoded pixel for profile boundary tests.
func mustPNGAvatar(t *testing.T) []byte {
	t.Helper()
	imageData := image.NewRGBA(image.Rect(0, 0, 1, 1))
	imageData.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, imageData); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// TestValidAvatarRejectsMalformedImages covers profile upload input validation.
func TestValidAvatarRejectsMalformedImages(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("not an image"), []byte("\x89PNG\r\n\x1a\n")} {
		if validAvatar(data) {
			t.Fatalf("accepted malformed avatar %q", data)
		}
	}
}

// TestValidAvatarAcceptsPNG covers a complete supported image.
func TestValidAvatarAcceptsPNG(t *testing.T) {
	if !validAvatar(mustPNGAvatar(t)) {
		t.Fatal("rejected valid PNG avatar")
	}
}

// TestValidAvatarAcceptsJPEG covers the second supported image format.
func TestValidAvatarAcceptsJPEG(t *testing.T) {
	imageData := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, imageData, nil); err != nil {
		t.Fatal(err)
	}
	if !validAvatar(buffer.Bytes()) {
		t.Fatal("rejected valid JPEG avatar")
	}
}

// TestAvatarContentTypeUsesImageBytes covers extensionless `/f` rendering.
func TestAvatarContentTypeUsesImageBytes(t *testing.T) {
	directory := filepath.Join(t.TempDir(), profileDirectoryName)
	if err := os.Mkdir(directory, 0o771); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "avatar.image")
	if err := os.WriteFile(path, mustPNGAvatar(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := safeContentType(path, false); got != "image/png" {
		t.Fatalf("avatar MIME = %q, want image/png", got)
	}
}

// TestProfilePathRejectsUnverifiedIdentity covers the profile path authority boundary.
func TestProfilePathRejectsUnverifiedIdentity(t *testing.T) {
	if _, err := profilePath(fileaccess.Subject{Username: "../another-user"}, "avatar.image"); err == nil {
		t.Fatal("accepted unverified user path")
	}
}

// TestSharedAvatarOwnerRejectsOtherPaths covers the narrow cross-user read exception.
func TestSharedAvatarOwnerRejectsOtherPaths(t *testing.T) {
	for _, path := range []string{"/home/alice/.lite-nas/ui.json", "/home/alice/photos/avatar.image", "/home/alice/.lite-nas/../avatar.image"} {
		if owner := sharedAvatarOwner(path); owner != "" {
			t.Fatalf("accepted unrelated path %q", path)
		}
	}
}

// TestSharedAvatarOwnerRecognizesFixedPath covers the permitted avatar path.
func TestSharedAvatarOwnerRecognizesFixedPath(t *testing.T) {
	if owner := sharedAvatarOwner("/home/alice/.lite-nas/avatar.image"); owner != "alice" {
		t.Fatalf("avatar owner = %q, want alice", owner)
	}
}

// TestProfileDirectoryOwnedByRejectsOtherUid covers fixed profile directory ownership.
func TestProfileDirectoryOwnedByRejectsOtherUid(t *testing.T) {
	directory := filepath.Join(t.TempDir(), profileDirectoryName)
	if err := os.Mkdir(directory, 0o771); err != nil {
		t.Fatal(err)
	}
	exists, err := profileDirectoryOwnedBy(directory, fileaccess.Subject{UID: strconv.Itoa(os.Getuid() + 1)})
	if err == nil || exists {
		t.Fatalf("accepted directory owned by another user: exists=%t error=%v", exists, err)
	}
}

// TestProfileDocumentInputUsesHumaValidation covers the JSON profile boundary.
func TestProfileDocumentInputUsesHumaValidation(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
		want bool
	}{
		{name: "valid object", body: `{"document":{"preferredMode":"dark"}}`, want: true},
		{name: "valid language preference", body: `{"document":{"preferredMode":"dark","preferredLanguage":"uk"}}`, want: true},
		{name: "missing document", body: `{}`, want: false},
		{name: "nonobject document", body: `{"document":"dark"}`, want: false},
		{name: "unknown field", body: `{"document":{},"path":"/home/bob"}`, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/profile/ui", bytes.NewBufferString(test.body))
			var input profileDocumentInput
			if got := decodeOperationDTO(httptest.NewRecorder(), request, &input); got != test.want {
				t.Fatalf("profile DTO accepted=%t, want %t", got, test.want)
			}
		})
	}
}

// TestProfileResourceRejectsUnmappedName covers the fixed profile route boundary.
func TestProfileResourceRejectsUnmappedName(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/profile/other", nil)
	response := httptest.NewRecorder()
	(&fileService{}).profileResource(response, request, fileaccess.Subject{})
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

// TestProfileResourceRejectsUnsupportedMethod covers fixed profile mutation methods.
func TestProfileResourceRejectsUnsupportedMethod(t *testing.T) {
	for _, path := range []string{"/profile/ui", "/profile/moments"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, path, nil)
			response := httptest.NewRecorder()
			(&fileService{}).profileResource(response, request, mustCurrentProfileSubject(t))
			if response.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 405", response.Code)
			}
		})
	}
}

// TestProfileResourceRejectsInvalidDocument covers the validated document boundary.
func TestProfileResourceRejectsInvalidDocument(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", bytes.NewBufferString(`{"document":null}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	(&fileService{}).profileResource(response, request, mustCurrentProfileSubject(t))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

// TestProfileResourceRejectsInvalidAvatar covers invalid image payloads before publication.
func TestProfileResourceRejectsInvalidAvatar(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/profile/avatar", bytes.NewBufferString("not an image"))
	response := httptest.NewRecorder()
	(&fileService{}).profileResource(response, request, mustCurrentProfileSubject(t))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

// TestProfileResourceRejectsOversizedDocument covers the fixed settings size limit.
func TestProfileResourceRejectsOversizedDocument(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/profile/admin-panel", bytes.NewBufferString(`{"document":{"note":"`+strings.Repeat("x", 8192)+`"}}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	(&fileService{}).profileResource(response, request, mustCurrentProfileSubject(t))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

// TestProfileResourceRejectsOversizedAvatar covers bounded image request bodies.
func TestProfileResourceRejectsOversizedAvatar(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/profile/avatar", bytes.NewReader(make([]byte, maxProfileAvatarBytes+1)))
	response := httptest.NewRecorder()
	(&fileService{}).profileResource(response, request, mustCurrentProfileSubject(t))
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", response.Code)
	}
}

// TestProfileDocumentRespectsConfiguredStorageScope covers fixed paths outside a service root.
func TestProfileDocumentRespectsConfiguredStorageScope(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", bytes.NewBufferString(`{"document":{"preferredMode":"dark"}}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	service.profileResource(response, request, mustCurrentProfileSubject(t))
	if response.Code != http.StatusForbidden {
		t.Fatalf("out-of-scope profile status = %d, want 403", response.Code)
	}
}

// TestEnsureProfileDirectoryReturnsCommandFailure covers failed owner-scoped setup.
func TestEnsureProfileDirectoryReturnsCommandFailure(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	service.userCommand = func(context.Context, fileaccess.Subject, string, ...string) error {
		return fileaccess.ErrDenied
	}
	directory := filepath.Join(root, profileDirectoryName)
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", nil)
	subject := fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}
	if err := service.ensureProfileDirectory(request, subject, directory); !errors.Is(err, fileaccess.ErrDenied) {
		t.Fatalf("setup error = %v, want denied", err)
	}
}

// TestEnsureProfileDirectoryRejectsOutsideStorageRoot covers fixed-path policy enforcement.
func TestEnsureProfileDirectoryRejectsOutsideStorageRoot(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	directory := filepath.Join(t.TempDir(), profileDirectoryName)
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", nil)
	subject := fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}
	if err := service.ensureProfileDirectory(request, subject, directory); err == nil {
		t.Fatal("out-of-scope profile directory was accepted")
	}
}

// TestProfileDirectoryRaceAcceptsOwnerDirectory covers a concurrent successful setup.
func TestProfileDirectoryRaceAcceptsOwnerDirectory(t *testing.T) {
	subject := fileaccess.Subject{UID: strconv.Itoa(os.Getuid())}
	if err := profileDirectoryRaceResult(t.TempDir(), subject, fileaccess.ErrInvalidPath); err != nil {
		t.Fatalf("existing owner directory error = %v", err)
	}
}

// TestProfileDirectoryRaceRetainsFailure covers a directory still absent after setup.
func TestProfileDirectoryRaceRetainsFailure(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "missing")
	err := profileDirectoryRaceResult(directory, fileaccess.Subject{UID: strconv.Itoa(os.Getuid())}, fileaccess.ErrInvalidPath)
	if !errors.Is(err, fileaccess.ErrInvalidPath) {
		t.Fatalf("race error = %v, want invalid path", err)
	}
}

// TestProfileDirectoryRaceRejectsUnexpectedOwner covers a replaced profile directory.
func TestProfileDirectoryRaceRejectsUnexpectedOwner(t *testing.T) {
	directory := t.TempDir()
	subject := fileaccess.Subject{UID: strconv.Itoa(os.Getuid() + 1)}
	if err := profileDirectoryRaceResult(directory, subject, fileaccess.ErrInvalidPath); !errors.Is(err, fileaccess.ErrDenied) {
		t.Fatalf("replaced directory error = %v, want denied", err)
	}
}

// TestUploadProfileBytesUsesExistingUploadFlow covers fixed profile file publication.
func TestUploadProfileBytesUsesExistingUploadFlow(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	target := filepath.Join(root, "ui.json")
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", nil)
	response := httptest.NewRecorder()
	service.uploadProfileBytes(response, request, fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}, target, []byte(`{"preferredMode":"dark"}`), false)
	if got := mustReadWorkflowFile(t, target); got != `{"preferredMode":"dark"}` {
		t.Fatalf("profile bytes = %q", got)
	}
}

// TestUploadProfileAvatarPublishesReadableMode covers the fixed public avatar exception.
func TestUploadProfileAvatarPublishesReadableMode(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	target := filepath.Join(root, "avatar.image")
	request := httptest.NewRequest(http.MethodPut, "/profile/avatar", nil)
	response := httptest.NewRecorder()
	service.uploadProfileBytes(response, request, fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}, target, mustPNGAvatar(t), true)
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("avatar mode = %04o, want 0644", info.Mode().Perm())
	}
}

// TestProfileDirectoryOwnedByAcceptsCurrentOwner covers existing fixed profile storage.
func TestProfileDirectoryOwnedByAcceptsCurrentOwner(t *testing.T) {
	exists, err := profileDirectoryOwnedBy(t.TempDir(), fileaccess.Subject{UID: strconv.Itoa(os.Getuid())})
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("owner directory was rejected")
	}
}

// TestProfileDirectoryOwnedByReportsAbsentDirectory covers first-time setup.
func TestProfileDirectoryOwnedByReportsAbsentDirectory(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "missing")
	exists, err := profileDirectoryOwnedBy(directory, fileaccess.Subject{UID: strconv.Itoa(os.Getuid())})
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("absent directory reported as present")
	}
}

// TestEnsureProfileDirectoryAcceptsExistingOwnerDirectory covers an idempotent setup.
func TestEnsureProfileDirectoryAcceptsExistingOwnerDirectory(t *testing.T) {
	service := &fileService{}
	directory := t.TempDir()
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", nil)
	err := service.ensureProfileDirectory(request, fileaccess.Subject{UID: strconv.Itoa(os.Getuid())}, directory)
	if err != nil {
		t.Fatalf("existing directory error = %v", err)
	}
}

// TestEnsureProfileDirectoryCreatesOwnerStorage covers first-time setup through the service policy.
func TestEnsureProfileDirectoryCreatesOwnerStorage(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	directory := filepath.Join(root, profileDirectoryName)
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, _ string, _ ...string) error {
		return os.Mkdir(directory, 0o700)
	}
	request := httptest.NewRequest(http.MethodPut, "/profile/ui", nil)
	subject := fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}
	if err := service.ensureProfileDirectory(request, subject, directory); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(directory); err != nil {
		t.Fatalf("profile directory was not created: %v", err)
	}
}

// TestSharedAvatarRejectsUnknownOwner covers cross-user avatar lookup failures.
func TestSharedAvatarRejectsUnknownOwner(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/home/lite-nas-no-such-user/.lite-nas/avatar.image", nil)
	response := httptest.NewRecorder()
	(&fileService{}).serveSharedAvatar(response, request, fileaccess.Subject{Username: "alice"}, request.URL.Path)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

// TestSharedAvatarLeavesOwnerReadsOnNormalPath covers owner-specific read policy.
func TestSharedAvatarLeavesOwnerReadsOnNormalPath(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/home/alice/.lite-nas/avatar.image", nil)
	response := httptest.NewRecorder()
	if (&fileService{}).serveSharedAvatar(response, request, fileaccess.Subject{Username: "alice"}, request.URL.Path) {
		t.Fatal("owner avatar bypassed normal file authorization")
	}
}
