package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// reverseGeocoder uses a persisted cache and a single low-rate public API client.
type reverseGeocoder struct {
	store    *galleryStore
	endpoint string
	client   *http.Client
	mu       sync.Mutex
	last     time.Time
}

// newReverseGeocoder keeps network failures bounded during a media scan.
func newReverseGeocoder(store *galleryStore, endpoint string) *reverseGeocoder {
	return &reverseGeocoder{store: store, endpoint: endpoint, client: &http.Client{
		Timeout:       3 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// onlinePlace returns a cached place or makes at most one public reverse lookup per 15 seconds.
// A rate-limited, failed, or missing lookup leaves the offline city label in place.
func (g *reverseGeocoder) onlinePlace(ctx context.Context, location sourceLocation) (cityPlace, bool) {
	if g == nil || g.endpoint == "" {
		return cityPlace{}, false
	}
	key := fmt.Sprintf("%.4f,%.4f", location.Latitude, location.Longitude)
	var cached cityPlace
	err := g.store.db.QueryRowContext(ctx, `SELECT place_name,search_key FROM reverse_geocode_cache WHERE coordinate_key=?`, key).Scan(&cached.name, &cached.search)
	if err == nil {
		return cached, true
	}
	if err != sql.ErrNoRows {
		return cityPlace{}, false
	}
	g.mu.Lock()
	if time.Since(g.last) < 15*time.Second {
		g.mu.Unlock()
		return cityPlace{}, false
	}
	g.last = time.Now()
	g.mu.Unlock()
	query := url.Values{
		"lat": {fmt.Sprintf("%.6f", location.Latitude)}, "lon": {fmt.Sprintf("%.6f", location.Longitude)},
		"format": {"jsonv2"}, "addressdetails": {"1"}, "namedetails": {"1"}, "accept-language": {"en"},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return cityPlace{}, false
	}
	request.Header.Set("User-Agent", "LiteNAS-Moments/0.1 (+https://github.com/oleksiybondar/lite-nas)")
	response, err := g.client.Do(request)
	if err != nil {
		return cityPlace{}, false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return cityPlace{}, false
	}
	var result struct {
		Address map[string]string `json:"address"`
		Names   map[string]string `json:"namedetails"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return cityPlace{}, false
	}
	city := firstPlace(result.Address, "city", "town", "village", "municipality", "hamlet")
	country := firstPlace(result.Address, "country")
	if city == "" || country == "" {
		return cityPlace{}, false
	}
	parts := []string{}
	if road := firstPlace(result.Address, "road", "pedestrian", "residential"); road != "" {
		parts = append(parts, road)
	}
	parts = append(parts, city, country)
	name := strings.Join(parts, ", ")
	searchTerms := []string{name}
	for _, value := range result.Names {
		if value != "" && len(value) < 256 {
			searchTerms = append(searchTerms, value)
		}
	}
	place := cityPlace{name: name, search: searchKey(strings.Join(searchTerms, " "))}
	_, _ = g.store.db.ExecContext(ctx, `INSERT OR IGNORE INTO reverse_geocode_cache(coordinate_key,place_name,search_key,created_at) VALUES(?,?,?,?)`,
		key, place.name, place.search, timestamp())
	return place, true
}

// firstPlace takes the most specific available OSM address field.
func firstPlace(address map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(address[key]); value != "" && len(value) <= 256 {
			return value
		}
	}
	return ""
}

// enrichLocation replaces a city-only tag when an online lookup is available.
func enrichLocation(ctx context.Context, item *scanMedia, info sourceMetadata, places *cityIndex) {
	if info.Location == nil || places == nil {
		return
	}
	place, ok := places.online.onlinePlace(ctx, *info.Location)
	if !ok {
		return
	}
	for index := range item.Tags {
		if item.Tags[index].Category == "LOCATION" {
			item.Tags[index] = scanTag{Category: "LOCATION", Value: place.name, Search: place.search}
			return
		}
	}
	item.Tags = append(item.Tags, scanTag{Category: "LOCATION", Value: place.name, Search: place.search})
}
