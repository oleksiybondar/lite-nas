package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"lite-nas/shared/workers"
)

// startAutomaticSearch checks saved galleries at startup and every hour.
func (m *scanManager) startAutomaticSearch(ctx context.Context) error {
	// The timer owns this one-slot tick channel; a long scan coalesces later ticks.
	// Neither worker closes it, and cancellation stops both workers.
	timer, ticks, err := workers.NewPollingTimerWorker(time.Hour, 1)
	if err != nil {
		return err
	}
	timer.Start(ctx)
	workers.NewBaseWorker(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticks:
				if err := m.processAutomaticSearch(ctx); err != nil {
					slog.Warn("automatic Moments search processing failed", "error", err)
				}
			}
		}
	}).Start(ctx)
	return nil
}

// processAutomaticSearch visits one gallery at a time, leaving active jobs alone.
func (m *scanManager) processAutomaticSearch(ctx context.Context) error {
	m.mu.Lock()
	busy := len(m.active) > 0
	m.mu.Unlock()
	if busy {
		return nil
	}
	rows, err := m.store.db.QueryContext(ctx, `SELECT id,owner_id FROM galleries ORDER BY id`)
	if err != nil {
		return err
	}
	type gallery struct{ id, owner string }
	var galleries []gallery
	for rows.Next() {
		var item gallery
		if err := rows.Scan(&item.id, &item.owner); err != nil {
			_ = rows.Close()
			return err
		}
		galleries = append(galleries, item)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	for _, item := range galleries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		paths, err := m.captionPaths(ctx, item.id, false)
		if err != nil {
			return err
		}
		if m.automaticCaptions && len(paths) > 0 && len(paths) <= 10000 && m.places != nil && m.places.captionScript != "" {
			job, err := m.startAutomaticCaption(item.id, item.owner, false)
			if err != nil {
				return err
			}
			if err := m.waitForSearchJob(ctx, job); err != nil {
				return err
			}
		}
		if m.automaticVectors && m.places != nil && m.places.embeddingScript != "" {
			job, err := m.startVector(item.id, item.owner)
			if err != nil {
				return err
			}
			if err := m.waitForSearchJob(ctx, job); err != nil {
				return err
			}
		}
	}
	return nil
}

// waitForSearchJob lets the hourly scheduler keep a single gallery in flight.
func (m *scanManager) waitForSearchJob(ctx context.Context, job scanJob) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	table := "media_scan_jobs"
	if job.Kind == "VECTOR" {
		table = "vector_scan_jobs"
	}
	for {
		var status string
		if err := m.store.db.QueryRowContext(ctx, `SELECT status FROM `+table+` WHERE id=?`, job.ID).Scan(&status); err != nil {
			return err
		}
		if status == "COMPLETED" || status == "PARTIAL" || status == "FAILED" || status == "CANCELLED" {
			m.mu.Lock()
			active := m.active[job.GalleryID]
			m.mu.Unlock()
			if active == nil || active.jobID != job.ID {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// startAutomaticCaption records a service-account caption pass for one gallery.
func (m *scanManager) startAutomaticCaption(galleryID, owner string, followVector bool) (scanJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.active) > 0 {
		return scanJob{}, huma.Error409Conflict("another gallery job is active")
	}
	select {
	case m.seats <- struct{}{}:
	default:
		return scanJob{}, huma.Error429TooManyRequests("scan capacity reached")
	}
	ctx, cancel := context.WithTimeout(context.Background(), captionJobLifetime)
	job := scanJob{ID: uuid.NewString(), GalleryID: galleryID, Kind: "CAPTION", Status: "PENDING", CreatedAt: timestamp()}
	if _, err := m.store.db.Exec(`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, owner, job.Kind, job.Status, job.CreatedAt); err != nil {
		cancel()
		<-m.seats
		return scanJob{}, err
	}
	m.active[galleryID] = &activeScan{jobID: job.ID, kind: job.Kind, automatic: true, followVector: followVector, cancel: cancel}
	go m.runAutomaticCaption(ctx, job)
	return job, nil
}

// runAutomaticCaption asks File Service to authorize each read as Moments.
func (m *scanManager) runAutomaticCaption(ctx context.Context, job scanJob) {
	defer m.release(job)
	finish := func(status, summary string) { m.finish(job.ID, status, summary) }
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='DISCOVERING',started_at=? WHERE id=?`, timestamp(), job.ID)
	paths, err := m.captionPaths(ctx, job.GalleryID, false)
	if err != nil || len(paths) > 10000 {
		finish("FAILED", "caption source listing failed or exceeds limit")
		return
	}
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='RUNNING',total_items=? WHERE id=?`, len(paths), job.ID)
	if len(paths) == 0 {
		finish("COMPLETED", "")
		return
	}
	if m.tokens == nil || m.tokens.Login(ctx) != nil {
		finish("FAILED", "service authentication unavailable")
		return
	}
	worker := startCaptionWorker(ctx, m.places)
	if worker == nil {
		finish("FAILED", "BLIP caption worker unavailable")
		return
	}
	failed := 0
	for index, path := range paths {
		if ctx.Err() != nil {
			finish("CANCELLED", "automatic caption job cancelled or timed out")
			return
		}
		token, expires, err := m.tokens.Token()
		if err == nil && time.Until(expires) < 5*time.Minute {
			err = m.tokens.Refresh(ctx)
			if err == nil {
				token, _, err = m.tokens.Token()
			}
		}
		if err != nil {
			finish("FAILED", "service authentication expired")
			return
		}
		caption := worker.caption(ctx, m.files, token, path)
		if caption == nil {
			failed++
		} else {
			_, err = m.store.db.ExecContext(ctx, `INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES(?,?,?,?,?) ON CONFLICT(path) DO UPDATE SET caption=excluded.caption,search_key=excluded.search_key,model=excluded.model,updated_at=excluded.updated_at WHERE media_captions.caption IS NOT excluded.caption`, path, *caption, searchKey(*caption), captionModel, timestamp())
			if err != nil {
				failed++
			}
		}
		_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET processed_items=?,successful_items=?,failed_items=? WHERE id=?`, index+1, index+1-failed, failed, job.ID)
		if worker.failed {
			finish("PARTIAL", "BLIP worker stopped")
			return
		}
	}
	if failed > 0 {
		finish("PARTIAL", "some images were denied or could not be captioned")
	} else {
		finish("COMPLETED", "")
	}
}
