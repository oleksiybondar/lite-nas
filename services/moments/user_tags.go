package main

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

// galleryTag describes either one gallery-level user tag or an aggregate scan tag.
// Count is the number of tagged media for AUTO tags and one for USER tags.
type galleryTag struct {
	ID       string `json:"id"`
	Type     string `json:"type" enum:"AUTO,USER"`
	Category string `json:"category"`
	Value    string `json:"value"`
	Count    int    `json:"count"`
}

// userTagBody is the editable gallery-level tag returned by the CRUD routes.
type userTagBody struct {
	ID        string `json:"id"`
	GalleryID string `json:"galleryId"`
	Type      string `json:"type" enum:"USER"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// userTagOutput encloses one created or updated user tag.
type userTagOutput struct{ Body userTagBody }

// userTagsOutput lists user-owned tags independently of derived media tags.
type userTagsOutput struct {
	Body struct {
		Items []userTagBody `json:"items"`
	}
}

// createUserTagInput validates the gallery path and editable value.
type createUserTagInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Value string `json:"value" required:"true" minLength:"1" maxLength:"160"`
	}
}

// userTagIDInput validates both UUIDs before a delete request reaches storage.
type userTagIDInput struct {
	GalleryAuth
	ID    string `path:"galleryId" format:"uuid"`
	TagID string `path:"tagId" format:"uuid"`
}

// patchUserTagInput validates both UUIDs and the replacement tag value.
type patchUserTagInput struct {
	GalleryAuth
	ID    string `path:"galleryId" format:"uuid"`
	TagID string `path:"tagId" format:"uuid"`
	Body  struct {
		Value string `json:"value" required:"true" minLength:"1" maxLength:"160"`
	}
}

// mountUserTagRoutes registers gallery-level tag CRUD in the OpenAPI document.
func mountUserTagRoutes(api huma.API, c galleryController) {
	huma.Get(api, "/api/galleries/{galleryId}/user-tags", c.listUserTags)
	huma.Post(api, "/api/galleries/{galleryId}/user-tags", c.createUserTag)
	huma.Get(api, "/api/galleries/{galleryId}/user-tags/{tagId}", c.getUserTag)
	huma.Patch(api, "/api/galleries/{galleryId}/user-tags/{tagId}", c.patchUserTag)
	huma.Delete(api, "/api/galleries/{galleryId}/user-tags/{tagId}", c.deleteUserTag)
}

// normalizedUserTag validates a nonblank label and computes its duplicate key.
func normalizedUserTag(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", huma.Error422UnprocessableEntity("tag value is required")
	}
	return value, norm.NFC.String(strings.ToLower(value)), nil
}

// galleryTags loads both durable user tags and published media tags for one gallery.
func (c galleryController) galleryTags(ctx context.Context, galleryID string) ([]galleryTag, error) {
	rows, err := c.store.db.QueryContext(ctx, `
		SELECT t.id,'AUTO',t.category,t.value,count(*)
		FROM gallery_media gm JOIN media_file_tags mt ON mt.path=gm.path JOIN tags t ON t.id=mt.tag_id
		WHERE gm.gallery_id=? GROUP BY t.id
		UNION ALL
		SELECT ut.id,'USER','USER',ut.value,1
		FROM gallery_user_tags ut WHERE ut.gallery_id=?
		UNION ALL
		SELECT pt.id,'USER','PERSON',pt.value,count(*)
		FROM gallery_media gm JOIN media_person_tags pt ON pt.path=gm.path
		WHERE gm.gallery_id=? GROUP BY pt.id
		ORDER BY 2,3,4`, galleryID, galleryID, galleryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []galleryTag{}
	for rows.Next() {
		var tag galleryTag
		if err := rows.Scan(&tag.ID, &tag.Type, &tag.Category, &tag.Value, &tag.Count); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// listUserTags returns gallery-level tags to anyone who can read the gallery.
func (c galleryController) listUserTags(ctx context.Context, input *galleryIDInput) (*userTagsOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.visibleGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT id,gallery_id,value,created_at,updated_at FROM gallery_user_tags WHERE gallery_id=? ORDER BY value,id`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("user tag lookup failed")
	}
	defer rows.Close()
	output := &userTagsOutput{}
	output.Body.Items = []userTagBody{}
	for rows.Next() {
		var tag userTagBody
		if err := rows.Scan(&tag.ID, &tag.GalleryID, &tag.Value, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
			return nil, huma.Error500InternalServerError("user tag lookup failed")
		}
		tag.Type = "USER"
		output.Body.Items = append(output.Body.Items, tag)
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("user tag lookup failed")
	}
	return output, nil
}

// getUserTag returns one gallery-level tag under the gallery visibility policy.
func (c galleryController) getUserTag(ctx context.Context, input *userTagIDInput) (*userTagOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.visibleGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	var tag userTagBody
	err = c.store.db.QueryRowContext(ctx, `SELECT id,gallery_id,value,created_at,updated_at FROM gallery_user_tags WHERE id=? AND gallery_id=?`, input.TagID, input.ID).Scan(&tag.ID, &tag.GalleryID, &tag.Value, &tag.CreatedAt, &tag.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, huma.Error404NotFound("user tag not found")
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("user tag lookup failed")
	}
	tag.Type = "USER"
	return &userTagOutput{Body: tag}, nil
}

// createUserTag persists an owner-supplied tag outside the scan replacement set.
func (c galleryController) createUserTag(ctx context.Context, input *createUserTagInput) (*userTagOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	value, key, err := normalizedUserTag(input.Body.Value)
	if err != nil {
		return nil, err
	}
	now := timestamp()
	tag := userTagBody{ID: uuid.NewString(), GalleryID: input.ID, Type: "USER", Value: value, CreatedAt: now, UpdatedAt: now}
	result, err := c.store.db.ExecContext(ctx, `INSERT INTO gallery_user_tags(id,gallery_id,value,normalized_key,created_at,updated_at) VALUES (?,?,?,?,?,?) ON CONFLICT(gallery_id,normalized_key) DO NOTHING`, tag.ID, input.ID, value, key, now, now)
	if err != nil {
		return nil, huma.Error500InternalServerError("user tag creation failed")
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, huma.Error409Conflict("user tag already exists")
	}
	return &userTagOutput{Body: tag}, nil
}

// patchUserTag changes one owned gallery's user tag without modifying scan tags.
func (c galleryController) patchUserTag(ctx context.Context, input *patchUserTagInput) (*userTagOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	value, key, err := normalizedUserTag(input.Body.Value)
	if err != nil {
		return nil, err
	}
	var tag userTagBody
	err = c.store.db.QueryRowContext(ctx, `SELECT id,gallery_id,value,created_at,updated_at FROM gallery_user_tags WHERE id=? AND gallery_id=?`, input.TagID, input.ID).Scan(&tag.ID, &tag.GalleryID, &tag.Value, &tag.CreatedAt, &tag.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, huma.Error404NotFound("user tag not found")
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("user tag lookup failed")
	}
	var duplicate int
	err = c.store.db.QueryRowContext(ctx, `SELECT 1 FROM gallery_user_tags WHERE gallery_id=? AND normalized_key=? AND id<>?`, input.ID, key, input.TagID).Scan(&duplicate)
	if err == nil {
		return nil, huma.Error409Conflict("user tag already exists")
	}
	if err != sql.ErrNoRows {
		return nil, huma.Error500InternalServerError("user tag lookup failed")
	}
	tag.Value, tag.UpdatedAt, tag.Type = value, timestamp(), "USER"
	result, err := c.store.db.ExecContext(ctx, `UPDATE OR IGNORE gallery_user_tags SET value=?,normalized_key=?,updated_at=? WHERE id=? AND gallery_id=?`, tag.Value, key, tag.UpdatedAt, tag.ID, tag.GalleryID)
	if err != nil {
		return nil, huma.Error500InternalServerError("user tag update failed")
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, huma.Error409Conflict("user tag already exists")
	}
	return &userTagOutput{Body: tag}, nil
}

// deleteUserTag removes only the requested owner-supplied gallery tag.
func (c galleryController) deleteUserTag(ctx context.Context, input *userTagIDInput) (*noContent, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	result, err := c.store.db.ExecContext(ctx, `DELETE FROM gallery_user_tags WHERE id=? AND gallery_id=?`, input.TagID, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("user tag deletion failed")
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, huma.Error404NotFound("user tag not found")
	}
	return &noContent{Status: http.StatusNoContent}, nil
}
