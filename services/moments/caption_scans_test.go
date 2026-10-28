package main

import (
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
)

// TestCaptionScanRouteCreatesCaptionJob covers the owner-only caption action.
func TestCaptionScanRouteCreatesCaptionJob(t *testing.T) {
	router := mustGalleryRouter(t, "alice")
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	body := mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/caption-scans", `{"force":false}`, http.StatusAccepted)
	job := mustGalleryJSON[scanJob](t, body)
	if job.Kind != "CAPTION" {
		t.Fatalf("job kind = %q", job.Kind)
	}
}

// mustCaptionScanFixture prepares two linked images and one saved caption.
func mustCaptionScanFixture(t *testing.T) (*scanManager, string) {
	t.Helper()
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/a.jpg','IMAGE'),('/home/alice/b.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/a.jpg','IMAGE'),('`+id+`','/home/alice/b.jpg','IMAGE')`,
		`INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES('/home/alice/a.jpg','Old caption','old caption','test','now')`,
	)
	return &scanManager{store: store}, id
}

// TestCaptionScanSkipsSavedCaptions covers the default lightweight caption run.
func TestCaptionScanSkipsSavedCaptions(t *testing.T) {
	manager, id := mustCaptionScanFixture(t)
	paths, err := manager.captionPaths(t.Context(), id, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{"/home/alice/b.jpg"}) {
		t.Fatalf("caption paths = %v", paths)
	}
}

// TestCaptionScanForceIncludesSavedCaptions covers explicit regeneration.
func TestCaptionScanForceIncludesSavedCaptions(t *testing.T) {
	manager, id := mustCaptionScanFixture(t)
	paths, err := manager.captionPaths(t.Context(), id, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{"/home/alice/a.jpg", "/home/alice/b.jpg"}) {
		t.Fatalf("forced caption paths = %v", paths)
	}
}

// TestScanJobKindMigrationKeepsLegacyJobs covers existing installed databases.
func TestScanJobKindMigrationKeepsLegacyJobs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "moments.db")
	store, err := openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`DROP TABLE media_scan_jobs`,
		`CREATE TABLE media_scan_jobs (id TEXT PRIMARY KEY,gallery_id TEXT,requested_by TEXT,status TEXT,total_items INTEGER,processed_items INTEGER NOT NULL DEFAULT 0,successful_items INTEGER NOT NULL DEFAULT 0,failed_items INTEGER NOT NULL DEFAULT 0,created_at TEXT,started_at TEXT,finished_at TEXT,error_summary TEXT)`,
		`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,status,created_at) VALUES('old','gallery','alice','COMPLETED','now')`,
	} {
		if _, err := store.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.close(); err != nil {
		t.Fatal(err)
	}
	store, err = openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	var kind string
	if err := store.db.QueryRow(`SELECT kind FROM media_scan_jobs WHERE id='old'`).Scan(&kind); err != nil {
		t.Fatal(err)
	}
	if kind != "MEDIA" {
		t.Fatalf("legacy job kind = %q", kind)
	}
}
