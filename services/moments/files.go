package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/shared/fileaccess"
)

// fileServiceClient calls only the configured loopback File Service origin.
type fileServiceClient struct {
	origin string
	client *http.Client
}

// sourceStat is the File Service-authorized source identity and kind.
type sourceStat struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

// sourceMetadata is File Service's authorized source-backed media response.
type sourceMetadata struct {
	Path        string `json:"path"`
	MediaType   string `json:"mediaType"`
	ContentType string `json:"contentType"`
	Width       *int   `json:"width"`
	Height      *int   `json:"height"`
	Tags        []struct {
		Category string `json:"category"`
		Value    string `json:"value"`
	} `json:"tags"`
}

// newFileServiceClient bounds internal metadata and discovery requests.
func newFileServiceClient(origin string) fileServiceClient {
	return fileServiceClient{origin: origin, client: &http.Client{
		Timeout:       15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// get performs a user-JWT-bound File Service read without token persistence.
func (c fileServiceClient) get(ctx context.Context, token, route, path, query string, output any) error {
	if token == "" || len(token) > 8192 || strings.ContainsAny(token, "\r\n") {
		return huma.Error401Unauthorized("invalid access token")
	}
	if fileaccess.ValidateHostPath(path) != nil {
		return huma.Error422UnprocessableEntity("invalid source path")
	}
	base, err := url.Parse(c.origin)
	if err != nil {
		return huma.Error500InternalServerError("File Service configuration invalid")
	}
	base.Path = route + path
	base.RawQuery = query
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return huma.Error500InternalServerError("File Service request failed")
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := c.client.Do(request)
	if err != nil {
		return huma.Error503ServiceUnavailable("File Service unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		switch response.StatusCode {
		case http.StatusUnauthorized:
			return huma.Error401Unauthorized("source token rejected")
		case http.StatusForbidden, http.StatusNotFound:
			return huma.Error422UnprocessableEntity("source path is inaccessible")
		default:
			return huma.Error503ServiceUnavailable("File Service unavailable")
		}
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 32<<20))
	if err := decoder.Decode(output); err != nil {
		return huma.Error503ServiceUnavailable("File Service returned invalid metadata")
	}
	return nil
}

// stat checks the path and expected source kind under the initiating user's ACL.
func (c fileServiceClient) stat(ctx context.Context, token, path, expectedKind string) (string, error) {
	var info sourceStat
	if err := c.get(ctx, token, "/stat", path, "", &info); err != nil {
		return "", err
	}
	if info.Kind != expectedKind || fileaccess.ValidateHostPath(info.Path) != nil {
		return "", huma.Error422UnprocessableEntity("source has the wrong kind")
	}
	return info.Path, nil
}

// listing retrieves the existing authorized recursive files-only array.
func (c fileServiceClient) listing(ctx context.Context, token, directory string) ([]string, error) {
	var entries []struct {
		Path string `json:"path"`
		Kind string `json:"kind"`
	}
	if err := c.get(ctx, token, "/ls", directory, "recursive&files-only", &entries); err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Kind != "file" || fileaccess.ValidateHostPath(entry.Path) != nil {
			return nil, errors.New("invalid File Service listing")
		}
		paths = append(paths, entry.Path)
	}
	return paths, nil
}

// metadata extracts only the source-backed fields File Service authorized.
func (c fileServiceClient) metadata(ctx context.Context, token, path string) (sourceMetadata, error) {
	var info sourceMetadata
	if err := c.get(ctx, token, "/metadata", path, "", &info); err != nil {
		return sourceMetadata{}, err
	}
	if info.Path == "" || fileaccess.ValidateHostPath(info.Path) != nil || info.Path != path {
		return sourceMetadata{}, errors.New("File Service returned a different source")
	}
	switch info.MediaType {
	case "IMAGE", "VIDEO", "AUDIO", "OTHER":
	default:
		return sourceMetadata{}, errors.New("File Service returned an invalid media type")
	}
	if info.Width != nil && (*info.Width < 1 || *info.Width > 100000) ||
		info.Height != nil && (*info.Height < 1 || *info.Height > 100000) {
		return sourceMetadata{}, errors.New("File Service returned invalid dimensions")
	}
	return info, nil
}
