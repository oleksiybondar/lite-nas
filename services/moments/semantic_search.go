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
	"math"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	embeddingModel  = "intfloat/multilingual-e5-small@ccc66d3bcd826f577e26b9a4072cc5fe3a7ad6a3"
	embeddingPython = "/usr/local/lib/lite-nas/moments/embedding-venv/bin/python3"
	embeddingSize   = 384
)

// semanticCandidate holds text and a cached vector for one visible physical image.
type semanticCandidate struct {
	path   string
	text   string
	hash   string
	vector []float32
	score  float32
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
	if !model.captionMu.TryLock() {
		return nil, errors.New("another model is active")
	}
	defer model.captionMu.Unlock()
	candidates, err := c.semanticCandidates(ctx, user, input)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	if len(candidates) > 10000 {
		return nil, errors.New("semantic search exceeds image limit")
	}
	texts := []string{"query: " + input.Q}
	stale := make([]int, 0, len(candidates))
	for index := range candidates {
		candidate := &candidates[index]
		var modelName, hash string
		var blob []byte
		err := c.store.db.QueryRowContext(ctx, `SELECT model,text_hash,vector FROM media_embeddings WHERE path=?`, candidate.path).Scan(&modelName, &hash, &blob)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil && modelName == embeddingModel && hash == candidate.hash {
			candidate.vector = decodeVector(blob)
		}
		if len(candidate.vector) != embeddingSize {
			stale = append(stale, index)
			texts = append(texts, "passage: "+candidate.text)
		}
	}
	vectors, err := runEmbeddingModel(ctx, model.embeddingScript, texts)
	if err != nil || len(vectors) != len(texts) {
		return nil, errors.New("semantic model failed")
	}
	for offset, index := range stale {
		candidate := &candidates[index]
		candidate.vector = vectors[offset+1]
		_, _ = c.store.db.ExecContext(ctx, `INSERT INTO media_embeddings(path,model,text_hash,vector) VALUES(?,?,?,?)
			ON CONFLICT(path) DO UPDATE SET model=excluded.model,text_hash=excluded.text_hash,vector=excluded.vector`,
			candidate.path, embeddingModel, candidate.hash, encodeVector(candidate.vector))
	}
	return rankSemanticCandidates(candidates, vectors[0]), nil
}

// rankSemanticCandidates sorts by cosine similarity of normalized vectors.
func rankSemanticCandidates(candidates []semanticCandidate, query []float32) []string {
	for index := range candidates {
		for dimension, value := range candidates[index].vector {
			candidates[index].score += value * query[dimension]
		}
	}
	sort.Slice(candidates, func(left, right int) bool {
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
		hash := sha256.Sum256([]byte(candidate.text))
		candidate.hash = hex.EncodeToString(hash[:])
	}
	return candidates, nil
}

// runEmbeddingModel invokes a separate, confined process for one request.
func runEmbeddingModel(ctx context.Context, script string, texts []string) ([][]float32, error) {
	request, err := json.Marshal(struct {
		Texts []string `json:"texts"`
	}{Texts: texts})
	if err != nil {
		return nil, err
	}
	modelCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	command := exec.CommandContext(modelCtx, embeddingPython, "-u", script)
	command.Stdin = bytes.NewReader(append(request, '\n'))
	command.Stderr = nil
	output, err := command.Output()
	if err != nil || len(output) > 24<<20 {
		return nil, errors.New("embedding worker unavailable or output too large")
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
		vectors[index] = decodeVector(blob)
		if len(vectors[index]) != embeddingSize {
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
func decodeVector(blob []byte) []float32 {
	if len(blob) != 4*embeddingSize {
		return nil
	}
	vector := make([]float32, embeddingSize)
	for index := range vector {
		value := math.Float32frombits(binary.LittleEndian.Uint32(blob[index*4:]))
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil
		}
		vector[index] = value
	}
	return vector
}
