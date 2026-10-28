package main

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// linkedGallery identifies one visible gallery containing a matching file.
type linkedGallery struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	OwnerID string `json:"ownerId"`
}

// mediaMatch identifies one physical image and its visible gallery links.
type mediaMatch struct {
	GalleryID       string          `json:"galleryId"`
	LinkedGalleries []linkedGallery `json:"linkedGalleries"`
	mediaResult
}

// mediaMatchPage has pagination independent of the gallery list.
type mediaMatchPage struct {
	Items  []mediaMatch `json:"items"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// searchMedia finds visible gallery image links whose path-keyed tags match the query.
func (c galleryController) searchMedia(ctx context.Context, user string, input *listGalleryInput, page *mediaMatchPage) error {
	page.Items = []mediaMatch{}
	page.Limit = input.MediaLimit
	if page.Limit == 0 {
		page.Limit = 50
	}
	page.Offset = input.MediaOffset
	if input.Q == "" && len(input.Location) == 0 {
		return nil
	}
	where := ` FROM gallery_media gm JOIN galleries g ON g.id=gm.gallery_id JOIN media_files mf ON mf.path=gm.path
		LEFT JOIN media_captions mc ON mc.path=gm.path
		WHERE gm.media_type IN ('IMAGE','LIVE_PHOTO') AND
		(g.owner_id=? OR g.visibility='PUBLIC' OR EXISTS (SELECT 1 FROM gallery_shares s WHERE s.gallery_id=g.id AND s.user_id=?))`
	args := []any{user, user}
	for _, filter := range []struct {
		column string
		values []string
	}{{"g.type", input.Type}, {"g.visibility", input.Visibility}, {"g.owner_id", input.OwnerID}, {"gm.media_type", input.MediaType}} {
		if len(filter.values) == 0 {
			continue
		}
		where += ` AND ` + filter.column + ` IN (` + strings.TrimSuffix(strings.Repeat("?,", len(filter.values)), ",") + `)`
		for _, value := range filter.values {
			args = append(args, value)
		}
	}
	if len(input.Location) != 0 {
		matches := make([]string, 0, len(input.Location))
		for _, location := range input.Location {
			matches = append(matches, `instr(t.search_key,?)>0`)
			args = append(args, searchKey(location))
		}
		where += ` AND EXISTS (SELECT 1 FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id
			WHERE mt.path=gm.path AND t.category='LOCATION' AND (` + strings.Join(matches, " OR ") + `))`
	}
	baseWhere := where
	baseArgs := append([]any(nil), args...)
	if input.Q != "" {
		where += ` AND (EXISTS (SELECT 1 FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id
			WHERE mt.path=gm.path AND instr(t.search_key,?)>0)
			OR EXISTS (SELECT 1 FROM media_person_tags pt WHERE pt.path=gm.path AND instr(pt.search_key,?)>0)
			OR EXISTS (SELECT 1 FROM media_captions mc WHERE mc.path=gm.path AND instr(mc.search_key,?)>0))`
		key := searchKey(input.Q)
		args = append(args, key, key, key)
	}
	if err := c.store.db.QueryRowContext(ctx, `SELECT count(DISTINCT gm.path)`+where, args...).Scan(&page.Total); err != nil {
		return huma.Error500InternalServerError("media search failed")
	}
	order := ` ORDER BY max(g.created_at) DESC,gm.path`
	queryOffset := page.Offset
	if input.Q != "" && (wantsSemanticSearch(input.Q) || page.Total == 0) {
		if paths, err := c.semanticPaths(ctx, user, input); err == nil {
			page.Total = int64(len(paths))
			if page.Offset >= len(paths) {
				return nil
			}
			end := min(page.Offset+page.Limit, len(paths))
			selected := paths[page.Offset:end]
			where = baseWhere + ` AND gm.path IN (` + strings.TrimSuffix(strings.Repeat("?,", len(selected)), ",") + `)`
			args = baseArgs
			for _, path := range selected {
				args = append(args, path)
			}
			order = ` ORDER BY CASE gm.path`
			for index, path := range selected {
				order += ` WHEN ? THEN ` + strconv.Itoa(index)
				args = append(args, path)
			}
			order += ` END`
			queryOffset = 0
		}
	}
	rows, err := c.store.db.QueryContext(ctx, `SELECT min(g.id),gm.path,max(gm.media_type),mf.width,mf.height,max(gm.live_photo_path),mc.caption`+where+
		` GROUP BY gm.path`+order+` LIMIT ? OFFSET ?`, append(args, page.Limit, queryOffset)...)
	if err != nil {
		return huma.Error500InternalServerError("media search failed")
	}
	for rows.Next() {
		var match mediaMatch
		var width, height sql.NullInt64
		var live, caption sql.NullString
		if err := rows.Scan(&match.GalleryID, &match.Path, &match.MediaType, &width, &height, &live, &caption); err != nil {
			_ = rows.Close()
			return huma.Error500InternalServerError("media search failed")
		}
		if width.Valid {
			value := int(width.Int64)
			match.Width = &value
		}
		if height.Valid {
			value := int(height.Int64)
			match.Height = &value
		}
		if live.Valid {
			match.LivePhotoPath = &live.String
		}
		if caption.Valid {
			match.Caption = &caption.String
		}
		match.Tags = []scanTag{}
		page.Items = append(page.Items, match)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return huma.Error500InternalServerError("media search failed")
	}
	_ = rows.Close()
	for index := range page.Items {
		item := &page.Items[index]
		item.LinkedGalleries = []linkedGallery{}
		links, err := c.store.db.QueryContext(ctx, `SELECT g.id,g.name,g.type,g.owner_id FROM gallery_media gm
			JOIN galleries g ON g.id=gm.gallery_id WHERE gm.path=? AND
			(g.owner_id=? OR g.visibility='PUBLIC' OR EXISTS (SELECT 1 FROM gallery_shares s WHERE s.gallery_id=g.id AND s.user_id=?))
			ORDER BY g.name,g.id`, item.Path, user, user)
		if err != nil {
			return huma.Error500InternalServerError("media gallery links failed")
		}
		for links.Next() {
			var link linkedGallery
			if err := links.Scan(&link.ID, &link.Name, &link.Type, &link.OwnerID); err != nil {
				_ = links.Close()
				return huma.Error500InternalServerError("media gallery links failed")
			}
			item.LinkedGalleries = append(item.LinkedGalleries, link)
		}
		if err := links.Err(); err != nil {
			_ = links.Close()
			return huma.Error500InternalServerError("media gallery links failed")
		}
		_ = links.Close()
		tags, err := c.store.db.QueryContext(ctx, `SELECT category,value FROM (
			SELECT t.category,t.value FROM media_file_tags mt JOIN tags t ON t.id=mt.tag_id WHERE mt.path=?
			UNION ALL SELECT 'PERSON',value FROM media_person_tags WHERE path=?
		) ORDER BY category,value`, item.Path, item.Path)
		if err != nil {
			return huma.Error500InternalServerError("media search tags failed")
		}
		for tags.Next() {
			var tag scanTag
			if err := tags.Scan(&tag.Category, &tag.Value); err != nil {
				_ = tags.Close()
				return huma.Error500InternalServerError("media search tags failed")
			}
			item.Tags = append(item.Tags, tag)
		}
		if err := tags.Err(); err != nil {
			_ = tags.Close()
			return huma.Error500InternalServerError("media search tags failed")
		}
		_ = tags.Close()
	}
	return nil
}
