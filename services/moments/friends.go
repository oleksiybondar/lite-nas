package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// friendRecord stores one user-maintained Unix login and optional display details.
type friendRecord struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// friendsOutput returns only the authenticated user's saved friend entries.
type friendsOutput struct {
	Body struct {
		Items    []friendRecord `json:"items"`
		Metadata friendPage     `json:"metadata"`
	}
}

// friendPage reports the count and page coordinates for saved contacts.
type friendPage struct {
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int64 `json:"total_pages"`
}

// friendInput validates friend metadata without resolving the login to an account.
type friendInput struct {
	GalleryAuth
	Username string `path:"username" minLength:"1" maxLength:"128" pattern:"^[A-Za-z0-9_][A-Za-z0-9._-]*$"`
	Body     struct {
		FirstName string `json:"firstName,omitempty" maxLength:"80"`
		LastName  string `json:"lastName,omitempty" maxLength:"80"`
		Nickname  string `json:"nickname,omitempty" maxLength:"80"`
	}
}

// friendPathInput validates a saved login for a bodyless removal request.
type friendPathInput struct {
	GalleryAuth
	Username string `path:"username" minLength:"1" maxLength:"128" pattern:"^[A-Za-z0-9_][A-Za-z0-9._-]*$"`
}

// friendListInput validates search and pagination for the authenticated saved-contact list.
type friendListInput struct {
	GalleryAuth
	Page int    `query:"page" default:"1" minimum:"1" maximum:"100000"`
	Size int    `query:"size" default:"20" minimum:"1" maximum:"500"`
	Q    string `query:"q" maxLength:"160"`
}

// mountFriendRoutes registers owner-scoped saved contacts without account discovery.
func mountFriendRoutes(api huma.API, controller galleryController) {
	huma.Get(api, "/api/friends", controller.friends)
	huma.Put(api, "/api/friends/{username}", controller.saveFriend)
	huma.Delete(api, "/api/friends/{username}", controller.deleteFriend)
}

// friends returns one searchable page of the caller's contacts, including unknown logins.
func (c galleryController) friends(ctx context.Context, input *friendListInput) (*friendsOutput, error) {
	owner, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	query := strings.TrimSpace(input.Q)
	where := ` WHERE owner_id=?`
	args := []any{owner}
	if query != "" {
		where += ` AND (username LIKE ? ESCAPE '\' OR first_name LIKE ? ESCAPE '\' OR last_name LIKE ? ESCAPE '\' OR nickname LIKE ? ESCAPE '\')`
		pattern := "%" + escapeFriendSearch(query) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	var totalCount int64
	if err := c.store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM friends`+where, args...).Scan(&totalCount); err != nil {
		return nil, huma.Error500InternalServerError("friend lookup failed")
	}
	page, size := input.Page, input.Size
	offset := (page - 1) * size
	args = append(args, size, offset)
	rows, err := c.store.db.QueryContext(ctx, `SELECT username,first_name,last_name,nickname,created_at,updated_at FROM friends`+where+` ORDER BY COALESCE(NULLIF(nickname,''),username),username LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, huma.Error500InternalServerError("friend lookup failed")
	}
	defer rows.Close()
	output := &friendsOutput{}
	output.Body.Items = []friendRecord{}
	output.Body.Metadata = friendPage{Page: page, Size: size, TotalCount: totalCount, TotalPages: (totalCount + int64(size) - 1) / int64(size)}
	for rows.Next() {
		var friend friendRecord
		if err := rows.Scan(&friend.Username, &friend.FirstName, &friend.LastName, &friend.Nickname, &friend.CreatedAt, &friend.UpdatedAt); err != nil {
			return nil, huma.Error500InternalServerError("friend lookup failed")
		}
		output.Body.Items = append(output.Body.Items, friend)
	}
	if rows.Err() != nil {
		return nil, huma.Error500InternalServerError("friend lookup failed")
	}
	return output, nil
}

// escapeFriendSearch makes LIKE wildcard characters literal in a contact search.
func escapeFriendSearch(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

// saveFriend inserts or updates a login and display details without testing account existence.
func (c galleryController) saveFriend(ctx context.Context, input *friendInput) (*noContent, error) {
	owner, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(input.Username)
	if username == "" || username == owner {
		return nil, huma.Error422UnprocessableEntity("invalid friend login")
	}
	now := timestamp()
	_, err = c.store.db.ExecContext(ctx, `INSERT INTO friends(owner_id,username,first_name,last_name,nickname,created_at,updated_at) VALUES(?,?,?,?,?,?,?) ON CONFLICT(owner_id,username) DO UPDATE SET first_name=excluded.first_name,last_name=excluded.last_name,nickname=excluded.nickname,updated_at=excluded.updated_at`, owner, username, strings.TrimSpace(input.Body.FirstName), strings.TrimSpace(input.Body.LastName), strings.TrimSpace(input.Body.Nickname), now, now)
	if err != nil {
		return nil, huma.Error500InternalServerError("friend update failed")
	}
	return &noContent{Status: http.StatusNoContent}, nil
}

// deleteFriend removes one saved login without checking whether it still exists.
func (c galleryController) deleteFriend(ctx context.Context, input *friendPathInput) (*noContent, error) {
	owner, err := c.userID(input.GalleryAuth)
	if err != nil {
		return nil, err
	}
	if _, err := c.store.db.ExecContext(ctx, `DELETE FROM friends WHERE owner_id=? AND username=?`, owner, input.Username); err != nil {
		return nil, huma.Error500InternalServerError("friend removal failed")
	}
	return &noContent{Status: http.StatusNoContent}, nil
}
