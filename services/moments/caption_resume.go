package main

import (
	"context"
	"fmt"
	"time"
)

// resumeCaptionJobs restarts durable caption work under its original deadline.
func (m *scanManager) resumeCaptionJobs(ctx context.Context) error {
	if err := m.cleanupFinishedCaptionGrants(ctx); err != nil {
		return err
	}
	rows, err := m.store.db.QueryContext(ctx, `SELECT id FROM media_scan_jobs WHERE kind='CAPTION' AND resume_pending=1 AND grant_token IS NOT NULL AND status IN ('PENDING','DISCOVERING','RUNNING','PAUSED') ORDER BY created_at`)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		job, user, err := scanFromRow(m.store.db.QueryRowContext(ctx, `SELECT `+scanColumns+` FROM media_scan_jobs WHERE id=?`, id))
		if err != nil {
			return err
		}
		created, err := time.Parse(time.RFC3339Nano, job.CreatedAt)
		if err != nil || !time.Now().Before(created.Add(captionJobLifetime)) {
			m.finish(job.ID, "FAILED", "caption job deadline expired")
			deadline := time.Now().Add(captionJobLifetime)
			if err == nil {
				deadline = captionGrantCleanupDeadline(created)
			}
			m.cleanupJobGrant(job.ID, deadline)
			continue
		}
		m.mu.Lock()
		select {
		case m.seats <- struct{}{}:
		default:
			m.mu.Unlock()
			return fmt.Errorf("caption restart exceeds scan capacity")
		}
		workerCtx, cancel := context.WithDeadline(context.Background(), created.Add(captionJobLifetime))
		m.active[job.GalleryID] = &activeScan{jobID: job.ID, kind: "CAPTION", cancel: cancel}
		m.mu.Unlock()
		go m.runCaptionScan(workerCtx, job, user, false, true)
	}
	return nil
}

// cleanupFinishedCaptionGrants resumes pending revocations after a restart.
func (m *scanManager) cleanupFinishedCaptionGrants(ctx context.Context) error {
	rows, err := m.store.db.QueryContext(ctx, `SELECT id,grant_token,created_at FROM media_scan_jobs WHERE kind='CAPTION' AND grant_token IS NOT NULL AND status IN ('COMPLETED','PARTIAL','FAILED','CANCELLED')`)
	if err != nil {
		return err
	}
	type pending struct {
		id, grant string
		deadline  time.Time
	}
	var grants []pending
	for rows.Next() {
		var id, grant, createdAt string
		if err := rows.Scan(&id, &grant, &createdAt); err != nil {
			_ = rows.Close()
			return err
		}
		created, err := time.Parse(time.RFC3339Nano, createdAt)
		deadline := time.Now().Add(captionJobLifetime)
		if err == nil {
			deadline = captionGrantCleanupDeadline(created)
		}
		grants = append(grants, pending{id: id, grant: grant, deadline: deadline})
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	for _, grant := range grants {
		go m.cleanCaptionGrant(grant.id, grant.grant, grant.deadline)
	}
	return nil
}

// cleanupJobGrant schedules revocation for an active job that cannot resume.
func (m *scanManager) cleanupJobGrant(jobID string, deadline time.Time) {
	grant, err := m.captionJobGrant(context.Background(), jobID)
	if err == nil {
		go m.cleanCaptionGrant(jobID, grant, deadline)
	}
}
