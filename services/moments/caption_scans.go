package main

import (
	"context"
	"time"
)

// captionPaths selects linked still images, omitting existing captions unless forced.
func (m *scanManager) captionPaths(ctx context.Context, galleryID string, force bool) ([]string, error) {
	query := `SELECT gm.path FROM gallery_media gm LEFT JOIN media_captions mc ON mc.path=gm.path
		WHERE gm.gallery_id=? AND gm.media_type IN ('IMAGE','LIVE_PHOTO')`
	if !force {
		query += ` AND mc.path IS NULL`
	}
	rows, err := m.store.db.QueryContext(ctx, query+` ORDER BY gm.path LIMIT 10001`, galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paths := []string{}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

// runCaptionScan generates path-keyed captions without changing source metadata.
func (m *scanManager) runCaptionScan(ctx context.Context, job scanJob, user string, force, resume bool) {
	defer m.release(job)
	created, _ := time.Parse(time.RFC3339Nano, job.CreatedAt)
	var grant string
	var err error
	if resume {
		grant, err = m.captionJobGrant(ctx, job.ID)
		if err != nil {
			m.finish(job.ID, "FAILED", "caption grant unavailable")
			return
		}
		defer func() { go m.cleanCaptionGrant(job.ID, grant, captionGrantCleanupDeadline(created)) }()
	}
	if !resume {
		_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='DISCOVERING',started_at=? WHERE id=?`, timestamp(), job.ID)
	}
	g, err := m.store.galleryByID(ctx, job.GalleryID)
	if err != nil || g.OwnerID != user {
		m.finish(job.ID, "FAILED", "gallery unavailable")
		return
	}
	var paths []string
	if resume {
		paths, err = m.captionJobPaths(ctx, job.ID)
	} else {
		paths, err = m.captionPaths(ctx, job.GalleryID, force)
	}
	if err != nil || len(paths) > 10000 {
		m.finish(job.ID, "FAILED", "caption source listing failed or exceeds scan limit")
		return
	}
	if len(paths) == 0 {
		m.finish(job.ID, "COMPLETED", "")
		return
	}
	if !resume {
		grant, err = m.issueCaptionGrant(ctx, job, m.currentToken(job.GalleryID, job.ID), paths)
		if err == nil {
			err = m.saveCaptionGrant(ctx, job, grant, paths)
			if err != nil {
				go m.cleanCaptionGrant(job.ID, grant, time.Now().Add(captionJobLifetime))
			}
		}
	}
	if err != nil {
		m.finish(job.ID, "FAILED", "caption grant unavailable")
		return
	}
	if !resume {
		defer func() { go m.cleanCaptionGrant(job.ID, grant, captionGrantCleanupDeadline(created)) }()
	}
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='RUNNING',error_summary=NULL WHERE id=?`, job.ID)
	worker := startCaptionWorker(ctx, m.places)
	if worker == nil {
		m.finish(job.ID, "FAILED", "BLIP caption worker unavailable")
		return
	}
	defer worker.close()
	succeeded, failed, processed := 0, 0, 0
	if resume {
		if err := m.store.db.QueryRowContext(ctx, `SELECT processed_items,successful_items,failed_items FROM media_scan_jobs WHERE id=?`, job.ID).Scan(&processed, &succeeded, &failed); err != nil || processed > len(paths) {
			m.finish(job.ID, "FAILED", "caption progress unavailable")
			return
		}
	}
	checkpointed := processed
	lastCheckpoint := time.Now()
	checkpoint := func() bool {
		if processed == checkpointed {
			return true
		}
		_, err := m.store.db.Exec(`UPDATE media_scan_jobs SET processed_items=?,successful_items=?,failed_items=? WHERE id=?`, processed, succeeded, failed, job.ID)
		if err != nil {
			m.finish(job.ID, "FAILED", "caption progress could not be saved")
			return false
		}
		checkpointed = processed
		lastCheckpoint = time.Now()
		return true
	}
	finish := func(status, summary string) {
		if checkpoint() {
			m.finish(job.ID, status, summary)
		}
	}
	for index := processed; index < len(paths); index++ {
		path := paths[index]
		if ctx.Err() != nil {
			finish("CANCELLED", "caption scan cancelled or timed out")
			return
		}
		caption := worker.captionWithGrant(ctx, m.files, grant, path)
		if ctx.Err() != nil {
			finish("CANCELLED", "caption scan cancelled or timed out")
			return
		}
		if caption == nil {
			failed++
		} else {
			_, err := m.store.db.ExecContext(ctx, `INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES(?,?,?,?,?)
				ON CONFLICT(path) DO UPDATE SET caption=excluded.caption,search_key=excluded.search_key,model=excluded.model,updated_at=excluded.updated_at
				WHERE media_captions.caption IS NOT excluded.caption`, path, *caption, searchKey(*caption), captionModel, timestamp())
			if err != nil {
				failed++
			} else {
				succeeded++
			}
		}
		processed = index + 1
		// Captions are durable individually; counters need only a bounded restart checkpoint.
		if processed-checkpointed >= 10 || time.Since(lastCheckpoint) >= 30*time.Minute {
			if !checkpoint() {
				return
			}
		}
		if worker.grantRejected {
			finish("FAILED", "caption grant rejected")
			return
		}
		if worker.failed {
			finish("PARTIAL", "BLIP caption worker stopped; previous captions retained")
			return
		}
	}
	if failed > 0 {
		finish("PARTIAL", "some images could not be captioned; previous captions retained")
		return
	}
	finish("COMPLETED", "")
}
