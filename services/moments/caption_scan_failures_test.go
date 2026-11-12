package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authcontract "lite-nas/shared/contracts/auth"
)

// mustCaptionRunJob persists one caption job with a worker seat for synchronous tests.
func mustCaptionRunJob(t *testing.T, manager *scanManager, galleryID string) scanJob {
	t.Helper()
	job := scanJob{ID: uuid.NewString(), GalleryID: galleryID, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if _, err := manager.store.db.Exec(`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, "alice", "CAPTION", "PENDING", job.CreatedAt); err != nil {
		t.Fatalf("create caption job: %v", err)
	}
	manager.seats = make(chan struct{}, 1)
	manager.seats <- struct{}{}
	return job
}

// mustCaptionJobStatus reads the durable outcome of a completed worker.
func mustCaptionJobStatus(t *testing.T, manager *scanManager, jobID string) string {
	t.Helper()
	var status string
	if err := manager.store.db.QueryRow(`SELECT status FROM media_scan_jobs WHERE id=?`, jobID).Scan(&status); err != nil {
		t.Fatalf("read caption job: %v", err)
	}
	return status
}

// TestCaptionRunRejectsChangedGalleryOwner protects ownership at worker time.
func TestCaptionRunRejectsChangedGalleryOwner(t *testing.T) {
	manager, galleryID := mustCaptionScanFixture(t)
	job := mustCaptionRunJob(t, manager, galleryID)
	manager.runCaptionScan(t.Context(), job, "bob", false, false)
	if status := mustCaptionJobStatus(t, manager, job.ID); status != "FAILED" {
		t.Fatalf("caption job status = %q", status)
	}
}

// TestCaptionRunCompletesEmptyGallery avoids a needless model launch.
func TestCaptionRunCompletesEmptyGallery(t *testing.T) {
	store, router := mustSearchFixture(t)
	galleryID := mustNamedGallery(t, router, "COLLECTION", "Empty", "")
	manager := &scanManager{store: store}
	job := mustCaptionRunJob(t, manager, galleryID)
	manager.runCaptionScan(t.Context(), job, "alice", false, false)
	if status := mustCaptionJobStatus(t, manager, job.ID); status != "COMPLETED" {
		t.Fatalf("caption job status = %q", status)
	}
}

// TestCaptionRunFailsWithoutGrantService protects source image reads.
func TestCaptionRunFailsWithoutGrantService(t *testing.T) {
	manager, galleryID := mustCaptionScanFixture(t)
	job := mustCaptionRunJob(t, manager, galleryID)
	manager.runCaptionScan(t.Context(), job, "alice", false, false)
	if status := mustCaptionJobStatus(t, manager, job.ID); status != "FAILED" {
		t.Fatalf("caption job status = %q", status)
	}
}

// TestCaptionRunResumeFailsWithoutSavedGrant prevents unauthorized restart.
func TestCaptionRunResumeFailsWithoutSavedGrant(t *testing.T) {
	manager, galleryID := mustCaptionScanFixture(t)
	job := mustCaptionRunJob(t, manager, galleryID)
	manager.runCaptionScan(t.Context(), job, "alice", false, true)
	if status := mustCaptionJobStatus(t, manager, job.ID); status != "FAILED" {
		t.Fatalf("resumed caption job status = %q", status)
	}
}

// TestExpiredCaptionJobFailsOnResume keeps an old grant from restarting work.
func TestExpiredCaptionJobFailsOnResume(t *testing.T) {
	manager, galleryID := mustCaptionScanFixture(t)
	job := mustPendingCaptionJob(t, manager, galleryID, []string{"/home/alice/a.jpg"})
	old := time.Now().Add(-3 * captionJobLifetime).UTC().Format(time.RFC3339Nano)
	if _, err := manager.store.db.Exec(`UPDATE media_scan_jobs SET created_at=? WHERE id=?`, old, job.ID); err != nil {
		t.Fatalf("age caption job: %v", err)
	}
	if err := manager.resumeCaptionJobs(t.Context()); err != nil {
		t.Fatalf("resume expired job: %v", err)
	}
	if status := mustCaptionJobStatus(t, manager, job.ID); status != "FAILED" {
		t.Fatalf("expired caption job status = %q", status)
	}
}

// TestFinishedCaptionGrantIsRevokedAfterRestart closes a durable effect.
func TestFinishedCaptionGrantIsRevokedAfterRestart(t *testing.T) {
	manager, galleryID := mustCaptionScanFixture(t)
	job := mustPendingCaptionJob(t, manager, galleryID, []string{"/home/alice/a.jpg"})
	if _, err := manager.store.db.Exec(`UPDATE media_scan_jobs SET status='COMPLETED' WHERE id=?`, job.ID); err != nil {
		t.Fatalf("finish caption job: %v", err)
	}
	manager.grants = subjectLookupFunc(func(_ context.Context, _ string, _ any, response any) error {
		response.(*authcontract.CaptionGrantRevokeResponse).Revoked = true
		return nil
	})
	if err := manager.cleanupFinishedCaptionGrants(t.Context()); err != nil {
		t.Fatalf("queue grant cleanup: %v", err)
	}
	deadline := time.After(time.Second)
	for {
		var remaining string
		if err := manager.store.db.QueryRow(`SELECT coalesce(grant_token,'') FROM media_scan_jobs WHERE id=?`, job.ID).Scan(&remaining); err != nil {
			t.Fatalf("read grant: %v", err)
		}
		if remaining == "" {
			return
		}
		select {
		case <-deadline:
			t.Fatal("finished job retained its caption grant")
		case <-time.After(time.Millisecond):
		}
	}
}
