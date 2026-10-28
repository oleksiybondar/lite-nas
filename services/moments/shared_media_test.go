package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// mustPublishedMedia prepares an album with one shared file for reconciliation tests.
func mustPublishedMedia(t *testing.T) (*galleryStore, scanJob) {
	t.Helper()
	store, err := openGalleryStore(filepath.Join(t.TempDir(), "moments.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	for _, statement := range []string{
		`INSERT INTO galleries(id,type,name,owner_id,visibility,upload_directory,created_at,updated_at) VALUES('gallery','ALBUM','Trip','alice','PRIVATE','/home/alice','now','now')`,
		`INSERT INTO media_files(path,media_type,width) VALUES('/home/alice/photo.jpg','IMAGE',640)`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('gallery','/home/alice/photo.jpg','IMAGE')`,
	} {
		if _, err := store.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	return store, scanJob{GalleryID: "gallery"}
}

// TestSharedMediaMigrationPreservesPublishedRows covers moments-backend/FR-009 migration.
func TestSharedMediaMigrationPreservesPublishedRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "moments.db")
	store, err := openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`DROP TABLE media_file_tags`,
		`DROP TABLE media_files`,
		`INSERT INTO galleries(id,type,name,owner_id,visibility,upload_directory,created_at,updated_at) VALUES('gallery','ALBUM','Trip','alice','PRIVATE','/home/alice','now','now')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type,width) VALUES('gallery','/home/alice/photo.jpg','IMAGE',640)`,
		`INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES('tag','LOCATION','Málaga, Spain','málaga, spain','malaga, spain')`,
		`INSERT INTO gallery_media_tags(gallery_id,path,tag_id) VALUES('gallery','/home/alice/photo.jpg','tag')`,
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
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM media_files mf JOIN media_file_tags mt ON mt.path=mf.path WHERE mt.tag_id='tag' AND mf.width=640`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("migrated media tags = %d, want 1", count)
	}
}

// TestFullScanKeepsExistingLinkIdentity covers moments-backend/FR-009 stable reconciliation.
func TestFullScanKeepsExistingLinkIdentity(t *testing.T) {
	store, job := mustPublishedMedia(t)
	var before, after int64
	if err := store.db.QueryRow(`SELECT rowid FROM gallery_media WHERE gallery_id='gallery'`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	manager := &scanManager{store: store}
	if err := manager.publish(context.Background(), job, "alice", "now", []scanMedia{{Path: "/home/alice/photo.jpg", Type: "IMAGE"}}, nil); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRow(`SELECT rowid FROM gallery_media WHERE gallery_id='gallery'`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("gallery link rowid changed from %d to %d", before, after)
	}
}

// TestFullScanKeepsUnchangedTagLink covers moments-backend/FR-009 tag reconciliation.
func TestFullScanKeepsUnchangedTagLink(t *testing.T) {
	store, job := mustPublishedMedia(t)
	manager := &scanManager{store: store}
	media := []scanMedia{{Path: "/home/alice/photo.jpg", Type: "IMAGE", Tags: []scanTag{{Category: "MEDIA_TYPE", Value: "IMAGE"}}}}
	if err := manager.publish(context.Background(), job, "alice", "now", media, nil); err != nil {
		t.Fatal(err)
	}
	var before, after int64
	if err := store.db.QueryRow(`SELECT rowid FROM media_file_tags WHERE path='/home/alice/photo.jpg'`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := manager.publish(context.Background(), job, "alice", "now", media, nil); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRow(`SELECT rowid FROM media_file_tags WHERE path='/home/alice/photo.jpg'`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("tag link rowid changed from %d to %d", before, after)
	}
}

// TestFullScanRemovesOnlyConfirmedDeletedPath covers moments-backend/FR-009 deletion.
func TestFullScanRemovesOnlyConfirmedDeletedPath(t *testing.T) {
	store, job := mustPublishedMedia(t)
	manager := &scanManager{store: store}
	if err := manager.publish(context.Background(), job, "alice", "now", nil, []string{"/home/alice/photo.jpg"}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM media_files WHERE path='/home/alice/photo.jpg'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("confirmed deleted source records = %d, want 0", count)
	}
}

// TestFullScanRetainsUnconfirmedAbsence covers moments-backend/FR-004 access-safe reconciliation.
func TestFullScanRetainsUnconfirmedAbsence(t *testing.T) {
	store, job := mustPublishedMedia(t)
	manager := &scanManager{store: store}
	if err := manager.publish(context.Background(), job, "alice", "now", nil, nil); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM gallery_media WHERE gallery_id='gallery'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("retained gallery links = %d, want 1", count)
	}
}

// TestSingleFileRefreshPreservesLivePhotoPairing keeps gallery-specific motion metadata.
func TestSingleFileRefreshPreservesLivePhotoPairing(t *testing.T) {
	store, _ := mustPublishedMedia(t)
	if _, err := store.db.Exec(`UPDATE gallery_media SET media_type='LIVE_PHOTO',live_photo_path='/home/alice/photo.mov' WHERE gallery_id='gallery'`); err != nil {
		t.Fatal(err)
	}
	controller := galleryController{store: store}
	gallery := galleryRecord{ID: "gallery", Type: "ALBUM", OwnerID: "alice", UploadDirectory: sql.NullString{String: "/home/alice", Valid: true}, UpdatedAt: "now"}
	item := scanMedia{Path: "/home/alice/photo.jpg", Type: "IMAGE"}
	if err := controller.publishUploadedMedia(context.Background(), gallery, &item); err != nil {
		t.Fatal(err)
	}
	var mediaType, motion string
	if err := store.db.QueryRow(`SELECT media_type,live_photo_path FROM gallery_media WHERE gallery_id='gallery'`).Scan(&mediaType, &motion); err != nil {
		t.Fatal(err)
	}
	if mediaType != "LIVE_PHOTO" || motion != "/home/alice/photo.mov" {
		t.Fatalf("pairing = %q %q", mediaType, motion)
	}
}
