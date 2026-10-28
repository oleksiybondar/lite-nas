package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"lite-nas/shared/browserauth"
	authcontract "lite-nas/shared/contracts/auth"
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

// TestCaptionGrantManifestSurvivesRestart preserves exact ordered source scope.
func TestCaptionGrantManifestSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "moments.db")
	store, err := openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	router := newRouterWithStore(authStub{session: browserauth.Session{UserID: "alice"}}, "", store)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	job := scanJob{ID: "job", GalleryID: galleryID, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if _, err := store.db.Exec(`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, "alice", "CAPTION", "RUNNING", job.CreatedAt); err != nil {
		t.Fatal(err)
	}
	manager := &scanManager{store: store}
	paths := []string{"/photos/lizard.jpg", "/photos/beetle.jpg"}
	if err := manager.saveCaptionGrant(context.Background(), job, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", paths); err != nil {
		t.Fatal(err)
	}
	if err := store.close(); err != nil {
		t.Fatal(err)
	}
	store, err = openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	reloaded, err := (&scanManager{store: store}).captionJobPaths(context.Background(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reloaded, paths) {
		t.Fatalf("reloaded caption paths = %v", reloaded)
	}
}

// TestCaptionJobResumesAtNextImage covers a service restart during a long scan.
func TestCaptionJobResumesAtNextImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "moments.db")
	store, err := openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	router := newRouterWithStore(authStub{session: browserauth.Session{UserID: "alice"}}, "", store)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	first, second := "/photos/first.jpg", "/photos/second.jpg"
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/photos/first.jpg','IMAGE'),('/photos/second.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+galleryID+`','/photos/first.jpg','IMAGE'),('`+galleryID+`','/photos/second.jpg','IMAGE')`,
	)
	job := scanJob{ID: "resume-job", GalleryID: galleryID, CreatedAt: timestamp()}
	if _, err := store.db.Exec(`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, "alice", "CAPTION", "RUNNING", job.CreatedAt); err != nil {
		t.Fatal(err)
	}
	grant := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ"
	if err := (&scanManager{store: store}).saveCaptionGrant(t.Context(), job, grant, []string{first, second}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE media_scan_jobs SET processed_items=1,successful_items=1 WHERE id=?`, job.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.close(); err != nil {
		t.Fatal(err)
	}
	store, err = openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	script := filepath.Join(t.TempDir(), "caption.py")
	if err := os.WriteFile(script, []byte("import sys\nprint('{\"ready\":true}', flush=True)\nfor line in sys.stdin: print('{\"caption\":\"A lizard\"}', flush=True)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	requested := make(chan string, 2)
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requested <- request.URL.Path
		return &http.Response{StatusCode: http.StatusOK, ContentLength: 5, Body: io.NopCloser(strings.NewReader("image"))}, nil
	})}}
	manager := newScanManager(store, authStub{}, files, &cityIndex{captionScript: script})
	manager.grants = subjectLookupFunc(func(_ context.Context, subject string, _ any, response any) error {
		if subject == authcontract.CaptionGrantRevokeRPCSubject {
			response.(*authcontract.CaptionGrantRevokeResponse).Revoked = true
		}
		return nil
	})
	if err := manager.resumeCaptionJobs(t.Context()); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(5 * time.Second)
	for {
		var status, remaining string
		var resumePending int
		if err := store.db.QueryRow(`SELECT status,coalesce(grant_token,''),resume_pending FROM media_scan_jobs WHERE id=?`, job.ID).Scan(&status, &remaining, &resumePending); err != nil {
			t.Fatal(err)
		}
		if status == "COMPLETED" && remaining == "" && resumePending == 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("resumed job status = %q, remaining grant = %t, resume flag = %d", status, remaining != "", resumePending)
		case <-time.After(10 * time.Millisecond):
		}
	}
	close(requested)
	var paths []string
	for item := range requested {
		paths = append(paths, item)
	}
	if !reflect.DeepEqual(paths, []string{second}) {
		t.Fatalf("resumed source requests = %v", paths)
	}
}

// mustPendingCaptionJob persists a resumable job for restart policy tests.
func mustPendingCaptionJob(t *testing.T, manager *scanManager, galleryID string, paths []string) scanJob {
	t.Helper()
	job := scanJob{ID: uuid.NewString(), GalleryID: galleryID, CreatedAt: timestamp()}
	if _, err := manager.store.db.Exec(`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, "alice", "CAPTION", "RUNNING", job.CreatedAt); err != nil {
		t.Fatal(err)
	}
	if err := manager.saveCaptionGrant(t.Context(), job, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", paths); err != nil {
		t.Fatal(err)
	}
	return job
}

// TestCaptionResumeFlagPreventsRestart keeps disabled jobs out of startup work.
func TestCaptionResumeFlagPreventsRestart(t *testing.T) {
	fixture, galleryID := mustCaptionScanFixture(t)
	job := mustPendingCaptionJob(t, fixture, galleryID, []string{"/home/alice/a.jpg"})
	if _, err := fixture.store.db.Exec(`UPDATE media_scan_jobs SET resume_pending=0 WHERE id=?`, job.ID); err != nil {
		t.Fatal(err)
	}
	manager := newScanManager(fixture.store, authStub{}, fileServiceClient{})
	if err := manager.resumeCaptionJobs(t.Context()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	var status string
	if err := fixture.store.db.QueryRow(`SELECT status FROM media_scan_jobs WHERE id=?`, job.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "RUNNING" {
		t.Fatalf("job without resume flag changed to %q", status)
	}
}

// TestCaptionProgressWaitsForCheckpoint avoids an SD write after each image.
func TestCaptionProgressWaitsForCheckpoint(t *testing.T) {
	fixture, galleryID := mustCaptionScanFixture(t)
	job := mustPendingCaptionJob(t, fixture, galleryID, []string{"/home/alice/a.jpg", "/home/alice/b.jpg"})
	script := filepath.Join(t.TempDir(), "caption.py")
	if err := os.WriteFile(script, []byte("import sys\nprint('{\"ready\":true}', flush=True)\nfor line in sys.stdin: print('{\"caption\":\"A lizard\"}', flush=True)\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	secondRequested := make(chan struct{})
	releaseSecond := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-releaseSecond:
		default:
			close(releaseSecond)
		}
	})
	files := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/home/alice/b.jpg" {
			close(secondRequested)
			<-releaseSecond
		}
		return &http.Response{StatusCode: http.StatusOK, ContentLength: 5, Body: io.NopCloser(strings.NewReader("image"))}, nil
	})}}
	manager := newScanManager(fixture.store, authStub{}, files, &cityIndex{captionScript: script})
	manager.grants = subjectLookupFunc(func(_ context.Context, _ string, _ any, _ any) error { return nil })
	if err := manager.resumeCaptionJobs(t.Context()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-secondRequested:
	case <-time.After(5 * time.Second):
		t.Fatal("second image was not requested")
	}
	var processed int
	if err := fixture.store.db.QueryRow(`SELECT processed_items FROM media_scan_jobs WHERE id=?`, job.ID).Scan(&processed); err != nil {
		t.Fatal(err)
	}
	close(releaseSecond)
	deadline := time.After(5 * time.Second)
	for {
		var status, grant string
		if err := fixture.store.db.QueryRow(`SELECT status,coalesce(grant_token,'') FROM media_scan_jobs WHERE id=?`, job.ID).Scan(&status, &grant); err != nil {
			t.Fatal(err)
		}
		if status == "COMPLETED" && grant == "" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("caption job did not complete: %s", status)
		case <-time.After(10 * time.Millisecond):
		}
	}
	if processed != 0 {
		t.Fatalf("progress before checkpoint = %d", processed)
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
