package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mustGeocoderFixture serves one controlled reverse-geocoding response.
func mustGeocoderFixture(t *testing.T, response string, status int) (*reverseGeocoder, *int) {
	t.Helper()
	store, _ := mustSearchFixture(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls++
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return newReverseGeocoder(store, server.URL), &calls
}

// TestReverseGeocoderCachesStreetPlace verifies one lookup is reused for the same coordinates.
func TestReverseGeocoderCachesStreetPlace(t *testing.T) {
	geocoder, calls := mustGeocoderFixture(t, `{"address":{"road":"Calle Larios","city":"Málaga","country":"Spain"}}`, http.StatusOK)
	location := sourceLocation{Latitude: 36.72, Longitude: -4.42}
	geocoder.onlinePlace(context.Background(), location)
	geocoder.onlinePlace(context.Background(), location)
	if *calls != 1 {
		t.Fatalf("reverse lookup calls = %d, want 1", *calls)
	}
}

// TestReverseGeocoderIncludesStreet checks the human-readable place label.
func TestReverseGeocoderIncludesStreet(t *testing.T) {
	geocoder, _ := mustGeocoderFixture(t, `{"address":{"road":"Calle Larios","city":"Málaga","country":"Spain"}}`, http.StatusOK)
	place, _ := geocoder.onlinePlace(context.Background(), sourceLocation{Latitude: 36.72, Longitude: -4.42})
	if place.name != "Calle Larios, Málaga, Spain" {
		t.Fatalf("reverse place = %q", place.name)
	}
}

// TestReverseGeocoderFailureLeavesOfflineFallbackAvailable checks the soft online dependency.
func TestReverseGeocoderFailureLeavesOfflineFallbackAvailable(t *testing.T) {
	geocoder, _ := mustGeocoderFixture(t, `{}`, http.StatusServiceUnavailable)
	_, ok := geocoder.onlinePlace(context.Background(), sourceLocation{Latitude: 36.72, Longitude: -4.42})
	if ok {
		t.Fatal("failed online lookup returned a place")
	}
}

// TestEnrichLocation keeps online and offline place labels predictable during a scan.
func TestEnrichLocation(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		before []scanTag
		want   string
	}{
		{name: "replace offline city", status: http.StatusOK, before: []scanTag{{Category: "LOCATION", Value: "Málaga, Spain"}}, want: "Calle Larios, Málaga, Spain"},
		{name: "add missing location", status: http.StatusOK, want: "Calle Larios, Málaga, Spain"},
		{name: "retain offline fallback", status: http.StatusServiceUnavailable, before: []scanTag{{Category: "LOCATION", Value: "Málaga, Spain"}}, want: "Málaga, Spain"},
	} {
		t.Run(test.name, func(t *testing.T) {
			geocoder, _ := mustGeocoderFixture(t, `{"address":{"road":"Calle Larios","city":"Málaga","country":"Spain"}}`, test.status)
			item := scanMedia{Tags: test.before}
			location := sourceLocation{Latitude: 36.72, Longitude: -4.42}
			enrichLocation(context.Background(), &item, sourceMetadata{Location: &location}, &cityIndex{online: geocoder})
			if len(item.Tags) != 1 || item.Tags[0].Value != test.want {
				t.Fatalf("location tags = %+v, want %q", item.Tags, test.want)
			}
		})
	}
}
