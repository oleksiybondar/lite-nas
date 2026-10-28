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
func (m *scanManager) runCaptionScan(ctx context.Context, job scanJob, user string, force bool) {
	defer m.release(job)
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='DISCOVERING',started_at=? WHERE id=?`, timestamp(), job.ID)
	g, err := m.store.galleryByID(ctx, job.GalleryID)
	if err != nil || g.OwnerID != user {
		m.finish(job.ID, "FAILED", "gallery unavailable")
		return
	}
	paths, err := m.captionPaths(ctx, job.GalleryID, force)
	if err != nil || len(paths) > 10000 {
		m.finish(job.ID, "FAILED", "caption source listing failed or exceeds scan limit")
		return
	}
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='RUNNING',total_items=? WHERE id=?`, len(paths), job.ID)
	if len(paths) == 0 {
		m.finish(job.ID, "COMPLETED", "")
		return
	}
	worker := startCaptionWorker(ctx, m.places)
	if worker == nil {
		m.finish(job.ID, "FAILED", "BLIP caption worker unavailable")
		return
	}
	defer worker.close()
	succeeded, failed := 0, 0
	for index, path := range paths {
		if ctx.Err() != nil {
			m.finish(job.ID, "CANCELLED", "caption scan cancelled or timed out")
			return
		}
		token := m.currentToken(job.GalleryID, job.ID)
		session, authErr := m.auth.Me(time.Now(), token)
		if authErr != nil || session.UserID != user {
			m.finish(job.ID, "FAILED", "user access token expired; restart scan")
			return
		}
		caption := worker.caption(ctx, m.files, token, path)
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
		if index%10 == 9 || index == len(paths)-1 {
			_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET processed_items=?,successful_items=?,failed_items=? WHERE id=?`, index+1, succeeded, failed, job.ID)
		}
		if worker.failed {
			m.finish(job.ID, "PARTIAL", "BLIP caption worker stopped; previous captions retained")
			return
		}
	}
	if failed > 0 {
		m.finish(job.ID, "PARTIAL", "some images could not be captioned; previous captions retained")
		return
	}
	m.finish(job.ID, "COMPLETED", "")
}
