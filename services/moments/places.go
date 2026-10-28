package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// cityPlace is one GeoNames populated place and its searchable alternate names.
type cityPlace struct {
	name, search        string
	latitude, longitude float64
}

// cityIndex groups city coordinates into degree cells for bounded offline lookup.
type cityIndex struct {
	cells           map[[2]int][]cityPlace
	online          *reverseGeocoder
	captionScript   string
	captionMu       sync.Mutex
	embeddingScript string
}

// loadCityIndex reads the packaged GeoNames extract once at service startup.
func loadCityIndex(path string) (*cityIndex, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()
	if len(archive.File) != 1 || archive.File[0].Name != "cities1000.txt" {
		return nil, errors.New("GeoNames archive must contain cities1000.txt")
	}
	reader, err := archive.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return parseCities(reader)
}

// parseCities validates GeoNames fields before adding places to the lookup grid.
func parseCities(reader io.Reader) (*cityIndex, error) {
	index := &cityIndex{cells: make(map[[2]int][]cityPlace)}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	count := 0
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 15 {
			return nil, fmt.Errorf("invalid GeoNames city record at line %d", count+1)
		}
		latitude, latErr := strconv.ParseFloat(fields[4], 64)
		longitude, lonErr := strconv.ParseFloat(fields[5], 64)
		if latErr != nil || lonErr != nil || math.IsNaN(latitude) || math.IsNaN(longitude) ||
			latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 || fields[1] == "" {
			return nil, fmt.Errorf("invalid GeoNames city coordinates at line %d", count+1)
		}
		country := fields[8]
		if region, err := language.ParseRegion(country); err == nil {
			country = display.English.Regions().Name(region)
		}
		name := fields[1] + ", " + country
		search := searchKey(name + " " + fields[2] + " " + fields[3] + " " + fields[8])
		place := cityPlace{name: name, search: search, latitude: latitude, longitude: longitude}
		cell := [2]int{int(math.Floor(latitude)), int(math.Floor(longitude))}
		index.cells[cell] = append(index.cells[cell], place)
		count++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("GeoNames city extract is empty")
	}
	return index, nil
}

// nearestCity returns a named place within 50 km, avoiding labels far from GPS.
func (index *cityIndex) nearestCity(latitude, longitude float64) (cityPlace, bool) {
	if index == nil || math.IsNaN(latitude) || math.IsNaN(longitude) || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return cityPlace{}, false
	}
	latRadius := 1
	lonRadius := int(math.Ceil(50/(111.2*math.Max(0.01, math.Cos(latitude*math.Pi/180))))) + 1
	bestDistance := 50.0
	var best cityPlace
	found := false
	for lat := int(math.Floor(latitude)) - latRadius; lat <= int(math.Floor(latitude))+latRadius; lat++ {
		for lon := int(math.Floor(longitude)) - lonRadius; lon <= int(math.Floor(longitude))+lonRadius; lon++ {
			wrapped := lon
			if wrapped < -180 {
				wrapped += 360
			} else if wrapped > 180 {
				wrapped -= 360
			}
			for _, city := range index.cells[[2]int{lat, wrapped}] {
				distance := cityDistance(latitude, longitude, city.latitude, city.longitude)
				if distance < bestDistance {
					best, bestDistance, found = city, distance, true
				}
			}
		}
	}
	return best, found
}

// cityDistance computes great-circle distance in kilometres for WGS84 points.
func cityDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0
	latDelta := (lat2 - lat1) * math.Pi / 180
	lonDelta := (lon2 - lon1) * math.Pi / 180
	a := math.Pow(math.Sin(latDelta/2), 2) + math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Pow(math.Sin(lonDelta/2), 2)
	return 2 * earthRadius * math.Asin(math.Min(1, math.Sqrt(a)))
}
