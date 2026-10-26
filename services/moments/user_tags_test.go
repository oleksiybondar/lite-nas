package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// mustGalleryRequest performs an expected successful route action for tag tests.
func mustGalleryRequest(t *testing.T, router http.Handler, method, path, body string, status int) []byte {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer test")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != status {
		t.Fatalf("%s %s status = %d, want %d: %s", method, path, response.Code, status, response.Body.String())
	}
	return response.Body.Bytes()
}

// mustGalleryJSON decodes a successful fixture response.
func mustGalleryJSON[T any](t *testing.T, body []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

// mustNamedGallery creates an album or collection with controlled searchable fields.
func mustNamedGallery(t *testing.T, router http.Handler, kind, name, description string) string {
	t.Helper()
	input := map[string]string{"type": kind, "name": name, "description": description}
	if kind == "ALBUM" {
		input["uploadDirectory"] = "/home/alice/photos"
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	created := mustGalleryRequest(t, router, http.MethodPost, "/api/galleries", string(encoded), http.StatusOK)
	return mustGalleryJSON[struct {
		ID string `json:"id"`
	}](t, created).ID
}

// mustUserTag creates a user tag as a successful test setup action.
func mustUserTag(t *testing.T, router http.Handler, galleryID, value string) userTagBody {
	t.Helper()
	encoded, err := json.Marshal(map[string]string{"value": value})
	if err != nil {
		t.Fatal(err)
	}
	response := mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+galleryID+"/user-tags", string(encoded), http.StatusOK)
	return mustGalleryJSON[userTagBody](t, response)
}

// mustCompletedScan waits for a successful scan before examining published tags.
func mustCompletedScan(t *testing.T, router http.Handler, galleryID string) {
	t.Helper()
	started := mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+galleryID+"/scans", "", http.StatusAccepted)
	job := mustGalleryJSON[scanJob](t, started)
	for attempts := 0; attempts < 100; attempts++ {
		body := mustGalleryRequest(t, router, http.MethodGet, "/api/scans/"+job.ID, "", http.StatusOK)
		job = mustGalleryJSON[scanJob](t, body)
		if job.Status == "COMPLETED" {
			return
		}
		if job.Status == "FAILED" || job.Status == "PARTIAL" || job.Status == "CANCELLED" {
			t.Fatalf("scan ended with %s", job.Status)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("scan did not complete")
}

// TestUserTagCRUD covers moments-backend/FR-007.
func TestUserTagCRUD(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	tag := mustUserTag(t, router, id, "Family")
	mustGalleryRequest(t, router, http.MethodPatch, "/api/galleries/"+id+"/user-tags/"+tag.ID, `{"value":"Friends"}`, http.StatusOK)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id, "", http.StatusOK)
	gallery := mustGalleryJSON[struct {
		Tags []galleryTag `json:"tags"`
	}](t, body)
	if len(gallery.Tags) != 1 || gallery.Tags[0].Type != "USER" || gallery.Tags[0].Value != "Friends" {
		t.Fatalf("gallery tags = %+v", gallery.Tags)
	}
}

// TestUserTagReadByID covers moments-backend/FR-007.
func TestUserTagReadByID(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	tag := mustUserTag(t, router, id, "Family")
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/user-tags/"+tag.ID, "", http.StatusOK)
	loaded := mustGalleryJSON[userTagBody](t, body)
	if loaded.ID != tag.ID || loaded.Type != "USER" || loaded.Value != "Family" {
		t.Fatalf("loaded user tag = %+v", loaded)
	}
}

// TestUserTagDelete covers moments-backend/FR-007.
func TestUserTagDelete(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	tag := mustUserTag(t, router, id, "Family")
	mustGalleryRequest(t, router, http.MethodDelete, "/api/galleries/"+id+"/user-tags/"+tag.ID, "", http.StatusNoContent)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/user-tags", "", http.StatusOK)
	tags := mustGalleryJSON[struct {
		Items []userTagBody `json:"items"`
	}](t, body)
	if len(tags.Items) != 0 {
		t.Fatalf("user tags after deletion = %+v", tags.Items)
	}
}

// TestUserTagDuplicateIsRejected covers moments-backend/FR-007.
func TestUserTagDuplicateIsRejected(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	mustUserTag(t, router, id, "Family")
	request := httptest.NewRequest(http.MethodPost, "/api/galleries/"+id+"/user-tags", strings.NewReader(`{"value":" family "}`))
	request.Header.Set("Authorization", "Bearer test")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d: %s", response.Code, response.Body.String())
	}
}

// TestGallerySearchAndPagination covers moments-backend/FR-005 and FR-007.
func TestGallerySearchAndPagination(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	album := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	collection := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	mustUserTag(t, router, album, "Travel")
	mustUserTag(t, router, collection, "Travel")
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=travel&page=2&size=1", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Items    []galleryBody     `json:"items"`
		Metadata galleryPagination `json:"metadata"`
	}](t, body)
	if len(page.Items) != 1 || len(page.Items[0].Tags) != 1 || page.Items[0].Tags[0].Value != "Travel" || page.Metadata.Page != 2 || page.Metadata.Size != 1 || page.Metadata.TotalCount != 2 || page.Metadata.TotalPages != 2 {
		t.Fatalf("gallery page = %+v", page)
	}
}

// TestGallerySearchMatchesDescriptions covers moments-backend/FR-005.
func TestGallerySearchMatchesDescriptions(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	mustNamedGallery(t, router, "COLLECTION", "Saved", "Mountain weekend")
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=mountain", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Metadata galleryPagination `json:"metadata"`
	}](t, body)
	if page.Metadata.TotalCount != 1 {
		t.Fatalf("search total = %d, want 1", page.Metadata.TotalCount)
	}
}

// TestUserTagSurvivesScan covers moments-backend/FR-007.
func TestUserTagSurvivesScan(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	mustUserTag(t, router, id, "Travel")
	mustCompletedScan(t, router, id)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/tags", "", http.StatusOK)
	tags := mustGalleryJSON[struct {
		Items []galleryTag `json:"items"`
	}](t, body)
	found := false
	for _, tag := range tags.Items {
		if tag.Type == "USER" && tag.Value == "Travel" {
			found = true
		}
	}
	if !found {
		t.Fatalf("user tag missing after scan: %+v", tags.Items)
	}
}

// TestGallerySearchMatchesScanTags covers moments-backend/FR-005.
func TestGallerySearchMatchesScanTags(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	mustCompletedScan(t, router, id)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q="+url.QueryEscape("IMAGE"), "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Metadata galleryPagination `json:"metadata"`
	}](t, body)
	if page.Metadata.TotalCount != 1 {
		t.Fatalf("scan tag search total = %d, want 1", page.Metadata.TotalCount)
	}
}

// TestGallerySearchMatchesFileTypeTags covers moments-backend/FR-005.
func TestGallerySearchMatchesFileTypeTags(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	mustCompletedScan(t, router, id)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=JPG", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Metadata galleryPagination `json:"metadata"`
	}](t, body)
	if page.Metadata.TotalCount != 1 {
		t.Fatalf("file type search total = %d, want 1", page.Metadata.TotalCount)
	}
}

// TestGallerySearchTypeFilter covers moments-backend/FR-005.
func TestGallerySearchTypeFilter(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	album := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	collection := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	mustUserTag(t, router, album, "Travel")
	mustUserTag(t, router, collection, "Travel")
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=travel&type=COLLECTION", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Items []galleryBody `json:"items"`
	}](t, body)
	if len(page.Items) != 1 || page.Items[0].Type != "COLLECTION" {
		t.Fatalf("filtered galleries = %+v", page.Items)
	}
}

// TestUserTagMutationRequiresOwner covers moments-backend/FR-002 and FR-007.
func TestUserTagMutationRequiresOwner(t *testing.T) {
	router := mustGalleryRouterWithAuth(t, actorAuth{})
	id := mustCreateCollectionAs(t, router, "alice", "PUBLIC")
	request := httptest.NewRequest(http.MethodPost, "/api/galleries/"+id+"/user-tags", strings.NewReader(`{"value":"Family"}`))
	request.Header.Set("Authorization", "Bearer bob")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("non-owner status = %d, want 403", response.Code)
	}
}

// TestGalleryPaginationRejectsInvalidPage covers moments-backend/FR-005.
func TestGalleryPaginationRejectsInvalidPage(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	request := httptest.NewRequest(http.MethodGet, "/api/galleries?page=0", nil)
	request.Header.Set("Authorization", "Bearer test")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid page status = %d, want 422", response.Code)
	}
}

// TestUserTagRoutesAppearInOpenAPI covers moments-backend/IR-002 and FR-007.
func TestUserTagRoutesAppearInOpenAPI(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	request := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "/api/galleries/{galleryId}/user-tags/{tagId}") {
		t.Fatalf("user tag OpenAPI path missing: %d", response.Code)
	}
}
