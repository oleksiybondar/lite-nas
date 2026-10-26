package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// galleryStore owns durable Moments metadata and its schema migration.
type galleryStore struct {
	db      *sql.DB
	assetMu sync.Mutex
}

// openGalleryStore opens one SQLite database with bounded write contention.
func openGalleryStore(path string) (*galleryStore, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("Moments database path must be clean and absolute")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, statement := range []string{
		"PRAGMA foreign_keys=ON",
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		`CREATE TABLE IF NOT EXISTS galleries (
			id TEXT PRIMARY KEY, type TEXT NOT NULL CHECK(type IN ('ALBUM','COLLECTION')),
			name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', owner_id TEXT NOT NULL,
			visibility TEXT NOT NULL CHECK(visibility IN ('PRIVATE','PUBLIC')),
			upload_directory TEXT,
			wall_strategy TEXT NOT NULL DEFAULT 'mosaic' CHECK(wall_strategy IN ('mosaic','justified','masonry','grid')),
			thumbnail_json TEXT,
			thumbnail_asset TEXT,
			thumbnail_path TEXT,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			CHECK ((type='ALBUM' AND upload_directory IS NOT NULL) OR (type='COLLECTION' AND upload_directory IS NULL))
		)`,
		`CREATE INDEX IF NOT EXISTS galleries_owner_created ON galleries(owner_id, created_at DESC, id DESC)`,
		`CREATE INDEX IF NOT EXISTS galleries_public_created ON galleries(visibility, created_at DESC, id DESC)`,
		`CREATE TABLE IF NOT EXISTS gallery_shares (
			gallery_id TEXT NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
			user_id TEXT NOT NULL, created_at TEXT NOT NULL,
			PRIMARY KEY(gallery_id,user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS gallery_shares_user ON gallery_shares(user_id,gallery_id)`,
		`CREATE TABLE IF NOT EXISTS gallery_references (
			id TEXT PRIMARY KEY, gallery_id TEXT NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
			reference_type TEXT NOT NULL CHECK(reference_type IN ('FILE','DIRECTORY')),
			path TEXT NOT NULL, created_at TEXT NOT NULL,
			UNIQUE(gallery_id,reference_type,path)
		)`,
		`CREATE INDEX IF NOT EXISTS gallery_refs_gallery ON gallery_references(gallery_id,reference_type,path)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY, category TEXT NOT NULL, value TEXT NOT NULL,
			normalized_key TEXT NOT NULL, UNIQUE(category,normalized_key)
		)`,
		`CREATE TABLE IF NOT EXISTS gallery_media (
			gallery_id TEXT NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
			path TEXT NOT NULL, media_type TEXT NOT NULL,
			width INTEGER, height INTEGER, live_photo_path TEXT,
			PRIMARY KEY(gallery_id,path)
		)`,
		`CREATE INDEX IF NOT EXISTS gallery_media_type ON gallery_media(gallery_id,media_type)`,
		`CREATE TABLE IF NOT EXISTS gallery_media_tags (
			gallery_id TEXT NOT NULL, path TEXT NOT NULL,
			tag_id TEXT NOT NULL REFERENCES tags(id),
			PRIMARY KEY(gallery_id,path,tag_id),
			FOREIGN KEY(gallery_id,path) REFERENCES gallery_media(gallery_id,path) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS gallery_media_tags_tag ON gallery_media_tags(tag_id,gallery_id)`,
		`CREATE TABLE IF NOT EXISTS gallery_user_tags (
			id TEXT PRIMARY KEY,
			gallery_id TEXT NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
			value TEXT NOT NULL,
			normalized_key TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(gallery_id,normalized_key)
		)`,
		`CREATE INDEX IF NOT EXISTS gallery_user_tags_gallery ON gallery_user_tags(gallery_id,value)`,
		`CREATE TABLE IF NOT EXISTS media_scan_jobs (
			id TEXT PRIMARY KEY,
			gallery_id TEXT NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
			requested_by TEXT NOT NULL,
			status TEXT NOT NULL,
			total_items INTEGER,
			processed_items INTEGER NOT NULL DEFAULT 0,
			successful_items INTEGER NOT NULL DEFAULT 0,
			failed_items INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			started_at TEXT,
			finished_at TEXT,
			error_summary TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS scan_jobs_gallery ON media_scan_jobs(gallery_id,created_at DESC)`,
		`UPDATE media_scan_jobs SET status='FAILED',finished_at=datetime('now'),error_summary='interrupted by service restart' WHERE status IN ('PENDING','DISCOVERING','RUNNING')`,
	} {
		if _, err = db.Exec(statement); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("initialize Moments database: %w", err)
		}
	}
	columns, err := db.Query(`PRAGMA table_info(galleries)`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("inspect Moments gallery schema: %w", err)
	}
	hasThumbnailPath := false
	for columns.Next() {
		var index, required, primaryKey int
		var name, dataType string
		var defaultValue sql.NullString
		if err := columns.Scan(&index, &name, &dataType, &required, &defaultValue, &primaryKey); err != nil {
			_ = columns.Close()
			_ = db.Close()
			return nil, fmt.Errorf("inspect Moments gallery schema: %w", err)
		}
		hasThumbnailPath = hasThumbnailPath || name == "thumbnail_path"
	}
	err = columns.Err()
	_ = columns.Close()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("inspect Moments gallery schema: %w", err)
	}
	if !hasThumbnailPath {
		if _, err := db.Exec(`ALTER TABLE galleries ADD COLUMN thumbnail_path TEXT`); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("migrate Moments gallery schema: %w", err)
		}
	}
	return &galleryStore{db: db}, nil
}

// close releases the database when the service stops.
func (store *galleryStore) close() error { return store.db.Close() }

// timestamp returns a stable UTC representation for persistence and responses.
func timestamp() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// galleryRecord is the service-owned persisted gallery representation.
type galleryRecord struct {
	ID, Type, Name, Description, OwnerID, Visibility              string
	UploadDirectory, ThumbnailJSON, ThumbnailAsset, ThumbnailPath sql.NullString
	WallStrategy                                                  string
	CreatedAt, UpdatedAt                                          string
}

// scanGallery reads one gallery row without exposing existence to the caller.
func scanGallery(row *sql.Row) (galleryRecord, error) {
	var g galleryRecord
	err := row.Scan(&g.ID, &g.Type, &g.Name, &g.Description, &g.OwnerID, &g.Visibility,
		&g.UploadDirectory, &g.WallStrategy, &g.ThumbnailJSON, &g.ThumbnailAsset, &g.ThumbnailPath, &g.CreatedAt, &g.UpdatedAt)
	return g, err
}

const galleryColumns = `id,type,name,description,owner_id,visibility,upload_directory,wall_strategy,thumbnail_json,thumbnail_asset,thumbnail_path,created_at,updated_at`

// galleryByID fetches a gallery for a visibility or ownership check.
func (store *galleryStore) galleryByID(ctx context.Context, id string) (galleryRecord, error) {
	return scanGallery(store.db.QueryRowContext(ctx, `SELECT `+galleryColumns+` FROM galleries WHERE id=?`, id))
}
