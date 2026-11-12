package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"

	"lite-nas/shared/browserauth"
)

// mustSearchFixture prepares a gallery store and owner router for media search tests.
func mustSearchFixture(t *testing.T) (*galleryStore, http.Handler) {
	t.Helper()
	store, err := openGalleryStore(filepath.Join(t.TempDir(), "moments.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.close() })
	router := newRouterWithStore(authStub{session: browserauth.Session{UserID: "alice"}}, "", store)
	return store, router
}

// mustSearchRows installs controlled path-keyed media records.
func mustSearchRows(t *testing.T, store *galleryStore, statements ...string) {
	t.Helper()
	for _, statement := range statements {
		if _, err := store.db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
}

// TestMediaSearchReturnsTaggedImages covers path-keyed image matches alongside gallery results.
func TestMediaSearchReturnsTaggedImages(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES('dog','OBJECT','Dog','dog','dog')`,
		`INSERT INTO media_file_tags(path,tag_id) VALUES('/home/alice/photo.jpg','dog')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q="+url.QueryEscape("dog"), "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Items []struct {
				GalleryID string `json:"galleryId"`
				Path      string `json:"path"`
			} `json:"items"`
		} `json:"matchingMedia"`
	}](t, body)
	if len(page.MatchingMedia.Items) != 1 || page.MatchingMedia.Items[0].GalleryID != id || page.MatchingMedia.Items[0].Path != "/home/alice/photo.jpg" {
		t.Fatalf("matching media = %+v", page.MatchingMedia.Items)
	}
}

// TestMediaSearchGroupsVisibleGalleryLinks covers one image shared across galleries without leaking hidden links.
func TestMediaSearchGroupsVisibleGalleryLinks(t *testing.T) {
	store, router := mustSearchFixture(t)
	first := mustNamedGallery(t, router, "COLLECTION", "A trip", "")
	second := mustNamedGallery(t, router, "COLLECTION", "B trip", "")
	mustSearchRows(t, store,
		`INSERT INTO galleries(id,type,name,owner_id,visibility,created_at,updated_at) VALUES('hidden','COLLECTION','Hidden','bob','PRIVATE','now','now')`,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+first+`','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+second+`','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('hidden','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES('dog','OBJECT','Dog','dog','dog')`,
		`INSERT INTO media_file_tags(path,tag_id) VALUES('/home/alice/photo.jpg','dog')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=dog", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Items []struct {
				LinkedGalleries []struct {
					ID string `json:"id"`
				} `json:"linkedGalleries"`
			} `json:"items"`
		} `json:"matchingMedia"`
	}](t, body)
	got := []string{}
	for _, item := range page.MatchingMedia.Items {
		for _, link := range item.LinkedGalleries {
			got = append(got, link.ID)
		}
	}
	if !reflect.DeepEqual(got, []string{first, second}) {
		t.Fatalf("visible links = %v, want %v", got, []string{first, second})
	}
}

// TestMediaSearchFindsCaption verifies generated words are searchable without predefined labels.
func TestMediaSearchFindsCaption(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES('/home/alice/photo.jpg','A dog on a sofa','a dog on a sofa','test','now')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=sofa", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Items []struct {
				Caption string `json:"caption"`
			} `json:"items"`
		} `json:"matchingMedia"`
	}](t, body)
	if len(page.MatchingMedia.Items) != 1 || page.MatchingMedia.Items[0].Caption != "A dog on a sofa" {
		t.Fatalf("caption matches = %+v", page.MatchingMedia.Items)
	}
}

// TestMediaSearchFindsSeparatedCaptionTerms covers phrases split by clothing names.
func TestMediaSearchFindsSeparatedCaptionTerms(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES('/home/alice/photo.jpg','A person in a blue t-shirt','a person in a blue t-shirt','test','now')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q="+url.QueryEscape("blue shirt"), "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Items []struct {
				Path string `json:"path"`
			} `json:"items"`
		} `json:"matchingMedia"`
	}](t, body)
	if len(page.MatchingMedia.Items) != 1 || page.MatchingMedia.Items[0].Path != "/home/alice/photo.jpg" {
		t.Fatalf("phrase matches = %+v", page.MatchingMedia.Items)
	}
}

// TestMediaSearchRequiresTermsOnOneFile prevents unrelated images from combining matches.
func TestMediaSearchRequiresTermsOnOneFile(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/blue.jpg','IMAGE'),('/home/alice/shirt.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/blue.jpg','IMAGE'),('`+id+`','/home/alice/shirt.jpg','IMAGE')`,
		`INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES('/home/alice/blue.jpg','A blue sky','a blue sky','test','now'),('/home/alice/shirt.jpg','A red shirt','a red shirt','test','now')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q="+url.QueryEscape("blue shirt"), "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Total int `json:"total"`
		} `json:"matchingMedia"`
	}](t, body)
	if page.MatchingMedia.Total != 0 {
		t.Fatalf("cross-image phrase matches = %d", page.MatchingMedia.Total)
	}
}

// TestMediaSearchUsesIndependentPagination covers moments-backend/FR-005 media paging.
func TestMediaSearchUsesIndependentPagination(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES('dog','OBJECT','Dog','dog','dog')`,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/a.jpg','IMAGE'),('/home/alice/b.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/a.jpg','IMAGE'),('`+id+`','/home/alice/b.jpg','IMAGE')`,
		`INSERT INTO media_file_tags(path,tag_id) VALUES('/home/alice/a.jpg','dog'),('/home/alice/b.jpg','dog')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=dog&mediaLimit=1&mediaOffset=1", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Items []struct {
				Path string `json:"path"`
			} `json:"items"`
		} `json:"matchingMedia"`
	}](t, body)
	if len(page.MatchingMedia.Items) != 1 || page.MatchingMedia.Items[0].Path != "/home/alice/b.jpg" {
		t.Fatalf("media page = %+v", page.MatchingMedia.Items)
	}
}

// TestMediaSearchExcludesVideos keeps object search limited to still images.
func TestMediaSearchExcludesVideos(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/movie.mp4','VIDEO')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/movie.mp4','VIDEO')`,
		`INSERT INTO tags(id,category,value,normalized_key,search_key) VALUES('dog','OBJECT','Dog','dog','dog')`,
		`INSERT INTO media_file_tags(path,tag_id) VALUES('/home/alice/movie.mp4','dog')`,
	)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/galleries?q=dog", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		MatchingMedia struct {
			Total int `json:"total"`
		} `json:"matchingMedia"`
	}](t, body)
	if page.MatchingMedia.Total != 0 {
		t.Fatalf("video matches = %d, want 0", page.MatchingMedia.Total)
	}
}

// TestPersonNameSurvivesRescan covers owner names retained outside derived metadata.
func TestPersonNameSurvivesRescan(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/photo.jpg','IMAGE')`,
	)
	request := `{"path":"/home/alice/photo.jpg","name":"Málaga"}`
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/media-person-tags", request, http.StatusOK)
	manager := &scanManager{store: store}
	if err := manager.publish(t.Context(), scanJob{GalleryID: id}, "alice", mustGalleryVersion(t, store, id),
		[]scanMedia{{Path: "/home/alice/photo.jpg", Type: "IMAGE"}}, nil); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT count(*) FROM media_person_tags WHERE path='/home/alice/photo.jpg' AND search_key='malaga'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("person tag count = %d, want 1", count)
	}
}

// mustGalleryVersion reads the current source version for a controlled scan fixture.
func mustGalleryVersion(t *testing.T, store *galleryStore, id string) string {
	t.Helper()
	var version string
	if err := store.db.QueryRow(`SELECT updated_at FROM galleries WHERE id=?`, id).Scan(&version); err != nil {
		t.Fatal(err)
	}
	return version
}
