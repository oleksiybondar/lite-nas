package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"lite-nas/shared/browserauth"
)

// scanJob is the durable, caller-visible snapshot of one media or caption scan.
type scanJob struct {
	ID              string  `json:"id"`
	GalleryID       string  `json:"galleryId"`
	Kind            string  `json:"kind"`
	Status          string  `json:"status"`
	TotalItems      *int    `json:"totalItems"`
	ProcessedItems  int     `json:"processedItems"`
	SuccessfulItems int     `json:"successfulItems"`
	FailedItems     int     `json:"failedItems"`
	CreatedAt       string  `json:"createdAt"`
	StartedAt       *string `json:"startedAt"`
	FinishedAt      *string `json:"finishedAt"`
	ErrorSummary    *string `json:"errorSummary"`
}

// scanOutput gives start and polling routes the same progress shape.
type scanOutput struct {
	Status int
	Body   scanJob
}

// scanIDInput validates a scan identifier and caller credentials.
type scanIDInput struct {
	GalleryAuth
	ID string `path:"scanId" format:"uuid"`
}

// scanListInput bounds job history for one visible gallery.
type scanListInput struct {
	GalleryAuth
	ID     string `path:"galleryId" format:"uuid"`
	Kind   string `query:"kind" enum:"MEDIA,CAPTION"`
	Limit  int    `query:"limit" minimum:"1" maximum:"100"`
	Offset int    `query:"offset" minimum:"0" maximum:"100000"`
}

// captionScanInput accepts an explicit request to regenerate saved captions.
type captionScanInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Force bool `json:"force"`
	}
}

// scanListOutput pages scan history without exposing another user's jobs.
type scanListOutput struct {
	Body struct {
		Items []scanJob `json:"items"`
	}
}

// activeScan holds the user token only in process memory and a cancellation handle.
type activeScan struct {
	jobID  string
	kind   string
	token  string
	cancel context.CancelFunc
}

// scanManager bounds concurrent work and owns user-token renewal.
type scanManager struct {
	store  *galleryStore
	auth   browserauth.AuthService
	files  fileServiceClient
	grants subjectLookup
	mu     sync.Mutex
	active map[string]*activeScan
	seats  chan struct{}
	places *cityIndex
}

// newScanManager creates the bounded scan worker state.
func newScanManager(store *galleryStore, auth browserauth.AuthService, files fileServiceClient, places ...*cityIndex) *scanManager {
	manager := &scanManager{store: store, auth: auth, files: files, active: make(map[string]*activeScan), seats: make(chan struct{}, 2)}
	if len(places) > 0 {
		manager.places = places[0]
	}
	return manager
}

// mountScanRoutes registers start, status, history, and cancellation.
func mountScanRoutes(api huma.API, c galleryController) {
	huma.Post(api, "/api/galleries/{galleryId}/scans", c.startScan)
	huma.Post(api, "/api/galleries/{galleryId}/caption-scans", c.startCaptionScan)
	huma.Get(api, "/api/galleries/{galleryId}/scans", c.listScans)
	huma.Get(api, "/api/scans/{scanId}", c.getScan)
	huma.Delete(api, "/api/scans/{scanId}", c.cancelScan)
}

// startCaptionScan starts a caption-only job for an owner-visible gallery.
func (c galleryController) startCaptionScan(ctx context.Context, input *captionScanInput) (*scanOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	job, err := c.scans.start(input.ID, user, accessToken(input.Authorization, input.AccessToken), "CAPTION", input.Body.Force)
	if err != nil {
		return nil, err
	}
	return &scanOutput{Status: http.StatusAccepted, Body: job}, nil
}

// scanFromRow reads a durable job without materializing credentials.
func scanFromRow(row *sql.Row) (scanJob, string, error) {
	var job scanJob
	var owner string
	var total sql.NullInt64
	var started, finished, summary sql.NullString
	err := row.Scan(&job.ID, &job.GalleryID, &owner, &job.Kind, &job.Status, &total,
		&job.ProcessedItems, &job.SuccessfulItems, &job.FailedItems,
		&job.CreatedAt, &started, &finished, &summary)
	if total.Valid {
		value := int(total.Int64)
		job.TotalItems = &value
	}
	if started.Valid {
		job.StartedAt = &started.String
	}
	if finished.Valid {
		job.FinishedAt = &finished.String
	}
	if summary.Valid {
		job.ErrorSummary = &summary.String
	}
	return job, owner, err
}

const scanColumns = `id,gallery_id,requested_by,kind,status,total_items,processed_items,successful_items,failed_items,created_at,started_at,finished_at,error_summary`

// lookupScan enforces same-user polling and hides other users' job IDs.
func (c galleryController) lookupScan(ctx context.Context, id, user string) (scanJob, error) {
	job, owner, err := scanFromRow(c.store.db.QueryRowContext(ctx, `SELECT `+scanColumns+` FROM media_scan_jobs WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) || owner != user {
		return scanJob{}, huma.Error404NotFound("scan not found")
	}
	if err != nil {
		return scanJob{}, huma.Error500InternalServerError("scan lookup failed")
	}
	return job, nil
}

// startScan authorizes the gallery and starts a full scan with the current user JWT.
func (c galleryController) startScan(ctx context.Context, input *galleryIDInput) (*scanOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	job, err := c.scans.start(input.ID, user, accessToken(input.Authorization, input.AccessToken), "MEDIA", false)
	if err != nil {
		return nil, err
	}
	return &scanOutput{Status: http.StatusAccepted, Body: job}, nil
}

// listScans returns only jobs initiated by the authenticated owner.
func (c galleryController) listScans(ctx context.Context, input *scanListInput) (*scanListOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT `+scanColumns+` FROM media_scan_jobs WHERE gallery_id=? AND requested_by=? AND (?='' OR kind=?) ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`, input.ID, user, input.Kind, input.Kind, limit, input.Offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("scan history failed")
	}
	defer rows.Close()
	output := &scanListOutput{}
	output.Body.Items = []scanJob{}
	for rows.Next() {
		var job scanJob
		var owner string
		var total sql.NullInt64
		var started, finished, summary sql.NullString
		if err := rows.Scan(&job.ID, &job.GalleryID, &owner, &job.Kind, &job.Status, &total,
			&job.ProcessedItems, &job.SuccessfulItems, &job.FailedItems,
			&job.CreatedAt, &started, &finished, &summary); err != nil {
			return nil, huma.Error500InternalServerError("scan history failed")
		}
		if total.Valid {
			value := int(total.Int64)
			job.TotalItems = &value
		}
		if started.Valid {
			job.StartedAt = &started.String
		}
		if finished.Valid {
			job.FinishedAt = &finished.String
		}
		if summary.Valid {
			job.ErrorSummary = &summary.String
		}
		output.Body.Items = append(output.Body.Items, job)
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("scan history failed")
	}
	return output, nil
}

// getScan permits same-user polling to refresh the in-memory access JWT.
func (c galleryController) getScan(ctx context.Context, input *scanIDInput) (*scanOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	job, err := c.lookupScan(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, job.GalleryID, user); err != nil {
		return nil, err
	}
	c.scans.renew(job.GalleryID, job.ID, accessToken(input.Authorization, input.AccessToken))
	return &scanOutput{Status: http.StatusOK, Body: job}, nil
}

// cancelScan stops active work and keeps its previous published metadata.
func (c galleryController) cancelScan(ctx context.Context, input *scanIDInput) (*scanOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	job, err := c.lookupScan(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, job.GalleryID, user); err != nil {
		return nil, err
	}
	c.scans.cancel(job.GalleryID, job.ID)
	return &scanOutput{Status: http.StatusOK, Body: job}, nil
}

// start reserves one gallery and one global worker seat before persisting a job.
func (m *scanManager) start(galleryID, user, token, kind string, force bool) (scanJob, error) {
	m.mu.Lock()
	if _, exists := m.active[galleryID]; exists {
		m.mu.Unlock()
		return scanJob{}, huma.Error409Conflict("gallery scan already active")
	}
	if kind == "CAPTION" {
		for _, active := range m.active {
			if active.kind == "CAPTION" {
				m.mu.Unlock()
				return scanJob{}, huma.Error429TooManyRequests("caption scan already active")
			}
		}
	}
	select {
	case m.seats <- struct{}{}:
	default:
		m.mu.Unlock()
		return scanJob{}, huma.Error429TooManyRequests("scan capacity reached")
	}
	scanTimeout := 20 * time.Minute
	if kind == "CAPTION" {
		scanTimeout = captionJobLifetime
	}
	workerCtx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	job := scanJob{ID: uuid.NewString(), GalleryID: galleryID, Kind: kind, Status: "PENDING", CreatedAt: timestamp()}
	_, err := m.store.db.Exec(`INSERT INTO media_scan_jobs(id,gallery_id,requested_by,kind,status,created_at) VALUES(?,?,?,?,?,?)`, job.ID, galleryID, user, job.Kind, job.Status, job.CreatedAt)
	if err != nil {
		cancel()
		<-m.seats
		m.mu.Unlock()
		return scanJob{}, huma.Error500InternalServerError("scan creation failed")
	}
	m.active[galleryID] = &activeScan{jobID: job.ID, kind: kind, token: token, cancel: cancel}
	m.mu.Unlock()
	if kind == "CAPTION" {
		go m.runCaptionScan(workerCtx, job, user, force, false)
	} else {
		go m.run(workerCtx, job, user)
	}
	return job, nil
}

// renew replaces only the active job's volatile token after same-user validation.
func (m *scanManager) renew(galleryID, jobID, token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if active := m.active[galleryID]; active != nil && active.jobID == jobID {
		active.token = token
	}
}

// cancel signals the exact active job; the worker persists terminal status.
func (m *scanManager) cancel(galleryID, jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if active := m.active[galleryID]; active != nil && active.jobID == jobID {
		active.cancel()
	}
}

// currentToken reads the latest same-user credential without logging it.
func (m *scanManager) currentToken(galleryID, jobID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if active := m.active[galleryID]; active != nil && active.jobID == jobID {
		return active.token
	}
	return ""
}

// scanMedia is one staged, gallery-scoped source record.
type scanMedia struct {
	Path          string
	Type          string
	Width, Height *int
	LivePath      *string
	Caption       *string
	Tags          []scanTag
}

// scanTag is a normalized source or generated media-type association.
type scanTag struct {
	Category string `json:"category"`
	Value    string `json:"value"`
	Search   string `json:"-"`
}

// scanMediaFromSource converts one authorized File Service result into source tags.
func scanMediaFromSource(info sourceMetadata, places *cityIndex) scanMedia {
	item := scanMedia{Path: info.Path, Type: info.MediaType, Width: info.Width, Height: info.Height}
	item.Tags = append(item.Tags, scanTag{Category: "MEDIA_TYPE", Value: info.MediaType})
	if extension := strings.TrimPrefix(filepath.Ext(info.Path), "."); extension != "" {
		item.Tags = append(item.Tags, scanTag{Category: "FILE_TYPE", Value: strings.ToUpper(extension)})
	}
	for _, tag := range info.Tags {
		if tag.Category != "" && tag.Value != "" && len(tag.Category) <= 64 && len(tag.Value) <= 256 {
			item.Tags = append(item.Tags, scanTag{Category: tag.Category, Value: tag.Value})
		}
	}
	if info.Location != nil {
		if city, found := places.nearestCity(info.Location.Latitude, info.Location.Longitude); found {
			item.Tags = append(item.Tags, scanTag{Category: "LOCATION", Value: city.name, Search: city.search})
		}
	}
	return item
}

// run discovers authorized files, extracts metadata, then atomically publishes success.
func (m *scanManager) run(ctx context.Context, job scanJob, user string) {
	defer m.release(job)
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='DISCOVERING',started_at=? WHERE id=?`, timestamp(), job.ID)
	g, err := m.store.galleryByID(ctx, job.GalleryID)
	if err != nil || g.OwnerID != user {
		m.finish(job.ID, "FAILED", "gallery unavailable")
		return
	}
	paths, err := m.discover(ctx, job, g)
	if err != nil {
		m.finish(job.ID, "FAILED", "source discovery failed")
		return
	}
	if len(paths) > 10000 {
		m.finish(job.ID, "FAILED", "gallery exceeds scan limit")
		return
	}
	total := len(paths)
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status='RUNNING',total_items=? WHERE id=?`, total, job.ID)
	media := make([]scanMedia, 0, total)
	failed := 0
	for index, path := range paths {
		if ctx.Err() != nil {
			m.finish(job.ID, "CANCELLED", "scan cancelled or timed out")
			return
		}
		token := m.currentToken(job.GalleryID, job.ID)
		session, authErr := m.auth.Me(time.Now(), token)
		if authErr != nil || session.UserID != user {
			m.finish(job.ID, "FAILED", "user access token expired; restart scan")
			return
		}
		info, fetchErr := m.files.metadata(ctx, token, path)
		if fetchErr != nil {
			failed++
		} else {
			item := scanMediaFromSource(info, m.places)
			enrichLocation(ctx, &item, info, m.places)
			media = append(media, item)
		}
		if index%10 == 9 || index == len(paths)-1 {
			_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET processed_items=?,successful_items=?,failed_items=? WHERE id=?`, index+1, len(media), failed, job.ID)
		}
	}
	if failed != 0 {
		m.finish(job.ID, "PARTIAL", "some files could not be extracted; previous metadata retained")
		return
	}
	pairLivePhotos(media)
	deleted, err := m.confirmDeleted(ctx, job, paths)
	if err != nil {
		m.finish(job.ID, "FAILED", "source deletion could not be confirmed")
		return
	}
	if err := m.publish(ctx, job, user, g.UpdatedAt, media, deleted); err != nil {
		m.finish(job.ID, "FAILED", "metadata publication failed")
		return
	}
	m.finish(job.ID, "COMPLETED", "")
}

// release frees one finished job's gallery reservation and worker seat.
func (m *scanManager) release(job scanJob) {
	m.mu.Lock()
	if active := m.active[job.GalleryID]; active != nil && active.jobID == job.ID {
		active.cancel()
		delete(m.active, job.GalleryID)
	}
	m.mu.Unlock()
	<-m.seats
}

// confirmDeleted checks only previously linked paths missing from discovery.
func (m *scanManager) confirmDeleted(ctx context.Context, job scanJob, discovered []string) ([]string, error) {
	seen := make(map[string]bool, len(discovered))
	for _, path := range discovered {
		seen[path] = true
	}
	rows, err := m.store.db.QueryContext(ctx, `SELECT path FROM gallery_media WHERE gallery_id=?`, job.GalleryID)
	if err != nil {
		return nil, err
	}
	var candidates []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if !seen[path] {
			candidates = append(candidates, path)
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	deleted := make([]string, 0, len(candidates))
	for _, path := range candidates {
		missing, err := m.files.missing(ctx, m.currentToken(job.GalleryID, job.ID), path)
		if err != nil {
			return nil, err
		}
		if missing {
			deleted = append(deleted, path)
		}
	}
	return deleted, nil
}

// discover resolves every source through the caller-authorized File Service API.
func (m *scanManager) discover(ctx context.Context, job scanJob, g galleryRecord) ([]string, error) {
	token := m.currentToken(job.GalleryID, job.ID)
	seen := map[string]struct{}{}
	appendPaths := func(paths []string) {
		for _, path := range paths {
			seen[path] = struct{}{}
		}
	}
	if g.Type == "ALBUM" {
		paths, err := m.files.listing(ctx, token, g.UploadDirectory.String)
		if err != nil {
			return nil, err
		}
		appendPaths(paths)
	} else {
		rows, err := m.store.db.QueryContext(ctx, `SELECT reference_type,path FROM gallery_references WHERE gallery_id=? ORDER BY reference_type,path`, g.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var kind, path string
			if err := rows.Scan(&kind, &path); err != nil {
				return nil, err
			}
			if kind == "FILE" {
				canonical, err := m.files.stat(ctx, token, path, "file")
				if err != nil {
					missing, probeErr := m.files.missing(ctx, token, path)
					if probeErr != nil || !missing {
						return nil, err
					}
					continue
				}
				seen[canonical] = struct{}{}
			} else {
				paths, err := m.files.listing(ctx, token, path)
				if err != nil {
					return nil, err
				}
				appendPaths(paths)
			}
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

// pairLivePhotos marks a same-directory, case-sensitive still/MOV pair.
func pairLivePhotos(media []scanMedia) {
	movies := map[string]string{}
	for _, item := range media {
		if strings.EqualFold(filepath.Ext(item.Path), ".mov") {
			movies[strings.TrimSuffix(item.Path, filepath.Ext(item.Path))] = item.Path
		}
	}
	for index := range media {
		if media[index].Type != "IMAGE" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(media[index].Path))
		if ext != ".heic" && ext != ".heif" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		if movie, found := movies[strings.TrimSuffix(media[index].Path, filepath.Ext(media[index].Path))]; found {
			media[index].Type = "LIVE_PHOTO"
			media[index].LivePath = &movie
			media[index].Tags = append(media[index].Tags, scanTag{Category: "MEDIA_TYPE", Value: "LIVE_PHOTO"})
		}
	}
}

// publish reconciles the target gallery's discovered paths in one transaction.
func (m *scanManager) publish(ctx context.Context, job scanJob, user, sourceVersion string, media []scanMedia, deleted []string) error {
	tx, err := m.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var owner, version string
	if err := tx.QueryRowContext(ctx, `SELECT owner_id,updated_at FROM galleries WHERE id=?`, job.GalleryID).Scan(&owner, &version); err != nil || owner != user || version != sourceVersion {
		return errors.New("gallery source changed during scan")
	}
	for _, item := range media {
		if err := m.store.writeMediaFile(ctx, tx, item); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO gallery_media(gallery_id,path,media_type,live_photo_path) VALUES(?,?,?,?)
			ON CONFLICT(gallery_id,path) DO UPDATE SET media_type=excluded.media_type,live_photo_path=excluded.live_photo_path
			WHERE gallery_media.media_type IS NOT excluded.media_type OR gallery_media.live_photo_path IS NOT excluded.live_photo_path`, job.GalleryID, item.Path, item.Type, item.LivePath); err != nil {
			return err
		}
	}
	for _, path := range deleted {
		if _, err := tx.ExecContext(ctx, `DELETE FROM gallery_media WHERE gallery_id=? AND path=?`, job.GalleryID, path); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM media_files WHERE path=? AND NOT EXISTS (SELECT 1 FROM gallery_media WHERE path=?)`, path, path); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// finish records a terminal status without persisting a bearer token.
func (m *scanManager) finish(id, status, summary string) {
	var value any
	if summary != "" {
		value = summary
	}
	_, _ = m.store.db.Exec(`UPDATE media_scan_jobs SET status=?,resume_pending=0,finished_at=?,error_summary=? WHERE id=?`, status, timestamp(), value, id)
}
