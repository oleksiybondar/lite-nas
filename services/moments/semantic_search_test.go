package main

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestGallerySemanticCandidatesIncludesManualMetadata verifies gallery text reaches the model.
func TestGallerySemanticCandidatesIncludesManualMetadata(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Wedding photo session", "")
	mustSearchRows(t, store, `INSERT INTO gallery_user_tags(id,gallery_id,value,normalized_key,created_at,updated_at)
		VALUES('tag','`+id+`','Bride and groom','bride and groom','now','now')`)
	candidates, err := (galleryController{store: store}).gallerySemanticCandidates(t.Context(), "alice", &listGalleryInput{Q: "весілля"})
	if err != nil || len(candidates) != 1 || !strings.Contains(candidates[0].text, "Wedding photo session") || !strings.Contains(candidates[0].text, "Bride and groom") {
		t.Fatalf("gallery semantic candidates = %+v, error = %v", candidates, err)
	}
}

// TestSemanticModelWaitsForPriorQuery keeps a quick edit from falling back to zero results.
func TestSemanticModelWaitsForPriorQuery(t *testing.T) {
	model := &cityIndex{}
	model.captionMu.Lock()
	time.AfterFunc(75*time.Millisecond, model.captionMu.Unlock)
	err := acquireSemanticModel(t.Context(), model)
	if err != nil {
		t.Fatalf("wait for prior model: %v", err)
	}
	model.captionMu.Unlock()
}

// TestUkrainianPhraseUsesGalleryMetadata verifies the exact query reaches manual fields.
func TestUkrainianPhraseUsesGalleryMetadata(t *testing.T) {
	store, router := mustSearchFixture(t)
	mustNamedGallery(t, router, "COLLECTION", "Little girl", "A child holding flowers")
	candidates, err := (galleryController{store: store}).gallerySemanticCandidates(t.Context(), "alice", &listGalleryInput{Q: "маленька дівчинка і квіти"})
	if err != nil || len(candidates) != 1 || !strings.Contains(candidates[0].text, "A child holding flowers") {
		t.Fatalf("gallery candidate = %+v, error = %v", candidates, err)
	}
}

// TestUkrainianPhraseUsesGeneratedCaption verifies the exact query reaches image captions.
func TestUkrainianPhraseUsesGeneratedCaption(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Photos", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/girl.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/girl.jpg','IMAGE')`,
		`INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES('/home/alice/girl.jpg','A little girl holding flowers','a little girl holding flowers','test','now')`,
	)
	candidates, err := (galleryController{store: store}).semanticCandidates(t.Context(), "alice", &listGalleryInput{Q: "маленька дівчинка і квіти"})
	if err != nil || len(candidates) != 1 || !strings.Contains(candidates[0].text, "A little girl holding flowers") {
		t.Fatalf("caption candidate = %+v, error = %v", candidates, err)
	}
}

// TestGallerySemanticCandidatesBoundsAllMetadata keeps long fields within worker limits.
func TestGallerySemanticCandidatesBoundsAllMetadata(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Wedding", strings.Repeat("с", 2000))
	mustSearchRows(t, store, `INSERT INTO gallery_user_tags(id,gallery_id,value,normalized_key,created_at,updated_at)
		VALUES('tag','`+id+`','`+strings.Repeat("т", 1000)+`','tag','now','now')`)
	candidates, err := (galleryController{store: store}).gallerySemanticCandidates(t.Context(), "alice", &listGalleryInput{Q: "весілля"})
	if err != nil || len(candidates) != 1 || len("title: none | text: "+candidates[0].text) > 2048 || !strings.Contains(candidates[0].text, "description:") {
		t.Fatalf("bounded candidates = %+v, error = %v", candidates, err)
	}
}

// TestSemanticPathsRequiresConfiguredModel preserves lexical search fallback.
func TestSemanticPathsRequiresConfiguredModel(t *testing.T) {
	_, err := (galleryController{}).semanticPaths(t.Context(), "alice", &listGalleryInput{Q: "blue sky"})
	if err == nil {
		t.Fatal("semantic search accepted an unconfigured model")
	}
}

// TestSemanticPathsRejectsMissingScript checks its local executable boundary.
func TestSemanticPathsRejectsMissingScript(t *testing.T) {
	controller := galleryController{scans: &scanManager{places: &cityIndex{embeddingScript: "/nonexistent/lite-nas-embedding.py"}}}
	_, err := controller.semanticPaths(t.Context(), "alice", &listGalleryInput{Q: "blue sky"})
	if err == nil {
		t.Fatal("semantic search accepted a missing model script")
	}
}

// TestEmbeddingWorkerRejectsMissingScript checks command failure propagation.
func TestEmbeddingWorkerRejectsMissingScript(t *testing.T) {
	_, err := runEmbeddingModel(context.Background(), "/nonexistent/lite-nas-embedding.py", "e5", 384, []string{"query: sky"})
	if err == nil {
		t.Fatal("missing embedding script succeeded")
	}
}

// TestSemanticCandidatesApplyGalleryFilters keeps ranking inside the requested scope.
func TestSemanticCandidatesApplyGalleryFilters(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/photo.jpg','IMAGE')`,
	)
	for _, test := range []struct {
		name  string
		input listGalleryInput
	}{
		{"gallery type", listGalleryInput{Type: []string{"ALBUM"}}},
		{"location", listGalleryInput{Location: []string{"Paris"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidates, err := (galleryController{store: store}).semanticCandidates(t.Context(), "alice", &test.input)
			if err != nil || len(candidates) != 0 {
				t.Fatalf("filtered candidates = %+v, error = %v", candidates, err)
			}
		})
	}
}

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
	first := make([]float32, 768)
	first[0] = 1
	second := make([]float32, 768)
	second[1] = 1
	paths := rankSemanticCandidates([]semanticCandidate{
		{path: "/a.jpg", vector: second},
		{path: "/b.jpg", vector: first},
	}, first)
	if !reflect.DeepEqual(paths, []string{"/b.jpg", "/a.jpg"}) {
		t.Fatalf("ranked paths = %v", paths)
	}
}

// TestSemanticRankingKeepsWordMatchesFirst covers hybrid phrase relevance.
func TestSemanticRankingKeepsWordMatchesFirst(t *testing.T) {
	query := make([]float32, 768)
	query[0] = 1
	strong := make([]float32, 768)
	strong[0] = 1
	weak := make([]float32, 768)
	weak[1] = 1
	paths := rankSemanticCandidates([]semanticCandidate{
		{path: "/semantic.jpg", vector: strong},
		{path: "/lexical.jpg", vector: weak, lexical: true},
	}, query)
	if !reflect.DeepEqual(paths, []string{"/lexical.jpg", "/semantic.jpg"}) {
		t.Fatalf("hybrid ranked paths = %v", paths)
	}
}

// TestSemanticVectorRejectsNonfiniteValues covers worker output validation.
func TestSemanticVectorRejectsNonfiniteValues(t *testing.T) {
	vector := make([]float32, 768)
	vector[0] = float32(math.Inf(1))
	if decoded := decodeVector(encodeVector(vector), 768); decoded != nil {
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

// TestSemanticCandidateRecognizesSeparatedWords covers a blue t-shirt caption.
func TestSemanticCandidateRecognizesSeparatedWords(t *testing.T) {
	store, router := mustSearchFixture(t)
	id := mustNamedGallery(t, router, "COLLECTION", "Trip", "")
	mustSearchRows(t, store,
		`INSERT INTO media_files(path,media_type) VALUES('/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO gallery_media(gallery_id,path,media_type) VALUES('`+id+`','/home/alice/photo.jpg','IMAGE')`,
		`INSERT INTO media_captions(path,caption,search_key,model,updated_at) VALUES('/home/alice/photo.jpg','A person in a blue t-shirt','a person in a blue t-shirt','test','now')`,
	)
	controller := galleryController{store: store}
	candidates, err := controller.semanticCandidates(t.Context(), "alice", &listGalleryInput{Q: "blue shirt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 || !candidates[0].lexical {
		t.Fatalf("phrase candidate = %+v", candidates)
	}
}
