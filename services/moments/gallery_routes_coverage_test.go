package main

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// mustUploadGalleryThumbnail publishes a small valid cover for route tests.
func mustUploadGalleryThumbnail(t *testing.T, router http.Handler, id string) {
	t.Helper()
	var imageBytes bytes.Buffer
	if err := png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/api/galleries/"+id+"/thumbnail/image", &imageBytes)
	request.Header.Set("Authorization", "Bearer test")
	request.Header.Set("Content-Type", "image/png")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("thumbnail upload status = %d: %s", response.Code, response.Body.String())
	}
}

// TestCollectionFileReferenceRoundTrip covers explicit file membership.
func TestCollectionFileReferenceRoundTrip(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	path := "/home/alice/photos/photo.jpg"
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", `{"path":"`+path+`"}`, http.StatusOK)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/collections/"+id+"/references", "", http.StatusOK)
	refs := mustGalleryJSON[struct {
		Items []collectionReference `json:"items"`
	}](t, response)
	if len(refs.Items) != 1 || refs.Items[0].Path != path || refs.Items[0].Type != "FILE" {
		t.Fatalf("file references = %+v", refs.Items)
	}
}

// TestCollectionDirectoryReferenceRoundTrip covers explicit directory membership.
func TestCollectionDirectoryReferenceRoundTrip(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	path := "/home/alice/photos"
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/directories", `{"path":"`+path+`"}`, http.StatusOK)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/collections/"+id+"/references", "", http.StatusOK)
	refs := mustGalleryJSON[struct {
		Items []collectionReference `json:"items"`
	}](t, response)
	if len(refs.Items) != 1 || refs.Items[0].Path != path || refs.Items[0].Type != "DIRECTORY" {
		t.Fatalf("directory references = %+v", refs.Items)
	}
}

// TestCollectionReferenceRemoval clears one existing membership without deleting source media.
func TestCollectionReferenceRemoval(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	path := `{"path":"/home/alice/photos/photo.jpg"}`
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", path, http.StatusOK)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/collections/"+id+"/files", path, http.StatusOK)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/collections/"+id+"/references", "", http.StatusOK)
	refs := mustGalleryJSON[struct {
		Items []collectionReference `json:"items"`
	}](t, response)
	if len(refs.Items) != 0 {
		t.Fatalf("remaining references = %+v", refs.Items)
	}
}

// TestCollectionDirectoryRemoval clears a previously added directory membership.
func TestCollectionDirectoryRemoval(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	path := `{"path":"/home/alice/photos"}`
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/directories", path, http.StatusOK)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/collections/"+id+"/directories", path, http.StatusOK)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/collections/"+id+"/references", "", http.StatusOK)
	refs := mustGalleryJSON[struct {
		Items []collectionReference `json:"items"`
	}](t, response)
	if len(refs.Items) != 0 {
		t.Fatalf("remaining references = %+v", refs.Items)
	}
}

// TestCollectionReferenceRemovalIsIdempotent permits retrying a completed removal.
func TestCollectionReferenceRemovalIsIdempotent(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	response := mustGalleryRequest(t, router, http.MethodDelete, "/api/collections/"+id+"/files", `{"path":"/home/alice/photos/photo.jpg"}`, http.StatusOK)
	ref := mustGalleryJSON[collectionReference](t, response)
	if ref.ID != "" {
		t.Fatalf("absent reference ID = %q", ref.ID)
	}
}

// TestAlbumRejectsCollectionReferences protects the album and collection boundary.
func TestAlbumRejectsCollectionReferences(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Photos", "")
	mustGalleryRequest(t, router, http.MethodGet, "/api/collections/"+id+"/references", "", http.StatusUnprocessableEntity)
}

// TestGalleryShareRemoval revokes a previously granted explicit member.
func TestGalleryShareRemoval(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPut, "/api/galleries/"+id+"/shares/bob", "", http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/galleries/"+id+"/shares/bob", "", http.StatusNoContent)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/shares", "", http.StatusOK)
	shares := mustGalleryJSON[struct {
		UserIDs []string `json:"userIds"`
	}](t, response)
	if len(shares.UserIDs) != 0 {
		t.Fatalf("remaining shares = %+v", shares.UserIDs)
	}
}

// TestGalleryDeleteRemovesMetadata checks the owner deletion route.
func TestGalleryDeleteRemovesMetadata(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/galleries/"+id, "", http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id, "", http.StatusNotFound)
}

// TestGalleryThumbnailRead streams the image stored for a visible gallery.
func TestGalleryThumbnailRead(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustUploadGalleryThumbnail(t, router, id)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/thumbnail/image", "", http.StatusOK)
	if !bytes.HasPrefix(response, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("thumbnail response was not PNG image bytes")
	}
}

// TestGalleryThumbnailDelete removes a published cover image.
func TestGalleryThumbnailDelete(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustUploadGalleryThumbnail(t, router, id)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/galleries/"+id+"/thumbnail/image", "", http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/thumbnail/image", "", http.StatusNotFound)
}

// TestGalleryThumbnailMissing reports an absent cover to a permitted reader.
func TestGalleryThumbnailMissing(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/thumbnail/image", "", http.StatusNotFound)
}

// TestGalleryDeleteRemovesCover removes the gallery's stored image asset with its metadata.
func TestGalleryDeleteRemovesCover(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustUploadGalleryThumbnail(t, router, id)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/galleries/"+id, "", http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id, "", http.StatusNotFound)
}

// TestCollectionReferenceAddIsIdempotent prevents duplicate source membership.
func TestCollectionReferenceAddIsIdempotent(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	path := `{"path":"/home/alice/photos/photo.jpg"}`
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", path, http.StatusOK)
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", path, http.StatusOK)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/collections/"+id+"/references", "", http.StatusOK)
	refs := mustGalleryJSON[struct {
		Items []collectionReference `json:"items"`
	}](t, response)
	if len(refs.Items) != 1 {
		t.Fatalf("duplicate references = %+v", refs.Items)
	}
}

// TestGalleryShareRejectsOwner prevents self membership in the share table.
func TestGalleryShareRejectsOwner(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPut, "/api/galleries/"+id+"/shares/alice", "", http.StatusUnprocessableEntity)
}

// TestScanHistoryListsCompletedJob exposes only the requesting owner's scan history.
func TestScanHistoryListsCompletedJob(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Photos", "")
	mustCompletedScan(t, router, id)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/scans", "", http.StatusOK)
	history := mustGalleryJSON[struct {
		Items []scanJob `json:"items"`
	}](t, response)
	if len(history.Items) != 1 || history.Items[0].Status != "COMPLETED" {
		t.Fatalf("scan history = %+v", history.Items)
	}
}

// TestScanCancellationAcceptsCompletedJob keeps cancellation idempotent after completion.
func TestScanCancellationAcceptsCompletedJob(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Photos", "")
	mustCompletedScan(t, router, id)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/scans", "", http.StatusOK)
	history := mustGalleryJSON[struct {
		Items []scanJob `json:"items"`
	}](t, response)
	if len(history.Items) != 1 {
		t.Fatalf("scan history = %+v", history.Items)
	}
	mustGalleryRequest(t, router, http.MethodDelete, "/api/scans/"+history.Items[0].ID, "", http.StatusOK)
}

// TestThumbnailNameAllowsOnlyIssuedImageNames rejects unsafe persisted asset names.
func TestThumbnailNameAllowsOnlyIssuedImageNames(t *testing.T) {
	for _, test := range []struct {
		name string
		want bool
	}{
		{name: uuid.NewString() + ".png", want: true},
		{name: uuid.NewString() + ".jpg", want: true},
		{name: "../avatar.png"},
		{name: "cover.gif"},
		{name: "not-a-uuid.png"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := thumbnailName(test.name); got != test.want {
				t.Fatalf("thumbnail name accepted = %t, want %t", got, test.want)
			}
		})
	}
}

// TestCollectionRejectsTraversalReference keeps caller paths inside their source scope.
func TestCollectionRejectsTraversalReference(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", `{"path":"/home/alice/photos/../private.jpg"}`, http.StatusUnprocessableEntity)
}

// TestAlbumRejectsAddedReference prevents collection mutation on album metadata.
func TestAlbumRejectsAddedReference(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Photos", "")
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", `{"path":"/home/alice/photos/photo.jpg"}`, http.StatusUnprocessableEntity)
}

// TestShareListIncludesMember checks owner-visible explicit sharing metadata.
func TestShareListIncludesMember(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPut, "/api/galleries/"+id+"/shares/bob", "", http.StatusNoContent)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/shares", "", http.StatusOK)
	shares := mustGalleryJSON[struct {
		UserIDs []string `json:"userIds"`
	}](t, response)
	if len(shares.UserIDs) != 1 || shares.UserIDs[0] != "bob" {
		t.Fatalf("shares = %+v", shares.UserIDs)
	}
}

// TestThumbnailUploadRejectsInvalidMedia validates declared type and source bytes.
func TestThumbnailUploadRejectsInvalidMedia(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	for _, test := range []struct {
		name, mediaType, body string
		want                  int
	}{
		{name: "empty image", mediaType: "image/png", want: http.StatusBadRequest},
		{name: "unsupported request type", mediaType: "text/plain", body: "text", want: http.StatusUnsupportedMediaType},
		{name: "content mismatch", mediaType: "image/jpeg", body: "text", want: http.StatusUnsupportedMediaType},
		{name: "unsupported image bytes", mediaType: "application/octet-stream", body: "text", want: http.StatusUnsupportedMediaType},
		{name: "malformed PNG", mediaType: "image/png", body: "\x89PNG\r\n\x1a\n", want: http.StatusUnprocessableEntity},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/api/galleries/"+id+"/thumbnail/image", bytes.NewBufferString(test.body))
			request.Header.Set("Authorization", "Bearer test")
			request.Header.Set("Content-Type", test.mediaType)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("thumbnail status = %d, want %d: %s", response.Code, test.want, response.Body.String())
			}
		})
	}
}

// TestThumbnailReplacementKeepsCurrentCover retires a previous uploaded asset.
func TestThumbnailReplacementKeepsCurrentCover(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustUploadGalleryThumbnail(t, router, id)
	mustUploadGalleryThumbnail(t, router, id)
	response := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/thumbnail/image", "", http.StatusOK)
	if !bytes.HasPrefix(response, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("replacement thumbnail was not PNG image bytes")
	}
}

// TestThumbnailDeleteWithoutAsset is safe for a gallery that has no cover yet.
func TestThumbnailDeleteWithoutAsset(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/galleries/"+id+"/thumbnail/image", "", http.StatusNoContent)
}

// TestShareByUsernameReportsUnavailableLookup distinguishes missing identity service.
func TestShareByUsernameReportsUnavailableLookup(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/shares", `{"username":"bob"}`, http.StatusServiceUnavailable)
}

// TestCollectionFileRejectsDirectory requires a file for file membership.
func TestCollectionFileRejectsDirectory(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/files", `{"path":"/home/alice/photos"}`, http.StatusUnprocessableEntity)
}

// TestCollectionDirectoryRejectsFile requires a directory for directory membership.
func TestCollectionDirectoryRejectsFile(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+id+"/directories", `{"path":"/home/alice/photos/photo.jpg"}`, http.StatusUnprocessableEntity)
}

// TestThumbnailUploadRejectsMismatchedImageType compares MIME declaration with image bytes.
func TestThumbnailUploadRejectsMismatchedImageType(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	var imageBytes bytes.Buffer
	if err := png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/api/galleries/"+id+"/thumbnail/image", &imageBytes)
	request.Header.Set("Authorization", "Bearer test")
	request.Header.Set("Content-Type", "image/jpeg")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("thumbnail status = %d, want 415", response.Code)
	}
}

// TestCollectionReferenceRejectsNonOwner prevents a reader from changing membership.
func TestCollectionReferenceRejectsNonOwner(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PUBLIC")
	request := httptest.NewRequest(http.MethodPost, "/api/collections/"+id+"/files", bytes.NewBufferString(`{"path":"/home/alice/photos/photo.jpg"}`))
	request.Header.Set("Authorization", "Bearer bob")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("reference mutation status = %d, want 403", response.Code)
	}
}

// TestGalleryDeleteRejectsNonOwner protects another account's gallery metadata.
func TestGalleryDeleteRejectsNonOwner(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PUBLIC")
	request := httptest.NewRequest(http.MethodDelete, "/api/galleries/"+id, nil)
	request.Header.Set("Authorization", "Bearer bob")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("gallery deletion status = %d, want 403", response.Code)
	}
}

// TestGalleryCreateRejectsDirectoryThumbnail requires an image file as the cover source.
func TestGalleryCreateRejectsDirectoryThumbnail(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries", `{"type":"COLLECTION","name":"Saved","thumbnailPath":"/home/alice/photos"}`, http.StatusUnprocessableEntity)
}
