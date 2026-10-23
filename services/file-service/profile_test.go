package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"lite-nas/shared/fileaccess"
)

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
