package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	rbaccontract "lite-nas/shared/contracts/rbac"
)

// collectionReference is one explicit file or directory membership entry.
type collectionReference struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Path      string `json:"path"`
	CreatedAt string `json:"createdAt"`
}

// referenceInput validates one collection membership mutation.
type referenceInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Path string `json:"path" required:"true" minLength:"1" maxLength:"4096"`
	}
}

// referenceOutput returns the idempotently added or removed membership.
type referenceOutput struct{ Body collectionReference }

// referencesOutput lists the visible collection's references.
type referencesOutput struct {
	Body struct {
		Items []collectionReference `json:"items"`
	}
}

// shareInput validates a stable user identifier for membership changes.
type shareInput struct {
	GalleryAuth
	ID     string `path:"galleryId" format:"uuid"`
	UserID string `path:"userId" minLength:"1" maxLength:"128"`
}

// shareUsernameInput accepts one host login for owner-controlled sharing.
type shareUsernameInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Username string `json:"username" required:"true" minLength:"1" maxLength:"128" pattern:"^[A-Za-z0-9_][A-Za-z0-9._-]*$"`
	}
}

// shareUsernameOutput reports the resolved stable user ID after membership is added.
type shareUsernameOutput struct {
	Body struct {
		UserID string `json:"userId"`
	}
}

// sharesOutput contains the explicit members visible to the owner.
type sharesOutput struct {
	Body struct {
		UserIDs []string `json:"userIds"`
	}
}

// mountReferenceRoutes registers collection membership and read-only sharing.
func mountReferenceRoutes(api huma.API, c galleryController) {
	huma.Get(api, "/api/collections/{galleryId}/references", c.references)
	huma.Post(api, "/api/collections/{galleryId}/files", c.addFile)
	huma.Delete(api, "/api/collections/{galleryId}/files", c.removeFile)
	huma.Post(api, "/api/collections/{galleryId}/directories", c.addDirectory)
	huma.Delete(api, "/api/collections/{galleryId}/directories", c.removeDirectory)
	huma.Get(api, "/api/galleries/{galleryId}/shares", c.shares)
	huma.Post(api, "/api/galleries/{galleryId}/shares", c.addShareByUsername)
	huma.Put(api, "/api/galleries/{galleryId}/shares/{userId}", c.addShare)
	huma.Delete(api, "/api/galleries/{galleryId}/shares/{userId}", c.removeShare)
}

// addShareByUsername resolves a host login to the same stable UID used by sessions.
func (c galleryController) addShareByUsername(ctx context.Context, input *shareUsernameInput) (*shareUsernameOutput, error) {
	owner, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, owner); err != nil {
		return nil, err
	}
	if c.users == nil {
		return nil, huma.Error503ServiceUnavailable("user lookup unavailable")
	}
	var resolved rbaccontract.GetSubjectRolesResponse
	if err := c.users.Request(ctx, rbaccontract.GetSubjectRolesRPCSubject,
		rbaccontract.GetSubjectRolesRequest{Username: input.Body.Username}, &resolved); err != nil {
		return nil, huma.Error503ServiceUnavailable("user lookup unavailable")
	}
	uid, err := strconv.ParseUint(resolved.UID, 10, 32)
	if err != nil || uid == 0 || strconv.FormatUint(uid, 10) != resolved.UID {
		return nil, huma.Error404NotFound("user not found")
	}
	if resolved.UID == owner {
		return nil, huma.Error422UnprocessableEntity("owner cannot be shared with self")
	}
	if _, err := c.store.db.ExecContext(ctx,
		`INSERT INTO gallery_shares(gallery_id,user_id,created_at) VALUES(?,?,?) ON CONFLICT DO NOTHING`,
		input.ID, resolved.UID, timestamp()); err != nil {
		return nil, huma.Error500InternalServerError("share update failed")
	}
	output := &shareUsernameOutput{}
	output.Body.UserID = resolved.UID
	return output, nil
}

// references lists only membership in a gallery the caller may read.
func (c galleryController) references(ctx context.Context, input *galleryIDInput) (*referencesOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	g, err := c.visibleGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	if g.Type != "COLLECTION" {
		return nil, huma.Error422UnprocessableEntity("gallery is not a collection")
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT id,reference_type,path,created_at FROM gallery_references WHERE gallery_id=? ORDER BY reference_type,path,id`, g.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("reference lookup failed")
	}
	defer rows.Close()
	output := &referencesOutput{}
	output.Body.Items = []collectionReference{}
	for rows.Next() {
		var ref collectionReference
		if err := rows.Scan(&ref.ID, &ref.Type, &ref.Path, &ref.CreatedAt); err != nil {
			return nil, huma.Error500InternalServerError("reference lookup failed")
		}
		output.Body.Items = append(output.Body.Items, ref)
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("reference lookup failed")
	}
	return output, nil
}

// modifyReference uses one path for file/directory add/remove rules.
func (c galleryController) modifyReference(ctx context.Context, input *referenceInput, kind string, add bool) (*referenceOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	g, err := c.ownedGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	if g.Type != "COLLECTION" {
		return nil, huma.Error422UnprocessableEntity("gallery is not a collection")
	}
	path := input.Body.Path
	if err := validateGalleryPath(&path); err != nil {
		return nil, err
	}
	if add {
		expected := "file"
		if kind == "DIRECTORY" {
			expected = "directory"
		}
		path, err = c.files.stat(ctx, accessToken(input.Authorization, input.AccessToken), path, expected)
		if err != nil {
			return nil, err
		}
	}
	ref := collectionReference{ID: uuid.NewString(), Type: kind, Path: path, CreatedAt: timestamp()}
	tx, err := c.store.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("reference update failed")
	}
	defer tx.Rollback()
	changed := false
	if add {
		result, insertErr := tx.ExecContext(ctx, `INSERT INTO gallery_references(id,gallery_id,reference_type,path,created_at) VALUES(?,?,?,?,?) ON CONFLICT(gallery_id,reference_type,path) DO NOTHING`, ref.ID, g.ID, ref.Type, ref.Path, ref.CreatedAt)
		err = insertErr
		if err != nil {
			return nil, huma.Error500InternalServerError("reference update failed")
		}
		count, err := result.RowsAffected()
		if err != nil {
			return nil, huma.Error500InternalServerError("reference update failed")
		}
		changed = count != 0
		err = tx.QueryRowContext(ctx, `SELECT id,created_at FROM gallery_references WHERE gallery_id=? AND reference_type=? AND path=?`, g.ID, kind, path).Scan(&ref.ID, &ref.CreatedAt)
	} else {
		err = tx.QueryRowContext(ctx, `SELECT id,created_at FROM gallery_references WHERE gallery_id=? AND reference_type=? AND path=?`, g.ID, kind, path).Scan(&ref.ID, &ref.CreatedAt)
		if err == sql.ErrNoRows {
			return &referenceOutput{Body: collectionReference{Type: kind, Path: path}}, nil
		}
		if err == nil {
			_, err = tx.ExecContext(ctx, `DELETE FROM gallery_references WHERE id=?`, ref.ID)
			changed = true
		}
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("reference update failed")
	}
	if changed {
		if !add {
			// Unlink only media no longer covered by another explicit reference.
			if _, err := tx.ExecContext(ctx, `DELETE FROM gallery_media WHERE gallery_id=? AND NOT EXISTS (
				SELECT 1 FROM gallery_references r WHERE r.gallery_id=? AND
				((r.reference_type='FILE' AND r.path=gallery_media.path) OR
				 (r.reference_type='DIRECTORY' AND (r.path='/' OR instr(gallery_media.path,r.path||'/')=1))))`, g.ID, g.ID); err != nil {
				return nil, huma.Error500InternalServerError("reference update failed")
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE galleries SET updated_at=? WHERE id=?`, timestamp(), g.ID); err != nil {
			return nil, huma.Error500InternalServerError("reference update failed")
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, huma.Error500InternalServerError("reference update failed")
	}
	return &referenceOutput{Body: ref}, nil
}

// addFile adds one file reference without touching the original file.
func (c galleryController) addFile(ctx context.Context, input *referenceInput) (*referenceOutput, error) {
	return c.modifyReference(ctx, input, "FILE", true)
}

// removeFile removes one file reference without touching the original file.
func (c galleryController) removeFile(ctx context.Context, input *referenceInput) (*referenceOutput, error) {
	return c.modifyReference(ctx, input, "FILE", false)
}

// addDirectory adds one directory reference without reading its contents.
func (c galleryController) addDirectory(ctx context.Context, input *referenceInput) (*referenceOutput, error) {
	return c.modifyReference(ctx, input, "DIRECTORY", true)
}

// removeDirectory removes one directory reference without deleting its files.
func (c galleryController) removeDirectory(ctx context.Context, input *referenceInput) (*referenceOutput, error) {
	return c.modifyReference(ctx, input, "DIRECTORY", false)
}

// shares exposes explicit members only to the gallery owner.
func (c galleryController) shares(ctx context.Context, input *galleryIDInput) (*sharesOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT user_id FROM gallery_shares WHERE gallery_id=? ORDER BY user_id`, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("share lookup failed")
	}
	defer rows.Close()
	output := &sharesOutput{}
	output.Body.UserIDs = []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, huma.Error500InternalServerError("share lookup failed")
		}
		output.Body.UserIDs = append(output.Body.UserIDs, id)
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("share lookup failed")
	}
	return output, nil
}

// changeShare adds or removes one read-only member for the owner.
func (c galleryController) changeShare(ctx context.Context, input *shareInput, add bool) (*noContent, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.ownedGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	member := strings.TrimSpace(input.UserID)
	if member != input.UserID || member == "" || member == user {
		return nil, huma.Error422UnprocessableEntity("invalid member")
	}
	if add {
		_, err = c.store.db.ExecContext(ctx, `INSERT INTO gallery_shares(gallery_id,user_id,created_at) VALUES(?,?,?) ON CONFLICT DO NOTHING`, input.ID, member, timestamp())
	} else {
		_, err = c.store.db.ExecContext(ctx, `DELETE FROM gallery_shares WHERE gallery_id=? AND user_id=?`, input.ID, member)
	}
	if err != nil {
		return nil, huma.Error500InternalServerError("share update failed")
	}
	return &noContent{Status: 204}, nil
}

// addShare grants read-only metadata membership.
func (c galleryController) addShare(ctx context.Context, input *shareInput) (*noContent, error) {
	return c.changeShare(ctx, input, true)
}

// removeShare removes read-only metadata membership.
func (c galleryController) removeShare(ctx context.Context, input *shareInput) (*noContent, error) {
	return c.changeShare(ctx, input, false)
}
