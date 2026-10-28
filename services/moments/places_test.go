package main

import (
	"path/filepath"
	"testing"
)

// TestPackagedCityLookup covers moments-backend/FR-008 for Spanish and Norwegian cities.
func TestPackagedCityLookup(t *testing.T) {
	index, err := loadCityIndex(filepath.Join("data", "cities1000.zip"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name     string
		lat, lon float64
		want     string
	}{
		{name: "Malaga", lat: 36.7213, lon: -4.4214, want: "Málaga, Spain"},
		{name: "Drammen", lat: 59.7441, lon: 10.2045, want: "Drammen, Norway"},
	} {
		t.Run(test.name, func(t *testing.T) {
			city, found := index.nearestCity(test.lat, test.lon)
			if !found || city.name != test.want {
				t.Fatalf("city = %q (found %v), want %q", city.name, found, test.want)
			}
		})
	}
}

// TestGPSMetadataProducesSearchableLocation covers moments-backend/FR-008 scan tagging.
func TestGPSMetadataProducesSearchableLocation(t *testing.T) {
	index, err := loadCityIndex(filepath.Join("data", "cities1000.zip"))
	if err != nil {
		t.Fatal(err)
	}
	info := sourceMetadata{Path: "/home/alice/malaga.jpg", MediaType: "IMAGE"}
	info.Location = &sourceLocation{Latitude: 36.7213, Longitude: -4.4214}
	media := scanMediaFromSource(info, index)
	location := ""
	for _, tag := range media.Tags {
		if tag.Category == "LOCATION" {
			location = tag.Value
		}
	}
	if location != "Málaga, Spain" {
		t.Fatalf("location tag = %q, want Málaga, Spain", location)
	}
}
