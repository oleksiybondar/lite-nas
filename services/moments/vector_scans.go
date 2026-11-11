package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// startVectorScan authorizes an owner request for one background vector job.
func (c galleryController) startVectorScan(ctx context.Context, input *galleryIDInput) (*scanOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	job, err := c.scans.startVector(input.ID, user)
	if err != nil {
		return nil, err
	}
	return &scanOutput{Status: http.StatusAccepted, Body: job}, nil
}

// startVector reserves the gallery and persists a cancellable job before running it.
func (m *scanManager) startVector(galleryID, user string) (scanJob, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	job := scanJob{ID: uuid.NewString(), GalleryID: galleryID, Kind: "VECTOR", Status: "PENDING", CreatedAt: timestamp()}
	_, err := m.store.db.Exec(`INSERT INTO vector_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, user, job.Kind, job.Status, job.CreatedAt)
	if err != nil {
		cancel()
		<-m.seats
		return scanJob{}, huma.Error500InternalServerError("vector job creation failed")
	}
	m.active[galleryID] = &activeScan{jobID: job.ID, kind: job.Kind, cancel: cancel}
	go m.runVectorScan(ctx, job, user)
	return job, nil
}

// runVectorScan builds missing or stale gallery and image vectors in bounded requests.
func (m *scanManager) runVectorScan(ctx context.Context, job scanJob, user string) {
	defer m.release(job)
	finish := func(status, summary string) {
		_, _ = m.store.db.Exec(`UPDATE vector_scan_jobs SET status=?,finished_at=?,error_summary=? WHERE id=?`, status, timestamp(), summary, job.ID)
	}
	_, _ = m.store.db.Exec(`UPDATE vector_scan_jobs SET status='DISCOVERING',started_at=? WHERE id=?`, timestamp(), job.ID)
	controller := galleryController{store: m.store}
	galleries, err := controller.gallerySemanticCandidates(ctx, user, &listGalleryInput{}, job.GalleryID)
	if err != nil || len(galleries) != 1 {
		finish("FAILED", "gallery metadata unavailable")
		return
	}
	images, err := controller.semanticCandidates(ctx, user, &listGalleryInput{}, job.GalleryID)
	if err != nil || len(images) > 10000 {
		finish("FAILED", "image metadata unavailable or exceeds limit")
		return
	}
	_, _ = m.store.db.Exec(`UPDATE vector_scan_jobs SET status='RUNNING',total_items=? WHERE id=?`, len(images)+1, job.ID)
	if m.places == nil || m.places.embeddingScript == "" {
		finish("FAILED", "embedding model unavailable")
		return
	}
	if err := acquireSemanticModel(ctx, m.places); err != nil {
		finish("FAILED", "embedding model busy")
		return
	}
	defer m.places.captionMu.Unlock()
	modelID, size, _, prefix := embeddingSettings(m.places.embeddingModel)
	processed, succeeded, failed := 0, 0, 0
	saveProgress := func() {
		_, _ = m.store.db.Exec(`UPDATE vector_scan_jobs SET processed_items=?,successful_items=?,failed_items=? WHERE id=?`, processed, succeeded, failed, job.ID)
	}
	gallery := galleries[0]
	if ctx.Err() == nil {
		var cachedModel, hash string
		var blob []byte
		err := m.store.db.QueryRowContext(ctx, `SELECT model,text_hash,vector FROM gallery_embeddings WHERE gallery_id=?`, gallery.id).Scan(&cachedModel, &hash, &blob)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && (cachedModel != modelID || hash != gallery.hash || len(decodeVector(blob, size)) != size)) {
			vectors, vectorErr := runEmbeddingModel(ctx, m.places.embeddingScript, m.places.embeddingModel, size, []string{prefix + gallery.text})
			if vectorErr == nil {
				_, vectorErr = m.store.db.ExecContext(ctx, `INSERT INTO gallery_embeddings(gallery_id,model,text_hash,vector) VALUES(?,?,?,?) ON CONFLICT(gallery_id) DO UPDATE SET model=excluded.model,text_hash=excluded.text_hash,vector=excluded.vector`, gallery.id, modelID, gallery.hash, encodeVector(vectors[0]))
			}
			err = vectorErr
		}
		processed++
		if err == nil {
			succeeded++
		} else {
			failed++
		}
		saveProgress()
	}
	for start := 0; start < len(images) && ctx.Err() == nil; start += 100 {
		end := min(start+100, len(images))
		texts := []string{}
		stale := []int{}
		for index := start; index < end; index++ {
			item := images[index]
			var cachedModel, hash string
			var blob []byte
			err := m.store.db.QueryRowContext(ctx, `SELECT model,text_hash,vector FROM media_embeddings WHERE path=?`, item.path).Scan(&cachedModel, &hash, &blob)
			if err == nil && cachedModel == modelID && hash == item.hash && len(decodeVector(blob, size)) == size {
				processed++
				succeeded++
				continue
			}
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				processed++
				failed++
				continue
			}
			stale = append(stale, index)
			texts = append(texts, prefix+item.text)
		}
		if len(texts) > 0 {
			vectors, vectorErr := runEmbeddingModel(ctx, m.places.embeddingScript, m.places.embeddingModel, size, texts)
			for offset, index := range stale {
				processed++
				if vectorErr != nil {
					failed++
					continue
				}
				item := images[index]
				_, err := m.store.db.ExecContext(ctx, `INSERT INTO media_embeddings(path,model,text_hash,vector) VALUES(?,?,?,?) ON CONFLICT(path) DO UPDATE SET model=excluded.model,text_hash=excluded.text_hash,vector=excluded.vector`, item.path, modelID, item.hash, encodeVector(vectors[offset]))
				if err == nil {
					succeeded++
				} else {
					failed++
				}
			}
		}
		saveProgress()
	}
	if ctx.Err() != nil {
		finish("CANCELLED", "vector job cancelled or timed out")
	} else if failed > 0 {
		finish("PARTIAL", "some vectors could not be saved")
	} else {
		finish("COMPLETED", "")
	}
}
