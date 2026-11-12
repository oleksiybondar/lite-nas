package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"lite-nas/shared/fileaccess"
)

// createPersonTagInput identifies a linked image and one owner-supplied person name.
type createPersonTagInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Path string `json:"path" required:"true" minLength:"2" maxLength:"4096"`
		Name string `json:"name" required:"true" minLength:"1" maxLength:"160"`
	}
}

// personTagBody is the durable per-file name returned after creation.
type personTagBody struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

// personTagOutput encloses a newly created person tag.
type personTagOutput struct{ Body personTagBody }

// createPersonTag adds a name only to an image linked to a gallery the caller owns.
func (c galleryController) createPersonTag(ctx context.Context, input *createPersonTagInput) (*personTagOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	path := input.Body.Path
	name := strings.TrimSpace(input.Body.Name)
	if fileaccess.ValidateHostPath(path) != nil || name == "" {
		return nil, huma.Error422UnprocessableEntity("invalid image path or person name")
	}
	var linked int
	err = c.store.db.QueryRowContext(ctx, `SELECT 1 FROM gallery_media WHERE gallery_id=? AND path=? AND media_type IN ('IMAGE','LIVE_PHOTO')`, input.ID, path).Scan(&linked)
	if err == sql.ErrNoRows {
		return nil, huma.Error404NotFound("image not found in gallery")
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("image lookup failed")
	}
	tag := personTagBody{ID: uuid.NewString(), Path: path, Name: name}
	result, err := c.store.db.ExecContext(ctx, `INSERT INTO media_person_tags(id,path,value,normalized_key,search_key,created_at)
		VALUES(?,?,?,?,?,?) ON CONFLICT(path,normalized_key) DO NOTHING`, tag.ID, path, name,
		norm.NFC.String(strings.ToLower(name)), searchKey(name), timestamp())
	if err != nil {
		return nil, huma.Error500InternalServerError("person tag creation failed")
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, huma.Error409Conflict("person name already tagged on this image")
	}
	return &personTagOutput{Body: tag}, nil
}
