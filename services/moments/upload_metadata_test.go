package main

import (
	"net/http"
	"testing"
)

// TestUploadedMediaScansOneFile covers moments-backend/FR-009 incremental publication.
func TestUploadedMediaScansOneFile(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/media-metadata", `{"path":"/home/alice/photos/photo.jpg"}`, http.StatusOK)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+id+"/media-metadata", "", http.StatusOK)
	result := mustGalleryJSON[struct {
		Total int `json:"total"`
	}](t, body)
	if result.Total != 1 {
		t.Fatalf("uploaded media total = %d, want 1", result.Total)
	}
}

// TestUploadedMediaRejectsOutsideAlbum prevents scanning arbitrary authorized paths.
func TestUploadedMediaRejectsOutsideAlbum(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/media-metadata", `{"path":"/home/alice/other/photo.jpg"}`, http.StatusUnprocessableEntity)
}

// TestUploadedMediaLinksReferencingCollection reuses the one extracted path in an owned collection.
func TestUploadedMediaLinksReferencingCollection(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	album := mustNamedGallery(t, router, "ALBUM", "Holiday", "")
	collection := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+collection+"/directories", `{"path":"/home/alice/photos"}`, http.StatusOK)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+album+"/media-metadata", `{"path":"/home/alice/photos/photo.jpg"}`, http.StatusOK)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+collection+"/media-metadata", "", http.StatusOK)
	result := mustGalleryJSON[struct {
		Total int `json:"total"`
	}](t, body)
	if result.Total != 1 {
		t.Fatalf("collection media total = %d, want 1", result.Total)
	}
}

// TestAddingReferenceKeepsPublishedMedia covers non-destructive collection growth.
func TestAddingReferenceKeepsPublishedMedia(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	collection := mustNamedGallery(t, router, "COLLECTION", "Saved", "")
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+collection+"/files", `{"path":"/home/alice/photos/photo.jpg"}`, http.StatusOK)
	mustCompletedScan(t, router, collection)
	mustGalleryRequest(t, router, http.MethodPost, "/api/collections/"+collection+"/directories", `{"path":"/home/alice/photos"}`, http.StatusOK)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries/"+collection+"/media-metadata", "", http.StatusOK)
	result := mustGalleryJSON[struct {
		Total int `json:"total"`
	}](t, body)
	if result.Total != 1 {
		t.Fatalf("preserved media total = %d, want 1", result.Total)
	}
}
