package main

import (
	"context"
	"database/sql"

	"github.com/danielgtaylor/huma/v2"
)

// mediaResult is one published source record from a successful full scan.
type mediaResult struct {
	Path          string    `json:"path"`
	MediaType     string    `json:"mediaType"`
	Width         *int      `json:"width"`
	Height        *int      `json:"height"`
	LivePhotoPath *string   `json:"livePhotoPath"`
	Tags          []scanTag `json:"tags"`
}

// mediaResultsInput bounds a visible gallery's published scan rows.
type mediaResultsInput struct {
	GalleryAuth
	ID     string `path:"galleryId" format:"uuid"`
	Limit  int    `query:"limit" minimum:"1" maximum:"100"`
	Offset int    `query:"offset" minimum:"0" maximum:"100000"`
}

// mediaResultsOutput pages metadata separately from the direct File Service listing.
type mediaResultsOutput struct {
	Body struct {
		Items  []mediaResult `json:"items"`
		Total  int64         `json:"total"`
		Limit  int           `json:"limit"`
		Offset int           `json:"offset"`
	}
}

// tagsOutput combines published media tags with persistent user tags.
type tagsOutput struct {
	Body struct {
		Items []galleryTag `json:"items"`
	}
}

// mountResultRoutes provides paged scan metadata and aggregate filter options.
func mountResultRoutes(api huma.API, c galleryController) {
	huma.Get(api, "/api/galleries/{galleryId}/media-metadata", c.mediaResults)
	huma.Post(api, "/api/galleries/{galleryId}/media-metadata", c.scanUploadedMedia)
	huma.Get(api, "/api/galleries/{galleryId}/tags", c.tags)
}

// mediaResults returns only published scan rows under the gallery visibility policy.
func (c galleryController) mediaResults(ctx context.Context, input *mediaResultsInput) (*mediaResultsOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.visibleGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}
	output := &mediaResultsOutput{}
	output.Body.Items = []mediaResult{}
	output.Body.Limit, output.Body.Offset = limit, input.Offset
	if err := c.store.db.QueryRowContext(ctx, `SELECT count(*) FROM gallery_media WHERE gallery_id=?`, input.ID).Scan(&output.Body.Total); err != nil {
		return nil, huma.Error500InternalServerError("media metadata lookup failed")
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT gm.path,gm.media_type,mf.width,mf.height,gm.live_photo_path
		FROM gallery_media gm JOIN media_files mf ON mf.path=gm.path WHERE gm.gallery_id=? ORDER BY gm.path LIMIT ? OFFSET ?`, input.ID, limit, input.Offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("media metadata lookup failed")
	}
	defer rows.Close()
	for rows.Next() {
		var item mediaResult
		var width, height sql.NullInt64
		var live sql.NullString
		if err := rows.Scan(&item.Path, &item.MediaType, &width, &height, &live); err != nil {
			return nil, huma.Error500InternalServerError("media metadata lookup failed")
		}
		if width.Valid {
			value := int(width.Int64)
			item.Width = &value
		}
		if height.Valid {
			value := int(height.Int64)
			item.Height = &value
		}
		if live.Valid {
			item.LivePhotoPath = &live.String
		}
		item.Tags = []scanTag{}
		output.Body.Items = append(output.Body.Items, item)
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("media metadata lookup failed")
	}
	// Close the page cursor before reading tag rows; SQLite has one bounded connection.
	_ = rows.Close()
	for index := range output.Body.Items {
		item := &output.Body.Items[index]
		tagRows, err := c.store.db.QueryContext(ctx, `SELECT t.category,t.value FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id WHERE mt.path=? ORDER BY t.category,t.value`, item.Path)
		if err != nil {
			return nil, huma.Error500InternalServerError("media tag lookup failed")
		}
		for tagRows.Next() {
			var tag scanTag
			if err := tagRows.Scan(&tag.Category, &tag.Value); err != nil {
				_ = tagRows.Close()
				return nil, huma.Error500InternalServerError("media tag lookup failed")
			}
			item.Tags = append(item.Tags, tag)
		}
		if err := tagRows.Err(); err != nil {
			_ = tagRows.Close()
			return nil, huma.Error500InternalServerError("media tag lookup failed")
		}
		_ = tagRows.Close()
	}
	return output, nil
}

// tags summarizes both tag sources without including hidden galleries.
func (c galleryController) tags(ctx context.Context, input *galleryIDInput) (*tagsOutput, error) {
	user, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.visibleGallery(ctx, input.ID, user); err != nil {
		return nil, err
	}
	tags, err := c.galleryTags(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("tag lookup failed")
	}
	output := &tagsOutput{}
	output.Body.Items = tags
	return output, nil
}
