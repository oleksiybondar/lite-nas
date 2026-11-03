package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestConfirmDeletedChecksPreviouslyLinkedPath protects source deletion scope.
func TestConfirmDeletedChecksPreviouslyLinkedPath(t *testing.T) {
	store, router := mustSearchFixture(t)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+galleryID+`','/home/alice/photo.jpg','IMAGE')`,
	)
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, ".jpg") {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"path":"/home/alice","kind":"directory"}`))}, nil
	})}}
	manager := &scanManager{store: store, files: files, active: map[string]*activeScan{galleryID: {jobID: "job", token: "test"}}}
	deleted, err := manager.confirmDeleted(t.Context(), scanJob{ID: "job", GalleryID: galleryID}, nil)
	if err != nil || len(deleted) != 1 || deleted[0] != "/home/alice/photo.jpg" {
		t.Fatalf("confirmed deletion = %v, error = %v", deleted, err)
	}
}

// TestCollectionDiscoveryResolvesFileReference checks authorized source identity.
func TestCollectionDiscoveryResolvesFileReference(t *testing.T) {
	store, router := mustSearchFixture(t)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store, `INSERT INTO gallery_references(id,gallery_id,reference_type,path,created_at) VALUES('ref', '`+galleryID+`', 'FILE', '/home/alice/photo.jpg', 'now')`)
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"path":"/home/alice/photo.jpg","kind":"file"}`))}, nil
	})}}
	manager := &scanManager{store: store, files: files, active: map[string]*activeScan{galleryID: {jobID: "job", token: "test"}}}
	paths, err := manager.discover(t.Context(), scanJob{ID: "job", GalleryID: galleryID}, galleryRecord{ID: galleryID, Type: "COLLECTION"})
	if err != nil || len(paths) != 1 || paths[0] != "/home/alice/photo.jpg" {
		t.Fatalf("discovered paths = %v, error = %v", paths, err)
	}
}

// TestCollectionDiscoveryExpandsDirectoryReference checks recursive source listing.
func TestCollectionDiscoveryExpandsDirectoryReference(t *testing.T) {
	store, router := mustSearchFixture(t)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store, `INSERT INTO gallery_references(id,gallery_id,reference_type,path,created_at) VALUES('ref', '`+galleryID+`', 'DIRECTORY', '/home/alice', 'now')`)
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`[{"path":"/home/alice/photo.jpg","kind":"file"}]`))}, nil
	})}}
	manager := &scanManager{store: store, files: files, active: map[string]*activeScan{galleryID: {jobID: "job", token: "test"}}}
	paths, err := manager.discover(t.Context(), scanJob{ID: "job", GalleryID: galleryID}, galleryRecord{ID: galleryID, Type: "COLLECTION"})
	if err != nil || len(paths) != 1 || paths[0] != "/home/alice/photo.jpg" {
		t.Fatalf("discovered paths = %v, error = %v", paths, err)
	}
}

// TestCollectionDiscoverySkipsMissingFile keeps removed sources out of scans.
func TestCollectionDiscoverySkipsMissingFile(t *testing.T) {
	store, router := mustSearchFixture(t)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store, `INSERT INTO gallery_references(id,gallery_id,reference_type,path,created_at) VALUES('ref', '`+galleryID+`', 'FILE', '/home/alice/removed.jpg', 'now')`)
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, ".jpg") {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"path":"/home/alice","kind":"directory"}`))}, nil
	})}}
	manager := &scanManager{store: store, files: files, active: map[string]*activeScan{galleryID: {jobID: "job", token: "test"}}}
	paths, err := manager.discover(t.Context(), scanJob{ID: "job", GalleryID: galleryID}, galleryRecord{ID: galleryID, Type: "COLLECTION"})
	if err != nil || len(paths) != 0 {
		t.Fatalf("discovered removed file = %v, error = %v", paths, err)
	}
}
