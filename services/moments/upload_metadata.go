package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/shared/fileaccess"
)

// uploadedMediaInput names one completed upload in an owned album.
type uploadedMediaInput struct {
	GalleryAuth
	ID   string `path:"galleryId" format:"uuid"`
	Body struct {
		Path string `json:"path" required:"true" minLength:"2" maxLength:"4096"`
	}
}

// uploadedMediaOutput returns the metadata published for just this file.
type uploadedMediaOutput struct{ Body mediaResult }

// scanUploadedMedia extracts one authorized upload and preserves all other rows.
func (c galleryController) scanUploadedMedia(ctx context.Context, input *uploadedMediaInput) (*uploadedMediaOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	gallery, err := c.ownedGallery(ctx, input.ID, user)
	if err != nil {
		return nil, err
	}
	path := input.Body.Path
	if gallery.Type != "ALBUM" || fileaccess.ValidateHostPath(path) != nil ||
		!strings.HasPrefix(path, strings.TrimSuffix(gallery.UploadDirectory.String, "/")+"/") {
		return nil, huma.Error422UnprocessableEntity("upload is outside the album")
	}
	info, err := c.files.metadata(ctx, accessToken(input.Authorization, input.AccessToken), path)
	if err != nil {
		return nil, err
	}
	item := scanMediaFromSource(info, c.scans.places)
	enrichLocation(ctx, &item, info, c.scans.places)
	if item.Type == "IMAGE" {
		if captioner := startCaptionWorker(ctx, c.scans.places); captioner != nil {
			item.Caption = captioner.caption(ctx, c.files, accessToken(input.Authorization, input.AccessToken), path)
			captioner.close()
		}
	}
	if err := c.publishUploadedMedia(ctx, gallery, &item); err != nil {
		return nil, huma.Error500InternalServerError("uploaded metadata publication failed")
	}
	return &uploadedMediaOutput{Body: mediaResult{Path: item.Path, MediaType: item.Type, Width: item.Width, Height: item.Height, LivePhotoPath: item.LivePath, Caption: item.Caption, Tags: item.Tags}}, nil
}

// publishUploadedMedia replaces one path's derived data in a transaction.
func (c galleryController) publishUploadedMedia(ctx context.Context, gallery galleryRecord, item *scanMedia) error {
	tx, err := c.store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var currentVersion string
	if err := tx.QueryRowContext(ctx, `SELECT updated_at FROM galleries WHERE id=? AND owner_id=? AND upload_directory=?`, gallery.ID, gallery.OwnerID, gallery.UploadDirectory.String).Scan(&currentVersion); err != nil || currentVersion != gallery.UpdatedAt {
		return errors.New("album changed during upload scan")
	}
	if item.Type == "IMAGE" {
		var existingType string
		var existingMotion sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT media_type,live_photo_path FROM gallery_media WHERE gallery_id=? AND path=?`, gallery.ID, item.Path).Scan(&existingType, &existingMotion)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if existingType == "LIVE_PHOTO" && existingMotion.Valid {
			item.Type = "LIVE_PHOTO"
			item.LivePath = &existingMotion.String
			item.Tags = append(item.Tags, scanTag{Category: "MEDIA_TYPE", Value: "LIVE_PHOTO"})
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO gallery_media(gallery_id,path,media_type,live_photo_path) VALUES(?,?,?,?)
		ON CONFLICT(gallery_id,path) DO UPDATE SET media_type=excluded.media_type,width=NULL,height=NULL,live_photo_path=excluded.live_photo_path
		WHERE gallery_media.media_type IS NOT excluded.media_type OR gallery_media.live_photo_path IS NOT excluded.live_photo_path`, gallery.ID, item.Path, item.Type, item.LivePath); err != nil {
		return err
	}
	// An owned collection already referencing this source gains the same file record.
	collectionType := item.Type
	if collectionType == "LIVE_PHOTO" {
		collectionType = "IMAGE"
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO gallery_media(gallery_id,path,media_type,live_photo_path)
		SELECT DISTINCT g.id,?,?,NULL FROM galleries g JOIN gallery_references r ON r.gallery_id=g.id
		WHERE g.type='COLLECTION' AND g.owner_id=? AND
		((r.reference_type='FILE' AND r.path=?) OR
		 (r.reference_type='DIRECTORY' AND (r.path='/' OR instr(?,r.path||'/')=1)))`,
		item.Path, collectionType, gallery.OwnerID, item.Path, item.Path); err != nil {
		return err
	}
	if err := c.store.writeMediaFile(ctx, tx, *item); err != nil {
		return err
	}
	// A concurrent full scan must not replace metadata published for this upload.
	if _, err := tx.ExecContext(ctx, `UPDATE galleries SET updated_at=? WHERE id=?`, timestamp(), gallery.ID); err != nil {
		return err
	}
	return tx.Commit()
}
