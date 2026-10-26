package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/browserauth"
)

// roundTripFunc supplies controlled File Service responses without a listener.
type roundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements the HTTP transport used by the gallery fixture.
func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

// mustGalleryRouter prepares one isolated metadata store for route tests.
func mustGalleryRouter(t *testing.T, user string) http.Handler {
	return mustGalleryRouterWithAuth(t, authStub{session: browserauth.Session{UserID: user}})
}

// mustGalleryRouterWithAuth reuses the isolated transport and storage fixture.
func mustGalleryRouterWithAuth(t *testing.T, auth browserauth.AuthService) http.Handler {
	t.Helper()
	store, err := openGalleryStore(filepath.Join(t.TempDir(), "moments.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := `{"path":"/home/alice/photos","kind":"directory"}`
		if strings.HasPrefix(request.URL.Path, "/stat/") {
			path := strings.TrimPrefix(request.URL.Path, "/stat")
			kind := "directory"
			if strings.HasSuffix(path, ".jpg") {
				kind = "file"
			}
			body = `{"path":` + strconv.Quote(path) + `,"kind":` + strconv.Quote(kind) + `}`
		}
		if strings.HasPrefix(request.URL.Path, "/ls/") {
			body = `[{"path":"/home/alice/photos/photo.jpg","kind":"file"}]`
		}
		if strings.HasPrefix(request.URL.Path, "/metadata/") {
			body = `{"path":"/home/alice/photos/photo.jpg","mediaType":"IMAGE","contentType":"image/jpeg","tags":[]}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	thumbnailDir := filepath.Join(t.TempDir(), "thumbnails")
	if err := os.MkdirAll(thumbnailDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return newRouterWithRuntime(auth, "", store, files, thumbnailDir)
}

// actorAuth derives controlled test identities from validated token text.
type actorAuth struct{ authStub }

// Me returns a controlled principal for gallery visibility boundary tests.
func (actorAuth) Me(_ time.Time, token string) (browserauth.Session, error) {
	if token != "alice" && token != "bob" {
		return browserauth.Session{}, browserauth.ErrUnauthorized
	}
	return browserauth.Session{UserID: token}, nil
}

// mustCreateCollection executes the expected successful setup action.
func mustCreateCollection(t *testing.T, router http.Handler) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/galleries", strings.NewReader(`{"type":"COLLECTION","name":"Saved"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "test"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("create status = %d: %s", response.Code, response.Body.String())
	}
	var gallery struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &gallery); err != nil {
		t.Fatal(err)
	}
	return gallery.ID
}

// mustCreateCollectionAs prepares a gallery with one controlled test principal.
func mustCreateCollectionAs(t *testing.T, router http.Handler, owner, visibility string) string {
	t.Helper()
	body := `{"type":"COLLECTION","name":"Saved","visibility":"` + visibility + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/galleries", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+owner)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("create status = %d: %s", response.Code, response.Body.String())
	}
	var gallery struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &gallery); err != nil {
		t.Fatal(err)
	}
	return gallery.ID
}

// TestPrivateGalleryIsHidden covers moments-backend/FR-002.
func TestPrivateGalleryIsHidden(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PRIVATE")
	request := httptest.NewRequest(http.MethodGet, "/api/galleries/"+id, nil)
	request.Header.Set("Authorization", "Bearer bob")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

// TestPublicGalleryRequiresAuthentication covers moments-backend/FR-002.
func TestPublicGalleryRequiresAuthentication(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PUBLIC")
	request := httptest.NewRequest(http.MethodGet, "/api/galleries/"+id, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

// TestPublicGalleryVisibleToAuthenticatedUser covers moments-backend/FR-002.
func TestPublicGalleryVisibleToAuthenticatedUser(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PUBLIC")
	request := httptest.NewRequest(http.MethodGet, "/api/galleries/"+id, nil)
	request.Header.Set("Authorization", "Bearer bob")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
}

// TestGalleryListRejectsDuplicateFilter covers moments-backend/FR-005.
func TestGalleryListRejectsDuplicateFilter(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	request := httptest.NewRequest(http.MethodGet, "/api/galleries?type=ALBUM&type=ALBUM", nil)
	request.Header.Set("Authorization", "Bearer alice")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
}

// TestSharedGalleryVisibleToMember covers moments-backend/FR-002.
func TestSharedGalleryVisibleToMember(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PRIVATE")
	share := httptest.NewRequest(http.MethodPut, "/api/galleries/"+id+"/shares/bob", nil)
	share.Header.Set("Authorization", "Bearer alice")
	shared := httptest.NewRecorder()
	router.ServeHTTP(shared, share)
	if shared.Code != http.StatusNoContent {
		t.Fatalf("share status = %d: %s", shared.Code, shared.Body.String())
	}
	request := httptest.NewRequest(http.MethodGet, "/api/galleries/"+id, nil)
	request.Header.Set("Authorization", "Bearer bob")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("member status = %d: %s", response.Code, response.Body.String())
	}
}

// TestScanStatusIsSameUserOnly covers moments-backend/FR-004.
func TestScanStatusIsSameUserOnly(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PUBLIC")
	start := httptest.NewRequest(http.MethodPost, "/api/galleries/"+id+"/scans", nil)
	start.Header.Set("Authorization", "Bearer alice")
	started := httptest.NewRecorder()
	router.ServeHTTP(started, start)
	if started.Code != http.StatusAccepted {
		t.Fatalf("start status = %d: %s", started.Code, started.Body.String())
	}
	var job scanJob
	if err := json.Unmarshal(started.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/scans/"+job.ID, nil)
	request.Header.Set("Authorization", "Bearer bob")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

// TestThumbnailUploadStoresImage covers moments-backend/FR-006.
func TestThumbnailUploadStoresImage(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	bitmap := image.NewRGBA(image.Rect(0, 0, 1, 1))
	bitmap.Set(0, 0, color.RGBA{R: 255, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, bitmap); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPut, "/api/galleries/"+id+"/thumbnail/image", &data)
	request.Header.Set("Content-Type", "image/png")
	request.Header.Set("Authorization", "Bearer test")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("upload status = %d: %s", response.Code, response.Body.String())
	}
}

// TestThumbnailGenerationRoundTripsDataURL covers moments-backend/FR-006.
func TestThumbnailGenerationRoundTripsDataURL(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	recipe := `{"image":"data:image/png;base64,AAAA"}`
	create := httptest.NewRequest(http.MethodPost, "/api/galleries", strings.NewReader(`{"type":"COLLECTION","name":"Saved","thumbnailGeneration":`+recipe+`}`))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Authorization", "Bearer test")
	created := httptest.NewRecorder()
	router.ServeHTTP(created, create)
	if created.Code != http.StatusOK {
		t.Fatalf("create status = %d: %s", created.Code, created.Body.String())
	}
	var gallery struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &gallery); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/galleries/"+gallery.ID, nil)
	request.Header.Set("Authorization", "Bearer test")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	var body struct {
		ThumbnailGeneration json.RawMessage `json:"thumbnailGeneration"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if string(body.ThumbnailGeneration) != recipe {
		t.Fatalf("recipe = %s", body.ThumbnailGeneration)
	}
}

// TestAlbumScanPublishesMetadata covers moments-backend/FR-004.
func TestAlbumScanPublishesMetadata(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	create := httptest.NewRequest(http.MethodPost, "/api/galleries", strings.NewReader(`{"type":"ALBUM","name":"Summer","uploadDirectory":"/home/alice/photos"}`))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Authorization", "Bearer test")
	created := httptest.NewRecorder()
	router.ServeHTTP(created, create)
	if created.Code != http.StatusOK {
		t.Fatalf("create status = %d: %s", created.Code, created.Body.String())
	}
	var gallery struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &gallery); err != nil {
		t.Fatal(err)
	}
	start := httptest.NewRequest(http.MethodPost, "/api/galleries/"+gallery.ID+"/scans", nil)
	start.Header.Set("Authorization", "Bearer test")
	started := httptest.NewRecorder()
	router.ServeHTTP(started, start)
	if started.Code != http.StatusAccepted {
		t.Fatalf("start status = %d: %s", started.Code, started.Body.String())
	}
	var job scanJob
	if err := json.Unmarshal(started.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	completed := false
	for attempts := 0; attempts < 100; attempts++ {
		poll := httptest.NewRequest(http.MethodGet, "/api/scans/"+job.ID, nil)
		poll.Header.Set("Authorization", "Bearer test")
		status := httptest.NewRecorder()
		router.ServeHTTP(status, poll)
		if status.Code != http.StatusOK {
			t.Fatalf("poll status = %d: %s", status.Code, status.Body.String())
		}
		if err := json.Unmarshal(status.Body.Bytes(), &job); err != nil {
			t.Fatal(err)
		}
		if job.Status == "COMPLETED" {
			completed = true
			break
		}
		if job.Status == "FAILED" || job.Status == "PARTIAL" || job.Status == "CANCELLED" {
			t.Fatalf("scan status = %s: %s", job.Status, status.Body.String())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !completed {
		t.Fatal("scan did not complete")
	}
	resultRequest := httptest.NewRequest(http.MethodGet, "/api/galleries/"+gallery.ID+"/media-metadata", nil)
	resultRequest.Header.Set("Authorization", "Bearer test")
	resultResponse := httptest.NewRecorder()
	router.ServeHTTP(resultResponse, resultRequest)
	if resultResponse.Code != http.StatusOK {
		t.Fatalf("result status = %d: %s", resultResponse.Code, resultResponse.Body.String())
	}
	var result struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(resultResponse.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Fatalf("published media total = %d, want 1", result.Total)
	}
}

// TestCreateAlbumPersistsMetadata covers moments-backend/FR-001 and FR-003.
func TestCreateAlbumPersistsMetadata(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	request := httptest.NewRequest(http.MethodPost, "/api/galleries", strings.NewReader(`{"type":"ALBUM","name":"Summer","uploadDirectory":"/home/alice/photos"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "test"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
}
