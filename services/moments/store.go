package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
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
			normalized_key TEXT NOT NULL, search_key TEXT NOT NULL DEFAULT '', UNIQUE(category,normalized_key)
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
		`CREATE TABLE IF NOT EXISTS media_person_tags (
			id TEXT PRIMARY KEY,
			path TEXT NOT NULL REFERENCES media_files(path) ON DELETE CASCADE,
			value TEXT NOT NULL,
			normalized_key TEXT NOT NULL,
			search_key TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(path,normalized_key)
		)`,
		`CREATE INDEX IF NOT EXISTS media_person_tags_search ON media_person_tags(search_key,path)`,
		`CREATE TABLE IF NOT EXISTS media_captions (
			path TEXT PRIMARY KEY REFERENCES media_files(path) ON DELETE CASCADE,
			caption TEXT NOT NULL,
			search_key TEXT NOT NULL,
			model TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS reverse_geocode_cache (
			coordinate_key TEXT PRIMARY KEY,
			place_name TEXT NOT NULL,
			search_key TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
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
	if err := migrateTagSearchKeys(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate Moments tag search: %w", err)
	}
	if err := migrateSharedMedia(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate Moments shared media: %w", err)
	}
	return &galleryStore{db: db}, nil
}

// migrateSharedMedia copies existing gallery-scoped extraction into path-keyed tables once.
func migrateSharedMedia(db *sql.DB) error {
	var exists int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='media_files'`).Scan(&exists); err != nil {
		return err
	}
	if exists != 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, statement := range []string{
		`CREATE TABLE media_files (
			path TEXT PRIMARY KEY, media_type TEXT NOT NULL,
			width INTEGER, height INTEGER
		)`,
		`CREATE TABLE media_file_tags (
			path TEXT NOT NULL REFERENCES media_files(path) ON DELETE CASCADE,
			tag_id TEXT NOT NULL REFERENCES tags(id),
			PRIMARY KEY(path,tag_id)
		)`,
		`CREATE INDEX media_file_tags_tag ON media_file_tags(tag_id,path)`,
		`INSERT OR IGNORE INTO media_files(path,media_type,width,height)
			SELECT path,CASE WHEN media_type='LIVE_PHOTO' THEN 'IMAGE' ELSE media_type END,width,height FROM gallery_media ORDER BY gallery_id,path`,
		`INSERT OR IGNORE INTO media_file_tags(path,tag_id)
			SELECT path,tag_id FROM gallery_media_tags ORDER BY gallery_id,path,tag_id`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// writeMediaFile refreshes one path-keyed source record and its derived tags.
func (store *galleryStore) writeMediaFile(ctx context.Context, tx *sql.Tx, item scanMedia) error {
	sourceType := item.Type
	if sourceType == "LIVE_PHOTO" {
		sourceType = "IMAGE"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO media_files(path,media_type,width,height) VALUES(?,?,?,?)
		ON CONFLICT(path) DO UPDATE SET media_type=excluded.media_type,width=excluded.width,height=excluded.height
		WHERE media_files.media_type IS NOT excluded.media_type OR media_files.width IS NOT excluded.width OR media_files.height IS NOT excluded.height`, item.Path, sourceType, item.Width, item.Height); err != nil {
		return err
	}
	if item.Caption != nil {
		if _, err := tx.ExecContext(ctx, `INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES(?,?,?,?,?)
			ON CONFLICT(path) DO UPDATE SET caption=excluded.caption,search_key=excluded.search_key,model=excluded.model,updated_at=excluded.updated_at
			WHERE media_captions.caption IS NOT excluded.caption`, item.Path, *item.Caption, searchKey(*item.Caption), captionModel, timestamp()); err != nil {
			return err
		}
	}
	desired := []any{item.Path}
	for _, tag := range item.Tags {
		key := norm.NFC.String(strings.ToLower(strings.TrimSpace(tag.Value)))
		if key == "" {
			continue
		}
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(tag.Category+"\x00"+key)).String()
		search := searchKey(tag.Value)
		if tag.Search != "" {
			search = tag.Search
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES(?,?,?,?,?)
			ON CONFLICT(category,normalized_key) DO UPDATE SET search_key=excluded.search_key
			WHERE tags.search_key IS NOT excluded.search_key`, id, tag.Category, tag.Value, key, search); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO media_file_tags(path,tag_id) VALUES(?,?)`, item.Path, id); err != nil {
			return err
		}
		desired = append(desired, id)
	}
	if len(desired) == 1 {
		_, err := tx.ExecContext(ctx, `DELETE FROM media_file_tags WHERE path=?`, item.Path)
		return err
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(desired)-1), ",")
	if _, err := tx.ExecContext(ctx, `DELETE FROM media_file_tags WHERE path=? AND tag_id NOT IN (`+placeholders+`)`, desired...); err != nil {
		return err
	}
	return nil
}

// migrateTagSearchKeys adds and fills accent-folded keys without changing tag identity.
func migrateTagSearchKeys(db *sql.DB) error {
	columns, err := db.Query(`PRAGMA table_info(tags)`)
	if err != nil {
		return err
	}
	hasSearchKey := false
	for columns.Next() {
		var index, required, primaryKey int
		var name, dataType string
		var defaultValue sql.NullString
		if err := columns.Scan(&index, &name, &dataType, &required, &defaultValue, &primaryKey); err != nil {
			_ = columns.Close()
			return err
		}
		hasSearchKey = hasSearchKey || name == "search_key"
	}
	err = columns.Err()
	_ = columns.Close()
	if err != nil {
		return err
	}
	if !hasSearchKey {
		if _, err := db.Exec(`ALTER TABLE tags ADD COLUMN search_key TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	rows, err := db.Query(`SELECT id,value FROM tags WHERE search_key=''`)
	if err != nil {
		return err
	}
	type tagKey struct{ id, value string }
	var pending []tagKey
	for rows.Next() {
		var tag tagKey
		if err := rows.Scan(&tag.id, &tag.value); err != nil {
			_ = rows.Close()
			return err
		}
		pending = append(pending, tag)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}
	for _, tag := range pending {
		if _, err := db.Exec(`UPDATE tags SET search_key=? WHERE id=?`, searchKey(tag.value), tag.id); err != nil {
			return err
		}
	}
	return nil
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
