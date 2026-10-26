package main

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// thumbnailInput accepts one bounded original PNG or JPEG upload.
type thumbnailInput struct {
	Authorization string `header:"Authorization"`
	AccessToken   string `cookie:"lite-nas-at"`
	ID            string `path:"galleryId" format:"uuid"`
	ContentType   string `header:"Content-Type"`
	RawBody       []byte `contentType:"application/octet-stream"`
}

// mountThumbnailRoutes documents and serves the optional uploaded cover image.
func mountThumbnailRoutes(api huma.API, c galleryController) {
	binary := &huma.MediaType{Schema: &huma.Schema{Type: "string", Format: "binary"}}
	huma.Register(api, huma.Operation{
		OperationID: "putGalleryThumbnail", Method: http.MethodPut,
		Path: "/api/galleries/{galleryId}/thumbnail/image", MaxBodyBytes: 16 << 20,
		Summary: "Upload a PNG or JPEG gallery thumbnail",
		RequestBody: &huma.RequestBody{Content: map[string]*huma.MediaType{
			"image/png": binary, "image/jpeg": binary,
		}},
	}, c.putThumbnail)
	huma.Get(api, "/api/galleries/{galleryId}/thumbnail/image", c.getThumbnail)
	huma.Delete(api, "/api/galleries/{galleryId}/thumbnail/image", c.deleteThumbnail)
}

// thumbnailName accepts only backend-issued names, even if stored state is damaged.
func thumbnailName(name string) bool {
	if filepath.Base(name) != name || strings.Contains(name, "..") {
		return false
	}
	ext := filepath.Ext(name)
	if ext != ".png" && ext != ".jpg" {
		return false
	}
	_, err := uuid.Parse(strings.TrimSuffix(name, ext))
	return err == nil
}

// putThumbnail atomically publishes a validated image and removes the previous asset.
func (c galleryController) putThumbnail(ctx context.Context, input *thumbnailInput) (*galleryOutput, error) {
	user, err := c.userID(GalleryAuth{Authorization: input.Authorization, AccessToken: input.AccessToken})
	if err != nil {
		return nil, err
	}
	if c.thumbnailDir == "" {
		return nil, huma.Error503ServiceUnavailable("thumbnail storage unavailable")
	}
	if len(input.RawBody) == 0 || len(input.RawBody) > 16<<20 {
		return nil, huma.Error422UnprocessableEntity("thumbnail image size is invalid")
	}
	declaredType, _, err := mime.ParseMediaType(input.ContentType)
	if err != nil || declaredType != "application/octet-stream" && declaredType != "image/png" && declaredType != "image/jpeg" {
		return nil, huma.Error415UnsupportedMediaType("thumbnail request type is unsupported")
	}
	mediaType := http.DetectContentType(input.RawBody)
	if declaredType != "application/octet-stream" && declaredType != mediaType {
		return nil, huma.Error415UnsupportedMediaType("thumbnail content type does not match image bytes")
	}
	ext := ""
	switch mediaType {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	default:
		return nil, huma.Error415UnsupportedMediaType("thumbnail must be PNG or JPEG")
	}
	info, _, err := image.DecodeConfig(bytes.NewReader(input.RawBody))
	if err != nil || info.Width < 1 || info.Height < 1 || info.Width > 8192 || info.Height > 8192 || int64(info.Width)*int64(info.Height) > 50_000_000 {
		return nil, huma.Error422UnprocessableEntity("thumbnail dimensions are invalid")
	}
	c.store.assetMu.Lock()
	defer c.store.assetMu.Unlock()
	g, err := c.ownedGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(c.thumbnailDir, ".upload-*")
	if err != nil {
		return nil, huma.Error503ServiceUnavailable("thumbnail storage unavailable")
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(input.RawBody); err != nil {
		_ = file.Close()
		return nil, huma.Error503ServiceUnavailable("thumbnail storage unavailable")
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return nil, huma.Error503ServiceUnavailable("thumbnail storage unavailable")
	}
	if err := file.Close(); err != nil {
		return nil, huma.Error503ServiceUnavailable("thumbnail storage unavailable")
	}
	name := uuid.NewString() + ext
	if err := os.Rename(file.Name(), filepath.Join(c.thumbnailDir, name)); err != nil {
		return nil, huma.Error503ServiceUnavailable("thumbnail storage unavailable")
	}
	old := g.ThumbnailAsset
	g.ThumbnailAsset.String, g.ThumbnailAsset.Valid = name, true
	g.ThumbnailPath = sql.NullString{}
	g.UpdatedAt = timestamp()
	if _, err := c.store.db.ExecContext(ctx, `UPDATE galleries SET thumbnail_asset=?,thumbnail_path=NULL,updated_at=? WHERE id=? AND owner_id=?`, name, g.UpdatedAt, g.ID, user); err != nil {
		_ = os.Remove(filepath.Join(c.thumbnailDir, name))
		return nil, huma.Error500InternalServerError("thumbnail update failed")
	}
	if old.Valid && thumbnailName(old.String) {
		_ = os.Remove(filepath.Join(c.thumbnailDir, old.String))
	}
	return &galleryOutput{Body: galleryView(g)}, nil
}

// getThumbnail streams one gallery-owned image under the gallery visibility policy.
func (c galleryController) getThumbnail(ctx context.Context, input *galleryIDInput) (*huma.StreamResponse, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	g, err := c.visibleGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	if c.thumbnailDir == "" || !g.ThumbnailAsset.Valid || !thumbnailName(g.ThumbnailAsset.String) {
		return nil, huma.Error404NotFound("thumbnail not found")
	}
	file, err := os.Open(filepath.Join(c.thumbnailDir, g.ThumbnailAsset.String))
	if err != nil {
		return nil, huma.Error404NotFound("thumbnail not found")
	}
	return &huma.StreamResponse{Body: func(stream huma.Context) {
		defer file.Close()
		mediaType := "image/jpeg"
		if strings.HasSuffix(g.ThumbnailAsset.String, ".png") {
			mediaType = "image/png"
		}
		stream.SetHeader("Content-Type", mediaType)
		stream.SetHeader("Cache-Control", "private, no-store")
		stream.SetHeader("X-Content-Type-Options", "nosniff")
		_, _ = io.Copy(stream.BodyWriter(), file)
	}}, nil
}

// deleteThumbnail resets only the uploaded asset, preserving the optional recipe JSON.
func (c galleryController) deleteThumbnail(ctx context.Context, input *galleryIDInput) (*noContent, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	c.store.assetMu.Lock()
	defer c.store.assetMu.Unlock()
	g, err := c.ownedGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	if _, err := c.store.db.ExecContext(ctx, `UPDATE galleries SET thumbnail_asset=NULL,updated_at=? WHERE id=? AND owner_id=?`, timestamp(), g.ID, user); err != nil {
		return nil, huma.Error500InternalServerError("thumbnail deletion failed")
	}
	if c.thumbnailDir != "" && g.ThumbnailAsset.Valid && thumbnailName(g.ThumbnailAsset.String) {
		_ = os.Remove(filepath.Join(c.thumbnailDir, g.ThumbnailAsset.String))
	}
	return &noContent{Status: 204}, nil
}
