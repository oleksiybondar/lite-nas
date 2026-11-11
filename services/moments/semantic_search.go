package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	e5Model         = "intfloat/multilingual-e5-small@ccc66d3bcd826f577e26b9a4072cc5fe3a7ad6a3"
	gemmaModel      = "google/embeddinggemma-300m-q4@75a84c732f1884df76bec365346230e32f582c82"
	embeddingPython = "/usr/local/lib/lite-nas/moments/embedding-venv/bin/python3"
)

// embeddingSettings binds the selected model to its cache identity and input format.
func embeddingSettings(name string) (string, int, string, string) {
	if name == "embeddinggemma" {
		return gemmaModel, 768, "task: search result | query: ", "title: none | text: "
	}
	return e5Model, 384, "query: ", "passage: "
}

// semanticCandidate holds text and a cached vector for one visible physical image.
type semanticCandidate struct {
	path    string
	text    string
	hash    string
	vector  []float32
	score   float32
	lexical bool
}

// gallerySemanticCandidate holds the metadata and cached vector for one visible gallery.
type gallerySemanticCandidate struct {
	id      string
	text    string
	hash    string
	vector  []float32
	score   float32
	lexical bool
}

// semanticGalleryIDs ranks visible gallery metadata with the existing local model.
// A failed or busy model leaves the caller's lexical results untouched.
func (c galleryController) semanticGalleryIDs(ctx context.Context, user string, input *listGalleryInput) ([]string, error) {
	if c.scans == nil || c.scans.places == nil || c.scans.places.embeddingScript == "" {
		return nil, errors.New("semantic model unavailable")
	}
	model := c.scans.places
	if _, err := os.Stat(model.embeddingScript); err != nil {
		return nil, err
	}
	if _, err := os.Stat(embeddingPython); err != nil {
		return nil, err
	}
	if err := acquireSemanticModel(ctx, model); err != nil {
		return nil, err
	}
	defer model.captionMu.Unlock()
	candidates, err := c.gallerySemanticCandidates(ctx, user, input)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	if len(candidates) > 10000 {
		return nil, errors.New("semantic search exceeds gallery limit")
	}
	modelID, size, queryPrefix, passagePrefix := embeddingSettings(model.embeddingModel)
	texts := []string{queryPrefix + input.Q}
	stale := []int{}
	for index := range candidates {
		candidate := &candidates[index]
		var modelName, hash string
		var blob []byte
		err := c.store.db.QueryRowContext(ctx, `SELECT model,text_hash,vector FROM gallery_embeddings WHERE gallery_id=?`, candidate.id).Scan(&modelName, &hash, &blob)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil && modelName == modelID && hash == candidate.hash {
			candidate.vector = decodeVector(blob, size)
		}
		if len(candidate.vector) != size {
			stale = append(stale, index)
			texts = append(texts, passagePrefix+candidate.text)
		}
	}
	vectors, err := runEmbeddingModel(ctx, model.embeddingScript, model.embeddingModel, size, texts)
	if err != nil {
		return nil, fmt.Errorf("semantic model failed: %w", err)
	}
	if len(vectors) != len(texts) {
		return nil, errors.New("semantic model returned an unexpected vector count")
	}
	for offset, index := range stale {
		candidate := &candidates[index]
		candidate.vector = vectors[offset+1]
		_, _ = c.store.db.ExecContext(ctx, `INSERT INTO gallery_embeddings(gallery_id,model,text_hash,vector) VALUES(?,?,?,?)
			ON CONFLICT(gallery_id) DO UPDATE SET model=excluded.model,text_hash=excluded.text_hash,vector=excluded.vector`,
			candidate.id, modelID, candidate.hash, encodeVector(candidate.vector))
	}
	for index := range candidates {
		for dimension, value := range candidates[index].vector {
			candidates[index].score += value * vectors[0][dimension]
		}
	}
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].lexical != candidates[right].lexical {
			return candidates[left].lexical
		}
		if candidates[left].score != candidates[right].score {
			return candidates[left].score > candidates[right].score
		}
		return candidates[left].id < candidates[right].id
	})
	ids := make([]string, len(candidates))
	for index, candidate := range candidates {
		ids[index] = candidate.id
	}
	return ids, nil
}

// gallerySemanticCandidates applies the gallery list's visibility and filters before embedding metadata.
func (c galleryController) gallerySemanticCandidates(ctx context.Context, user string, input *listGalleryInput) ([]gallerySemanticCandidate, error) {
	query := `SELECT g.id,g.name,g.description,COALESCE((SELECT group_concat(value,'; ') FROM
		(SELECT value FROM gallery_user_tags WHERE gallery_id=g.id ORDER BY value)), '')
		FROM galleries g WHERE (g.owner_id=? OR g.visibility='PUBLIC' OR EXISTS
		(SELECT 1 FROM gallery_shares s WHERE s.gallery_id=g.id AND s.user_id=?))`
	args := []any{user, user}
	for _, filter := range []struct {
		column string
		values []string
	}{
		{"g.type", input.Type}, {"g.visibility", input.Visibility}, {"g.owner_id", input.OwnerID},
	} {
		if len(filter.values) == 0 {
			continue
		}
		query += ` AND ` + filter.column + ` IN (` + strings.TrimSuffix(strings.Repeat("?,", len(filter.values)), ",") + `)`
		for _, value := range filter.values {
			args = append(args, value)
		}
	}
	if len(input.MediaType) > 0 {
		query += ` AND EXISTS (SELECT 1 FROM gallery_media m WHERE m.gallery_id=g.id AND m.media_type IN (` + strings.TrimSuffix(strings.Repeat("?,", len(input.MediaType)), ",") + `))`
		for _, value := range input.MediaType {
			args = append(args, value)
		}
	}
	if len(input.Location) > 0 {
		parts := make([]string, 0, len(input.Location))
		for _, value := range input.Location {
			parts = append(parts, `instr(t.search_key,?)>0`)
			args = append(args, searchKey(value))
		}
		query += ` AND EXISTS (SELECT 1 FROM gallery_media gm JOIN media_file_tags mt ON mt.path=gm.path JOIN tags t ON t.id=mt.tag_id WHERE gm.gallery_id=g.id AND t.category='LOCATION' AND (` + strings.Join(parts, " OR ") + `))`
	}
	rows, err := c.store.db.QueryContext(ctx, query+` ORDER BY g.id LIMIT 10001`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates := []gallerySemanticCandidate{}
	for rows.Next() {
		var candidate gallerySemanticCandidate
		var name, description, tags string
		if err := rows.Scan(&candidate.id, &name, &description, &tags); err != nil {
			return nil, err
		}
		fullText := "name: " + name + "; tags: " + tags + "; description: " + description
		candidate.text = boundedEmbeddingText("name: "+name, 230) +
			"; tags: " + boundedEmbeddingText(tags, 750) +
			"; description: " + boundedEmbeddingText(description, 850)
		key := searchKey(fullText)
		candidate.lexical = matchesSearchTerms(key, searchKey(input.Q)) ||
			matchesSearchTerms(key, ukrainianLatinKey(input.Q))
		hash := sha256.Sum256([]byte(candidate.text))
		candidate.hash = hex.EncodeToString(hash[:])
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

// matchesSearchTerms checks each query word against one metadata search key.
func matchesSearchTerms(key, query string) bool {
	for _, term := range strings.Fields(query) {
		if !strings.Contains(key, term) {
			return false
		}
	}
	return true
}

// boundedEmbeddingText preserves complete UTF-8 characters within the worker's byte limit.
func boundedEmbeddingText(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	var result strings.Builder
	for _, letter := range value {
		if result.Len()+len(string(letter)) > limit {
			break
		}
		result.WriteRune(letter)
	}
	return result.String()
}

// wantsSemanticSearch keeps single Latin words on the inexpensive text path.
func wantsSemanticSearch(query string) bool {
	if len(strings.Fields(query)) > 1 {
		return true
	}
	for _, letter := range query {
		if unicode.IsLetter(letter) && !unicode.In(letter, unicode.Latin) {
			return true
		}
	}
	return false
}

// semanticPaths ranks visible image paths by an optional local text model.
// An unavailable or busy model returns an error so callers retain lexical search.
func (c galleryController) semanticPaths(ctx context.Context, user string, input *listGalleryInput) ([]string, error) {
	if c.scans == nil || c.scans.places == nil || c.scans.places.embeddingScript == "" {
		return nil, errors.New("semantic model unavailable")
	}
	model := c.scans.places
	if _, err := os.Stat(model.embeddingScript); err != nil {
		return nil, err
	}
	if _, err := os.Stat(embeddingPython); err != nil {
		return nil, err
	}
	if err := acquireSemanticModel(ctx, model); err != nil {
		return nil, err
	}
	defer model.captionMu.Unlock()
	candidates, err := c.semanticCandidates(ctx, user, input)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	if len(candidates) > 10000 {
		return nil, errors.New("semantic search exceeds image limit")
	}
	modelID, size, queryPrefix, passagePrefix := embeddingSettings(model.embeddingModel)
	texts := []string{queryPrefix + input.Q}
	stale := make([]int, 0, len(candidates))
	for index := range candidates {
		candidate := &candidates[index]
		var modelName, hash string
		var blob []byte
		err := c.store.db.QueryRowContext(ctx, `SELECT model,text_hash,vector FROM media_embeddings WHERE path=?`, candidate.path).Scan(&modelName, &hash, &blob)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil && modelName == modelID && hash == candidate.hash {
			candidate.vector = decodeVector(blob, size)
		}
		if len(candidate.vector) != size {
			stale = append(stale, index)
			texts = append(texts, passagePrefix+candidate.text)
		}
	}
	vectors, err := runEmbeddingModel(ctx, model.embeddingScript, model.embeddingModel, size, texts)
	if err != nil {
		return nil, fmt.Errorf("semantic model failed: %w", err)
	}
	if len(vectors) != len(texts) {
		return nil, errors.New("semantic model returned an unexpected vector count")
	}
	for offset, index := range stale {
		candidate := &candidates[index]
		candidate.vector = vectors[offset+1]
		_, _ = c.store.db.ExecContext(ctx, `INSERT INTO media_embeddings(path,model,text_hash,vector) VALUES(?,?,?,?)
			ON CONFLICT(path) DO UPDATE SET model=excluded.model,text_hash=excluded.text_hash,vector=excluded.vector`,
			candidate.path, modelID, candidate.hash, encodeVector(candidate.vector))
	}
	return rankSemanticCandidates(candidates, vectors[0]), nil
}

// acquireSemanticModel briefly waits for an older search to release the shared model.
func acquireSemanticModel(ctx context.Context, model *cityIndex) error {
	deadline, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := deadline.Err(); err != nil {
			return err
		}
		if model.captionMu.TryLock() {
			return nil
		}
		select {
		case <-deadline.Done():
			return deadline.Err()
		case <-ticker.C:
		}
	}
}

// rankSemanticCandidates sorts by cosine similarity of normalized vectors.
func rankSemanticCandidates(candidates []semanticCandidate, query []float32) []string {
	for index := range candidates {
		for dimension, value := range candidates[index].vector {
			candidates[index].score += value * query[dimension]
		}
	}
	sort.Slice(candidates, func(left, right int) bool {
		if candidates[left].lexical != candidates[right].lexical {
			return candidates[left].lexical
		}
		if candidates[left].score == candidates[right].score {
			return candidates[left].path < candidates[right].path
		}
		return candidates[left].score > candidates[right].score
	})
	paths := make([]string, len(candidates))
	for index, candidate := range candidates {
		paths[index] = candidate.path
	}
	return paths
}

// semanticCandidates collects only files in galleries visible to the caller.
func (c galleryController) semanticCandidates(ctx context.Context, user string, input *listGalleryInput) ([]semanticCandidate, error) {
	query := `SELECT DISTINCT gm.path,COALESCE(mc.caption,'') FROM gallery_media gm JOIN galleries g ON g.id=gm.gallery_id
		LEFT JOIN media_captions mc ON mc.path=gm.path WHERE gm.media_type IN ('IMAGE','LIVE_PHOTO') AND
		(g.owner_id=? OR g.visibility='PUBLIC' OR EXISTS (SELECT 1 FROM gallery_shares s WHERE s.gallery_id=g.id AND s.user_id=?))`
	args := []any{user, user}
	for _, filter := range []struct {
		column string
		values []string
	}{{"g.type", input.Type}, {"g.visibility", input.Visibility}, {"g.owner_id", input.OwnerID}, {"gm.media_type", input.MediaType}} {
		if len(filter.values) == 0 {
			continue
		}
		query += ` AND ` + filter.column + ` IN (` + strings.TrimSuffix(strings.Repeat("?,", len(filter.values)), ",") + `)`
		for _, value := range filter.values {
			args = append(args, value)
		}
	}
	if len(input.Location) > 0 {
		parts := make([]string, 0, len(input.Location))
		for _, value := range input.Location {
			parts = append(parts, `instr(t.search_key,?)>0`)
			args = append(args, searchKey(value))
		}
		query += ` AND EXISTS (SELECT 1 FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id
			WHERE mt.path=gm.path AND t.category='LOCATION' AND (` + strings.Join(parts, " OR ") + `))`
	}
	rows, err := c.store.db.QueryContext(ctx, query+` ORDER BY gm.path LIMIT 10001`, args...)
	if err != nil {
		return nil, err
	}
	candidates := []semanticCandidate{}
	for rows.Next() {
		var candidate semanticCandidate
		var caption string
		if err := rows.Scan(&candidate.path, &caption); err != nil {
			_ = rows.Close()
			return nil, err
		}
		candidate.text = caption
		candidates = append(candidates, candidate)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	for index := range candidates {
		candidate := &candidates[index]
		tags, err := c.store.db.QueryContext(ctx, `SELECT category,value FROM (
			SELECT t.category,t.value FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id WHERE mt.path=?
			UNION ALL SELECT 'PERSON',value FROM media_person_tags WHERE path=?
		) ORDER BY category,value`, candidate.path, candidate.path)
		if err != nil {
			return nil, err
		}
		for tags.Next() {
			var category, value string
			if err := tags.Scan(&category, &value); err != nil {
				_ = tags.Close()
				return nil, err
			}
			part := "; " + category + ": " + value
			if len(candidate.text)+len(part) <= 1900 {
				candidate.text += part
			}
		}
		err = tags.Err()
		_ = tags.Close()
		if err != nil {
			return nil, err
		}
		if candidate.text == "" {
			candidate.text = "image"
		}
		candidate.lexical = true
		key := searchKey(candidate.text)
		for _, term := range strings.Fields(searchKey(input.Q)) {
			if !strings.Contains(key, term) {
				candidate.lexical = false
				break
			}
		}
		hash := sha256.Sum256([]byte(candidate.text))
		candidate.hash = hex.EncodeToString(hash[:])
	}
	return candidates, nil
}

// runEmbeddingModel invokes a separate, confined process for one request.
func runEmbeddingModel(ctx context.Context, script, modelName string, size int, texts []string) ([][]float32, error) {
	request, err := json.Marshal(struct {
		Texts []string `json:"texts"`
		Model string   `json:"model"`
	}{Texts: texts, Model: modelName})
	if err != nil {
		return nil, err
	}
	modelCtx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()
	command := exec.CommandContext(modelCtx, embeddingPython, "-u", script)
	command.Stdin = bytes.NewReader(append(request, '\n'))
	command.Stderr = os.Stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("embedding worker failed: %w", err)
	}
	if len(output) > 48<<20 {
		return nil, errors.New("embedding worker output too large")
	}
	var response struct {
		Vectors []string `json:"vectors"`
	}
	if json.Unmarshal(output, &response) != nil || len(response.Vectors) != len(texts) {
		return nil, errors.New("invalid embedding worker response")
	}
	vectors := make([][]float32, len(texts))
	for index, encoded := range response.Vectors {
		blob, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, err
		}
		vectors[index] = decodeVector(blob, size)
		if len(vectors[index]) != size {
			return nil, errors.New("invalid embedding vector")
		}
	}
	return vectors, nil
}

// encodeVector stores fixed-width little-endian values in SQLite.
func encodeVector(vector []float32) []byte {
	blob := make([]byte, 4*len(vector))
	for index, value := range vector {
		binary.LittleEndian.PutUint32(blob[index*4:], math.Float32bits(value))
	}
	return blob
}

// decodeVector rejects wrong shapes and nonfinite model output.
func decodeVector(blob []byte, size int) []float32 {
	if len(blob) != 4*size {
		return nil
	}
	vector := make([]float32, size)
	for index := range vector {
		value := math.Float32frombits(binary.LittleEndian.Uint32(blob[index*4:]))
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil
		}
		vector[index] = value
	}
	return vector
}
