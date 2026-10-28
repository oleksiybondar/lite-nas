package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
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

// sourceLocation is the validated GPS pair supplied by File Service.
type sourceLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// sourceMetadata is File Service's authorized source-backed media response.
type sourceMetadata struct {
	Path        string          `json:"path"`
	MediaType   string          `json:"mediaType"`
	ContentType string          `json:"contentType"`
	Width       *int            `json:"width"`
	Height      *int            `json:"height"`
	Location    *sourceLocation `json:"location"`
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

// request performs a validated, user-JWT-bound File Service read without token persistence.
func (c fileServiceClient) request(ctx context.Context, token, route, path, query string) (*http.Response, error) {
	return c.requestWithScheme(ctx, "Bearer", token, route, path, query)
}

// captionRequest sends a job grant only to the exact File Service source route.
func (c fileServiceClient) captionRequest(ctx context.Context, grant, path string) (*http.Response, error) {
	return c.requestWithScheme(ctx, "CaptionGrant", grant, "", path, "")
}

// requestWithScheme keeps source path validation shared by user and job reads.
func (c fileServiceClient) requestWithScheme(ctx context.Context, scheme, token, route, path, query string) (*http.Response, error) {
	if token == "" || len(token) > 8192 || strings.ContainsAny(token, "\r\n") {
		return nil, huma.Error401Unauthorized("invalid access token")
	}
	if fileaccess.ValidateHostPath(path) != nil {
		return nil, huma.Error422UnprocessableEntity("invalid source path")
	}
	base, err := url.Parse(c.origin)
	if err != nil {
		return nil, huma.Error500InternalServerError("File Service configuration invalid")
	}
	base.Path = route + path
	base.RawQuery = query
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("File Service request failed")
	}
	request.Header.Set("Authorization", scheme+" "+token)
	response, err := c.client.Do(request)
	if err != nil {
		return nil, huma.Error503ServiceUnavailable("File Service unavailable")
	}
	return response, nil
}

// get decodes one successful File Service response with existing status mapping.
func (c fileServiceClient) get(ctx context.Context, token, route, path, query string, output any) error {
	response, err := c.request(ctx, token, route, path, query)
	if err != nil {
		return err
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

// missing confirms an absent file only after its parent directory is readable.
func (c fileServiceClient) missing(ctx context.Context, token, path string) (bool, error) {
	if _, err := c.stat(ctx, token, filepath.Dir(path), "directory"); err != nil {
		return false, err
	}
	response, err := c.request(ctx, token, "/stat", path, "")
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusNotFound:
		return true, nil
	case http.StatusForbidden:
		return false, nil
	case http.StatusOK:
		var info sourceStat
		if err := json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&info); err != nil || info.Path != path || (info.Kind != "file" && info.Kind != "directory") {
			return false, huma.Error503ServiceUnavailable("File Service returned invalid source identity")
		}
		return info.Kind != "file", nil
	default:
		return false, huma.Error503ServiceUnavailable("could not confirm source deletion")
	}
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
	if info.Location != nil && (info.Location.Latitude < -90 || info.Location.Latitude > 90 ||
		info.Location.Longitude < -180 || info.Location.Longitude > 180) {
		return sourceMetadata{}, errors.New("File Service returned invalid GPS coordinates")
	}
	return info, nil
}
