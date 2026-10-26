package main

import (
	"bytes"
	"context"
	"database/sql"
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
	rbaccontract "lite-nas/shared/contracts/rbac"
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
	return mustGalleryRouterWithLookup(t, auth, nil)
}

// mustGalleryRouterWithLookup supplies an optional RBAC lookup for share tests.
func mustGalleryRouterWithLookup(t *testing.T, auth browserauth.AuthService, users subjectLookup) http.Handler {
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
	return newRouterWithRuntime(auth, "", store, files, thumbnailDir, "/home", users)
}

// subjectLookupFunc resolves one controlled login during gallery route tests.
type subjectLookupFunc func(context.Context, string, any, any) error

// Request implements the narrow RBAC request boundary in galleryController.
func (fn subjectLookupFunc) Request(ctx context.Context, subject string, request any, response any) error {
	return fn(ctx, subject, request, response)
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

// TestShareByUsernameStoresResolvedUID covers a login entered in the owner form.
func TestShareByUsernameStoresResolvedUID(t *testing.T) {
	lookup := subjectLookupFunc(func(_ context.Context, subject string, request any, response any) error {
		if subject != rbaccontract.GetSubjectRolesRPCSubject || request.(rbaccontract.GetSubjectRolesRequest).Username != "oleksiybondar" {
			t.Fatal("unexpected RBAC lookup")
		}
		response.(*rbaccontract.GetSubjectRolesResponse).UID = "1002"
		return nil
	})
	router := mustGalleryRouterWithLookup(t, authStub{session: browserauth.Session{UserID: "1001"}}, lookup)
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/shares", `{"username":"oleksiybondar"}`, http.StatusOK)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/shares", "", http.StatusOK)
	shares := mustGalleryJSON[struct {
		UserIDs []string `json:"userIds"`
	}](t, body)
	if len(shares.UserIDs) != 1 || shares.UserIDs[0] != "1002" {
		t.Fatalf("shared user IDs = %+v", shares.UserIDs)
	}
}

// TestShareByUsernameRejectsOptionOperand keeps an option-shaped login out of the RBAC command.
func TestShareByUsernameRejectsOptionOperand(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/shares", `{"username":"--help"}`, http.StatusUnprocessableEntity)
}

// TestShareByUsernameRejectsMissingAccount covers an unresolved host login.
func TestShareByUsernameRejectsMissingAccount(t *testing.T) {
	lookup := subjectLookupFunc(func(_ context.Context, _ string, _ any, _ any) error { return nil })
	router := mustGalleryRouterWithLookup(t, authStub{session: browserauth.Session{UserID: "1001"}}, lookup)
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/shares", `{"username":"missing"}`, http.StatusNotFound)
}

// TestGalleryStoreMigratesThumbnailPath preserves deployed databases created before cover paths.
func TestGalleryStoreMigratesThumbnailPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE galleries (
		id TEXT PRIMARY KEY, type TEXT, name TEXT, description TEXT, owner_id TEXT,
		visibility TEXT, upload_directory TEXT, wall_strategy TEXT, thumbnail_json TEXT,
		thumbnail_asset TEXT, created_at TEXT, updated_at TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	var value sql.NullString
	if err := store.db.QueryRow(`SELECT thumbnail_path FROM galleries LIMIT 1`).Scan(&value); err != sql.ErrNoRows {
		t.Fatalf("thumbnail path migration query = %v, want no rows", err)
	}
}

// TestThumbnailPathRoundTrips verifies owner-selected NAS image paths in gallery metadata.
func TestThumbnailPathRoundTrips(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	id := mustCreateCollection(t, router)
	body := mustGalleryRequest(t, router, http.MethodPatch, "/api/galleries/"+id,
		`{"thumbnailPath":"/home/alice/photos/photo.jpg"}`, http.StatusOK)
	updated := mustGalleryJSON[struct {
		ThumbnailPath string `json:"thumbnailPath"`
	}](t, body)
	if updated.ThumbnailPath != "/home/alice/photos/photo.jpg" {
		t.Fatalf("thumbnail path = %q", updated.ThumbnailPath)
	}
}

// Requirements: moments-backend/FR-003. A selected layout is stored with each gallery type.
func TestGalleryCreatePersistsSelectedLayout(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "album", body: `{"type":"ALBUM","name":"Photos","uploadDirectory":"/home/alice/photos","wallRenderStrategy":"grid"}`},
		{name: "collection", body: `{"type":"COLLECTION","name":"Saved","wallRenderStrategy":"grid"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := mustGalleryRouter(t, "alice")
			body := mustGalleryRequest(t, router, http.MethodPost, "/api/galleries", test.body, http.StatusOK)
			gallery := mustGalleryJSON[struct {
				WallRenderStrategy string `json:"wallRenderStrategy"`
			}](t, body)
			if gallery.WallRenderStrategy != "grid" {
				t.Fatalf("layout = %q, want grid", gallery.WallRenderStrategy)
			}
		})
	}
}

// Requirements: moments-backend/FR-003. An owner's layout edit survives a later read.
func TestGalleryPatchPersistsSelectedLayout(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustCreateCollection(t, router)
	mustGalleryRequest(t, router, http.MethodPatch, "/api/galleries/"+id,
		`{"wallRenderStrategy":"masonry"}`, http.StatusOK)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id, "", http.StatusOK)
	gallery := mustGalleryJSON[struct {
		WallRenderStrategy string `json:"wallRenderStrategy"`
	}](t, body)
	if gallery.WallRenderStrategy != "masonry" {
		t.Fatalf("layout = %q, want masonry", gallery.WallRenderStrategy)
	}
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

// TestGalleryCreateDefaultsPublic covers the omitted visibility on both gallery types.
func TestGalleryCreateDefaultsPublic(t *testing.T) {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "album", body: `{"type":"ALBUM","name":"Photos","uploadDirectory":"/home/alice/photos"}`},
		{name: "collection", body: `{"type":"COLLECTION","name":"Saved"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := mustGalleryRouter(t, "1001")
			body := mustGalleryRequest(t, router, http.MethodPost, "/api/galleries", test.body, http.StatusOK)
			created := mustGalleryJSON[struct {
				Visibility string `json:"visibility"`
			}](t, body)
			if created.Visibility != "PUBLIC" {
				t.Fatalf("visibility = %q, want PUBLIC", created.Visibility)
			}
		})
	}
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
