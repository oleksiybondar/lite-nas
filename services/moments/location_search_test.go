package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"lite-nas/shared/browserauth"
)

// TestLocationFilterMatchesPartOfStoredPlace covers moments-backend/FR-005 for a place within a location tag.
func TestLocationFilterMatchesPartOfStoredPlace(t *testing.T) {
	store, err := openGalleryStore(filepath.Join(t.TempDir(), "moments.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	router := newRouterWithStore(authStub{session: browserauth.Session{UserID: "alice"}}, "", store)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	if _, err := store.db.Exec(`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES(?,?,'IMAGE')`, id, "/home/alice/photo.jpg"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO media_files(path,media_type) VALUES(?,'IMAGE')`, "/home/alice/photo.jpg"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES('place','LOCATION','Málaga, Spain','málaga, spain','malaga, spain')`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO media_file_tags(path,tag_id) VALUES(?,'place')`, "/home/alice/photo.jpg"); err != nil {
		t.Fatal(err)
	}
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?location="+url.QueryEscape("Malaga"), "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Metadata galleryPagination `json:"metadata"`
	}](t, body)
	if page.Metadata.TotalCount != 1 {
		t.Fatalf("location total = %d, want 1", page.Metadata.TotalCount)
	}
}

// TestExistingLocationTagsGainSearchKeys covers the additive migration of already published places.
func TestExistingLocationTagsGainSearchKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "moments.db")
	store, err := openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO tags(id,category,value,normalized_key) VALUES('place','LOCATION','Ålesund, Norge','ålesund, norge')`); err != nil {
		t.Fatal(err)
	}
	if err := store.close(); err != nil {
		t.Fatal(err)
	}
	store, err = openGalleryStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	var key sql.NullString
	if err := store.db.QueryRow(`SELECT search_key FROM tags WHERE id='place'`).Scan(&key); err != nil {
		t.Fatal(err)
	}
	if !key.Valid || key.String != "alesund, norge" {
		t.Fatalf("migrated key = %q, want alesund, norge", key.String)
	}
}

// TestSearchKeyFoldsPlaceSpellings covers common Spanish and Norwegian characters.
func TestSearchKeyFoldsPlaceSpellings(t *testing.T) {
	for _, test := range []struct{ source, want string }{
		{source: "Málaga", want: "malaga"},
		{source: "Ålesund", want: "alesund"},
		{source: "Tromsø", want: "tromso"},
		{source: "Bærum", want: "baerum"},
	} {
		t.Run(test.source, func(t *testing.T) {
			if got := searchKey(test.source); got != test.want {
				t.Fatalf("search key = %q, want %q", got, test.want)
			}
		})
	}
}
