package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"lite-nas/shared/browserauth"
	"lite-nas/shared/fileaccess"
)

// galleryController applies session ownership and visibility to gallery metadata.
type galleryController struct {
	auth         browserauth.AuthService
	store        *galleryStore
	files        fileServiceClient
	scans        *scanManager
	thumbnailDir string
	users        subjectLookup
}

// subjectLookup is the existing RBAC RPC used to resolve logins to stable IDs.
type subjectLookup interface {
	Request(context.Context, string, any, any) error
}

// GalleryAuth is the existing header-or-cookie access token boundary.
type GalleryAuth struct {
	Authorization string `header:"Authorization"`
	AccessToken   string `cookie:"lite-nas-at"`
}

// galleryBody is the public metadata view; original file bytes remain in File Service.
type galleryBody struct {
	ID                  string          `json:"id"`
	Type                string          `json:"type"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	OwnerID             string          `json:"ownerId"`
	Visibility          string          `json:"visibility"`
	UploadDirectory     *string         `json:"uploadDirectory"`
	WallStrategy        string          `json:"wallRenderStrategy"`
	ThumbnailGeneration json.RawMessage `json:"thumbnailGeneration,omitempty"`
	ThumbnailImageURL   *string         `json:"thumbnailImageUrl"`
	ThumbnailPath       *string         `json:"thumbnailPath"`
	Tags                []galleryTag    `json:"tags"`
	CreatedAt           string          `json:"createdAt"`
	UpdatedAt           string          `json:"updatedAt"`
}

// galleryOutput encloses one gallery in a stable JSON response.
type galleryOutput struct{ Body galleryBody }

// galleriesOutput contains a deterministic, paginated accessible list.
type galleriesOutput struct {
	Body struct {
		Items         []galleryBody     `json:"items"`
		Metadata      galleryPagination `json:"metadata"`
		MatchingMedia mediaMatchPage    `json:"matchingMedia"`
	}
}

// galleryPagination follows the page and size metadata used by LiteNAS lists.
type galleryPagination struct {
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int64 `json:"total_pages"`
}

// createGalleryInput validates the editable gallery fields at the HTTP boundary.
type createGalleryInput struct {
	GalleryAuth
	Body struct {
		Type                string          `json:"type" required:"true" enum:"ALBUM,COLLECTION"`
		Name                string          `json:"name" required:"true" minLength:"1" maxLength:"160"`
		Description         string          `json:"description,omitempty" maxLength:"4000"`
		Visibility          string          `json:"visibility,omitempty" enum:"PRIVATE,PUBLIC"`
		UploadDirectory     *string         `json:"uploadDirectory,omitempty" maxLength:"4096"`
		WallStrategy        string          `json:"wallRenderStrategy,omitempty" enum:"mosaic,justified,masonry,grid"`
		ThumbnailGeneration json.RawMessage `json:"thumbnailGeneration,omitempty"`
		ThumbnailPath       *string         `json:"thumbnailPath,omitempty" maxLength:"4096"`
	}
}

// galleryIDInput validates a gallery identifier and current-session credentials.
type galleryIDInput struct {
	GalleryAuth
	ID string `path:"galleryId" format:"uuid"`
}

// listGalleryInput bounds search and pagination before querying SQLite.
type listGalleryInput struct {
	GalleryAuth
	Page        int      `query:"page" default:"1" minimum:"1" maximum:"100000"`
	Size        int      `query:"size" default:"20" minimum:"1" maximum:"500"`
	Q           string   `query:"q" maxLength:"160"`
	Type        []string `query:"type,explode"`
	Visibility  []string `query:"visibility,explode"`
	OwnerID     []string `query:"ownerId,explode"`
	MediaType   []string `query:"mediaType,explode"`
	Location    []string `query:"location,explode"`
	MediaLimit  int      `query:"mediaLimit" default:"50" minimum:"1" maximum:"100"`
	MediaOffset int      `query:"mediaOffset" default:"0" minimum:"0" maximum:"100000"`
}

// patchGalleryInput permits owner-controlled mutable metadata only.
type patchGalleryInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Name                *string         `json:"name,omitempty" minLength:"1" maxLength:"160"`
		Description         *string         `json:"description,omitempty" maxLength:"4000"`
		Visibility          *string         `json:"visibility,omitempty" enum:"PRIVATE,PUBLIC"`
		UploadDirectory     *string         `json:"uploadDirectory,omitempty" maxLength:"4096"`
		WallStrategy        *string         `json:"wallRenderStrategy,omitempty" enum:"mosaic,justified,masonry,grid"`
		ThumbnailGeneration json.RawMessage `json:"thumbnailGeneration,omitempty"`
		ThumbnailPath       *string         `json:"thumbnailPath,omitempty" maxLength:"4096"`
	}
}

// noContent represents a successful metadata-only deletion.
type noContent struct{ Status int }

// mountGalleryRoutes registers the reviewed core gallery metadata contract.
func mountGalleryRoutes(api huma.API, controller galleryController) {
	huma.Register(api, huma.Operation{
		OperationID: "createGallery", Method: http.MethodPost,
		Path: "/api/galleries", MaxBodyBytes: 17 << 20,
	}, controller.create)
	huma.Get(api, "/api/galleries", controller.list)
	huma.Get(api, "/api/galleries/{galleryId}", controller.get)
	huma.Register(api, huma.Operation{
		OperationID: "patchGallery", Method: http.MethodPatch,
		Path: "/api/galleries/{galleryId}", MaxBodyBytes: 17 << 20,
	}, controller.patch)
	huma.Delete(api, "/api/galleries/{galleryId}", controller.delete)
	mountReferenceRoutes(api, controller)
	mountScanRoutes(api, controller)
	mountThumbnailRoutes(api, controller)
	mountResultRoutes(api, controller)
	mountUserTagRoutes(api, controller)
	mountFriendRoutes(api, controller)
}

// userID verifies the current LiteNAS session without accepting a caller-supplied owner.
func (c galleryController) userID(auth GalleryAuth) (string, error) {
	session, err := c.auth.Me(time.Now(), accessToken(auth.Authorization, auth.AccessToken))
	if err != nil || session.UserID == "" {
		return "", huma.Error401Unauthorized("invalid authentication")
	}
	return session.UserID, nil
}

// validateGalleryPath preserves File Service's clean absolute path namespace.
func validateGalleryPath(path *string) error {
	if path == nil || fileaccess.ValidateHostPath(*path) != nil {
		return huma.Error422UnprocessableEntity("path must be clean and absolute")
	}
	return nil
}

// resolveThumbnailPath accepts only an image file readable by the owner through File Service.
func (c galleryController) resolveThumbnailPath(ctx context.Context, auth GalleryAuth, path *string) (*string, error) {
	if path == nil {
		return nil, nil
	}
	if err := validateGalleryPath(path); err != nil {
		return nil, err
	}
	token := accessToken(auth.Authorization, auth.AccessToken)
	resolved, err := c.files.stat(ctx, token, *path, "file")
	if err != nil {
		return nil, err
	}
	metadata, err := c.files.metadata(ctx, token, resolved)
	if err != nil {
		return nil, err
	}
	if metadata.MediaType != "IMAGE" {
		return nil, huma.Error422UnprocessableEntity("thumbnail path must be an image")
	}
	return &resolved, nil
}

// validateThumbnailGeneration bounds opaque frontend JSON without interpreting it.
func validateThumbnailGeneration(value json.RawMessage) error {
	if len(value) == 0 {
		return nil
	}
	if len(value) > 16<<20 {
		return huma.NewError(http.StatusRequestEntityTooLarge, "thumbnail generation is too large")
	}
	if !json.Valid(value) {
		return huma.Error422UnprocessableEntity("thumbnail generation is invalid JSON")
	}
	depth, inString, escape := 0, false, false
	for _, char := range value {
		if inString {
			if escape {
				escape = false
			} else if char == '\\' {
				escape = true
			} else if char == '"' {
				inString = false
			}
			continue
		}
		switch char {
		case '"':
			inString = true
		case '{', '[':
			depth++
			if depth > 64 {
				return huma.Error422UnprocessableEntity("thumbnail generation nesting is too deep")
			}
		case '}', ']':
			depth--
		}
	}
	return nil
}

// galleryView shapes nullable persisted values and avoids leaking share identities.
func galleryView(g galleryRecord) galleryBody {
	view := galleryBody{
		ID: g.ID, Type: g.Type, Name: g.Name, Description: g.Description,
		OwnerID: g.OwnerID, Visibility: g.Visibility, WallStrategy: g.WallStrategy,
		CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
	if g.UploadDirectory.Valid {
		view.UploadDirectory = &g.UploadDirectory.String
	}
	if g.ThumbnailJSON.Valid {
		view.ThumbnailGeneration = json.RawMessage(g.ThumbnailJSON.String)
	}
	if g.ThumbnailAsset.Valid {
		url := "/api/galleries/" + g.ID + "/thumbnail/image"
		view.ThumbnailImageURL = &url
	}
	if g.ThumbnailPath.Valid {
		view.ThumbnailPath = &g.ThumbnailPath.String
	}
	return view
}

// canRead checks the authenticated-only public/share/owner policy.
func (c galleryController) canRead(ctx context.Context, g galleryRecord, user string) bool {
	if g.OwnerID == user || g.Visibility == "PUBLIC" {
		return true
	}
	var exists int
	err := c.store.db.QueryRowContext(ctx, `SELECT 1 FROM gallery_shares WHERE gallery_id=? AND user_id=?`, g.ID, user).Scan(&exists)
	return err == nil
}

// visibleGallery returns 404 for both unknown and hidden gallery IDs.
func (c galleryController) visibleGallery(ctx context.Context, id, user string) (galleryRecord, error) {
	g, err := c.store.galleryByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return galleryRecord{}, huma.Error404NotFound("gallery not found")
	}
	if err != nil {
		return galleryRecord{}, huma.Error500InternalServerError("gallery lookup failed")
	}
	if !c.canRead(ctx, g, user) {
		return galleryRecord{}, huma.Error404NotFound("gallery not found")
	}
	return g, nil
}

// ownedGallery applies the owner-only mutation policy without private enumeration.
func (c galleryController) ownedGallery(ctx context.Context, id, user string) (galleryRecord, error) {
	g, err := c.visibleGallery(ctx, id, user)
	if err != nil {
		return galleryRecord{}, err
	}
	if g.OwnerID != user {
		return galleryRecord{}, huma.Error403Forbidden("gallery owner required")
	}
	return g, nil
}

// create persists one album or collection without touching original files.
func (c galleryController) create(ctx context.Context, input *createGalleryInput) (*galleryOutput, error) {
	owner, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Body.Name)
	if name == "" {
		return nil, huma.Error422UnprocessableEntity("name is required")
	}
	if err := validateThumbnailGeneration(input.Body.ThumbnailGeneration); err != nil {
		return nil, err
	}
	thumbnailPath, err := c.resolveThumbnailPath(ctx, input.GalleryAuth, input.Body.ThumbnailPath)
	if err != nil {
		return nil, err
	}
	if input.Body.Type == "ALBUM" {
		if err := validateGalleryPath(input.Body.UploadDirectory); err != nil {
			return nil, err
		}
		resolved, err := c.files.stat(ctx, accessToken(input.Authorization, input.AccessToken), *input.Body.UploadDirectory, "directory")
		if err != nil {
			return nil, err
		}
		input.Body.UploadDirectory = &resolved
	} else if input.Body.UploadDirectory != nil {
		return nil, huma.Error422UnprocessableEntity("collections have no upload directory")
	}
	visibility := input.Body.Visibility
	if visibility == "" {
		visibility = "PUBLIC"
	}
	strategy := input.Body.WallStrategy
	if strategy == "" {
		strategy = "mosaic"
	}
	now := timestamp()
	g := galleryRecord{
		ID: uuid.NewString(), Type: input.Body.Type, Name: name,
		Description: input.Body.Description, OwnerID: owner, Visibility: visibility,
		WallStrategy: strategy, CreatedAt: now, UpdatedAt: now,
	}
	if input.Body.UploadDirectory != nil {
		g.UploadDirectory = sql.NullString{String: *input.Body.UploadDirectory, Valid: true}
	}
	if len(input.Body.ThumbnailGeneration) > 0 {
		g.ThumbnailJSON = sql.NullString{String: string(input.Body.ThumbnailGeneration), Valid: true}
	}
	if thumbnailPath != nil {
		g.ThumbnailPath = sql.NullString{String: *thumbnailPath, Valid: true}
	}
	_, err = c.store.db.ExecContext(ctx, `INSERT INTO galleries (`+galleryColumns+`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		g.ID, g.Type, g.Name, g.Description, g.OwnerID, g.Visibility, g.UploadDirectory,
		g.WallStrategy, g.ThumbnailJSON, g.ThumbnailAsset, g.ThumbnailPath, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		return nil, huma.Error500InternalServerError("gallery creation failed")
	}
	return c.galleryOutput(ctx, g)
}

// galleryOutput attaches both tag sources to a gallery metadata response.
func (c galleryController) galleryOutput(ctx context.Context, g galleryRecord) (*galleryOutput, error) {
	tags, err := c.galleryTags(ctx, g.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("gallery tag lookup failed")
	}
	view := galleryView(g)
	view.Tags = tags
	return &galleryOutput{Body: view}, nil
}

// get reads one gallery after enforcing metadata visibility.
func (c galleryController) get(ctx context.Context, input *galleryIDInput) (*galleryOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	g, err := c.visibleGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	return c.galleryOutput(ctx, g)
}

// list pages only galleries visible to the authenticated user.
func (c galleryController) list(ctx context.Context, input *listGalleryInput) (*galleriesOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	for _, filter := range []struct {
		values  []string
		allowed map[string]bool
	}{
		{input.Type, map[string]bool{"ALBUM": true, "COLLECTION": true}},
		{input.Visibility, map[string]bool{"PRIVATE": true, "PUBLIC": true}},
		{input.MediaType, map[string]bool{"IMAGE": true, "VIDEO": true, "AUDIO": true, "OTHER": true, "LIVE_PHOTO": true}},
		{input.OwnerID, nil},
		{input.Location, nil},
	} {
		if len(filter.values) > 10 {
			return nil, huma.Error422UnprocessableEntity("too many filter values")
		}
		seen := map[string]bool{}
		for _, value := range filter.values {
			if value == "" || len(value) > 160 || strings.TrimSpace(value) != value || seen[value] || (filter.allowed != nil && !filter.allowed[value]) {
				return nil, huma.Error422UnprocessableEntity("invalid filter value")
			}
			seen[value] = true
		}
	}
	page, size := input.Page, input.Size
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 20
	}
	offset := (page - 1) * size
	where := ` FROM galleries g WHERE (g.owner_id=? OR g.visibility='PUBLIC' OR EXISTS (SELECT 1 FROM gallery_shares s WHERE s.gallery_id=g.id AND s.user_id=?))`
	args := []any{user, user}
	appendFilter := func(column string, values []string) {
		if len(values) == 0 {
			return
		}
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
		where += ` AND ` + column + ` IN (` + placeholders + `)`
		for _, value := range values {
			args = append(args, value)
		}
	}
	appendFilter("g.type", input.Type)
	appendFilter("g.visibility", input.Visibility)
	appendFilter("g.owner_id", input.OwnerID)
	if len(input.MediaType) > 0 {
		where += ` AND EXISTS (SELECT 1 FROM gallery_media m WHERE m.gallery_id=g.id AND m.media_type IN (` + strings.TrimSuffix(strings.Repeat("?,", len(input.MediaType)), ",") + `))`
		for _, value := range input.MediaType {
			args = append(args, value)
		}
	}
	if len(input.Location) > 0 {
		matches := make([]string, 0, len(input.Location))
		for _, value := range input.Location {
			matches = append(matches, `instr(t.search_key,?)>0`)
			args = append(args, searchKey(value))
		}
		where += ` AND EXISTS (SELECT 1 FROM gallery_media gm JOIN media_file_tags mt ON mt.path=gm.path JOIN tags t ON t.id=mt.tag_id WHERE gm.gallery_id=g.id AND t.category='LOCATION' AND (` + strings.Join(matches, " OR ") + `))`
	}
	if input.Q != "" {
		where += ` AND (instr(lower(g.name),lower(?))>0 OR instr(lower(g.description),lower(?))>0
			OR EXISTS (SELECT 1 FROM gallery_user_tags ut WHERE ut.gallery_id=g.id AND instr(ut.normalized_key,?)>0)
			OR EXISTS (SELECT 1 FROM gallery_media gm JOIN media_file_tags mt ON mt.path=gm.path JOIN tags t ON t.id=mt.tag_id WHERE gm.gallery_id=g.id AND instr(t.search_key,?)>0)
			OR EXISTS (SELECT 1 FROM gallery_media gm JOIN media_person_tags pt ON pt.path=gm.path WHERE gm.gallery_id=g.id AND instr(pt.search_key,?)>0)
			OR EXISTS (SELECT 1 FROM gallery_media gm JOIN media_captions mc ON mc.path=gm.path WHERE gm.gallery_id=g.id AND instr(mc.search_key,?)>0))`
		key := norm.NFC.String(strings.ToLower(input.Q))
		args = append(args, input.Q, input.Q, key, searchKey(input.Q), searchKey(input.Q), searchKey(input.Q))
	}
	output := &galleriesOutput{}
	output.Body.Items = []galleryBody{}
	output.Body.Metadata.Page, output.Body.Metadata.Size = page, size
	if err := c.store.db.QueryRowContext(ctx, `SELECT count(*)`+where, args...).Scan(&output.Body.Metadata.TotalCount); err != nil {
		return nil, huma.Error500InternalServerError("gallery list failed")
	}
	output.Body.Metadata.TotalPages = (output.Body.Metadata.TotalCount + int64(size) - 1) / int64(size)
	rows, err := c.store.db.QueryContext(ctx, `SELECT `+galleryColumns+where+` ORDER BY g.created_at DESC,g.id DESC LIMIT ? OFFSET ?`, append(args, size, offset)...)
	if err != nil {
		return nil, huma.Error500InternalServerError("gallery list failed")
	}
	defer rows.Close()
	for rows.Next() {
		var g galleryRecord
		if err := rows.Scan(&g.ID, &g.Type, &g.Name, &g.Description, &g.OwnerID, &g.Visibility,
			&g.UploadDirectory, &g.WallStrategy, &g.ThumbnailJSON, &g.ThumbnailAsset, &g.ThumbnailPath, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, huma.Error500InternalServerError("gallery list failed")
		}
		output.Body.Items = append(output.Body.Items, galleryView(g))
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("gallery list failed")
	}
	// Release the page cursor before looking up tags on the single SQLite connection.
	_ = rows.Close()
	for index := range output.Body.Items {
		tags, err := c.galleryTags(ctx, output.Body.Items[index].ID)
		if err != nil {
			return nil, huma.Error500InternalServerError("gallery tag lookup failed")
		}
		output.Body.Items[index].Tags = tags
	}
	if err := c.searchMedia(ctx, user, input, &output.Body.MatchingMedia); err != nil {
		return nil, err
	}
	return output, nil
}

// patch updates one owner's gallery metadata while preserving immutable identity.
func (c galleryController) patch(ctx context.Context, input *patchGalleryInput) (*galleryOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	g, err := c.ownedGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	oldDirectory := g.UploadDirectory
	if err := validateThumbnailGeneration(input.Body.ThumbnailGeneration); err != nil {
		return nil, err
	}
	if input.Body.Name != nil {
		g.Name = strings.TrimSpace(*input.Body.Name)
		if g.Name == "" {
			return nil, huma.Error422UnprocessableEntity("name is required")
		}
	}
	if input.Body.Description != nil {
		g.Description = *input.Body.Description
	}
	if input.Body.Visibility != nil {
		g.Visibility = *input.Body.Visibility
	}
	if input.Body.UploadDirectory != nil {
		if g.Type != "ALBUM" {
			return nil, huma.Error422UnprocessableEntity("collections have no upload directory")
		}
		if err := validateGalleryPath(input.Body.UploadDirectory); err != nil {
			return nil, err
		}
		resolved, err := c.files.stat(ctx, accessToken(input.Authorization, input.AccessToken), *input.Body.UploadDirectory, "directory")
		if err != nil {
			return nil, err
		}
		g.UploadDirectory = sql.NullString{String: resolved, Valid: true}
	}
	if input.Body.WallStrategy != nil {
		g.WallStrategy = *input.Body.WallStrategy
	}
	if len(input.Body.ThumbnailGeneration) > 0 {
		g.ThumbnailJSON = sql.NullString{String: string(input.Body.ThumbnailGeneration), Valid: true}
	}
	if input.Body.ThumbnailPath != nil {
		resolved, err := c.resolveThumbnailPath(ctx, input.GalleryAuth, input.Body.ThumbnailPath)
		if err != nil {
			return nil, err
		}
		g.ThumbnailPath = sql.NullString{String: *resolved, Valid: true}
	}
	g.UpdatedAt = timestamp()
	tx, err := c.store.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("gallery update failed")
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `UPDATE galleries SET name=?,description=?,visibility=?,upload_directory=?,wall_strategy=?,thumbnail_json=?,thumbnail_path=?,updated_at=? WHERE id=? AND owner_id=?`,
		g.Name, g.Description, g.Visibility, g.UploadDirectory, g.WallStrategy, g.ThumbnailJSON, g.ThumbnailPath, g.UpdatedAt, g.ID, user)
	if err != nil {
		return nil, huma.Error500InternalServerError("gallery update failed")
	}
	if oldDirectory != g.UploadDirectory {
		if _, err := tx.ExecContext(ctx, `DELETE FROM gallery_media WHERE gallery_id=? AND instr(path,?)!=1`, g.ID, strings.TrimSuffix(g.UploadDirectory.String, "/")+"/"); err != nil {
			return nil, huma.Error500InternalServerError("gallery update failed")
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, huma.Error500InternalServerError("gallery update failed")
	}
	return c.galleryOutput(ctx, g)
}

// delete removes metadata and dependent rows without deleting original files.
func (c galleryController) delete(ctx context.Context, input *galleryIDInput) (*noContent, error) {
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
	if _, err := c.store.db.ExecContext(ctx, `DELETE FROM galleries WHERE id=? AND owner_id=?`, input.ID, user); err != nil {
		return nil, huma.Error500InternalServerError("gallery deletion failed")
	}
	if c.thumbnailDir != "" && g.ThumbnailAsset.Valid && thumbnailName(g.ThumbnailAsset.String) {
		_ = os.Remove(filepath.Join(c.thumbnailDir, g.ThumbnailAsset.String))
	}
	return &noContent{Status: 204}, nil
}
