package main

import (
	"math"
	"reflect"
	"testing"
)

// TestSemanticQuerySelection covers sentence and non-Latin routing.
func TestSemanticQuerySelection(t *testing.T) {
	for _, test := range []struct {
		name  string
		query string
		want  bool
	}{
		{"English sentence", "a dog on a sofa", true},
		{"Ukrainian word", "собака", true},
		{"English keyword", "dog", false},
		{"Danish keyword", "hund", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := wantsSemanticSearch(test.query); got != test.want {
				t.Fatalf("wants semantic search = %t, want %t", got, test.want)
			}
		})
	}
}

// TestSemanticRankingKeepsPathIdentity covers per-file result order.
func TestSemanticRankingKeepsPathIdentity(t *testing.T) {
	first := make([]float32, embeddingSize)
	first[0] = 1
	second := make([]float32, embeddingSize)
	second[1] = 1
	paths := rankSemanticCandidates([]semanticCandidate{
		{path: "/a.jpg", vector: second},
		{path: "/b.jpg", vector: first},
	}, first)
	if !reflect.DeepEqual(paths, []string{"/b.jpg", "/a.jpg"}) {
		t.Fatalf("ranked paths = %v", paths)
	}
}

// TestSemanticVectorRejectsNonfiniteValues covers worker output validation.
func TestSemanticVectorRejectsNonfiniteValues(t *testing.T) {
	vector := make([]float32, embeddingSize)
	vector[0] = float32(math.Inf(1))
	if decoded := decodeVector(encodeVector(vector)); decoded != nil {
		t.Fatalf("decoded nonfinite vector = %v", decoded)
	}
}

// TestSemanticCandidatesExcludePrivateGallery covers search ACL enforcement.
func TestSemanticCandidatesExcludePrivateGallery(t *testing.T) {
	store, router := mustSearchFixture(t)
	visible := mustNamedGallery(t, router, "COLLECTION", "Visible", "")
	hidden := mustNamedGallery(t, router, "COLLECTION", "Hidden", "")
	mustSearchRows(t, store,
		`UPDATE galleries SET owner_id='bob',visibility='PRIVATE' WHERE id='`+hidden+`'`,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/visible.jpg','IMAGE'),('/home/bob/hidden.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+visible+`','/home/alice/visible.jpg','IMAGE'),('`+hidden+`','/home/bob/hidden.jpg','IMAGE')`,
	)
	controller := galleryController{store: store}
	candidates, err := controller.semanticCandidates(t.Context(), "alice", &listGalleryInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || candidates[0].path != "/home/alice/visible.jpg" {
		t.Fatalf("semantic candidates = %+v", candidates)
	}
}
