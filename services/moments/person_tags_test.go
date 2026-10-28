package main

import (
	"net/http"
	"testing"
)

// TestPersonTagRejectsUnlinkedImage prevents names on paths outside the owner's gallery.
func TestPersonTagRejectsUnlinkedImage(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store, `INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/media-person-tags",
		`{"path":"/home/alice/photo.jpg","name":"Alice"}`, http.StatusNotFound)
}

// TestPersonTagRejectsVideo limits manual person names to linked images.
func TestPersonTagRejectsVideo(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/movie.mp4','VIDEO')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/movie.mp4','VIDEO')`,
	)
	mustGalleryRequest(t, router, http.MethodPost, "/api/galleries/"+id+"/media-person-tags",
		`{"path":"/home/alice/movie.mp4","name":"Alice"}`, http.StatusNotFound)
}
