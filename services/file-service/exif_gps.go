package main

import (
	"bytes"
	"encoding/binary"
	"math"
)

// gpsLocation is the validated WGS84 pair returned for a source JPEG.
type gpsLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// jpegGPSLocation follows the first EXIF segment to a complete GPS IFD.
func jpegGPSLocation(data []byte) *gpsLocation {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return nil
	}
	for offset := 2; offset+4 <= len(data); {
		if data[offset] != 0xff {
			return nil
		}
		marker := data[offset+1]
		if marker == 0xda || marker == 0xd9 {
			return nil
		}
		length := int(binary.BigEndian.Uint16(data[offset+2:]))
		if length < 2 || offset+2+length > len(data) {
			return nil
		}
		segment := data[offset+4 : offset+2+length]
		if marker == 0xe1 && bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
			return tiffGPSLocation(segment[6:])
		}
		offset += 2 + length
	}
	return nil
}

// gpsIFD returns checked 12-byte entries from one TIFF directory.
func gpsIFD(data []byte, order binary.ByteOrder, offset uint32) map[uint16][]byte {
	if uint64(offset)+2 > uint64(len(data)) {
		return nil
	}
	count := int(order.Uint16(data[offset:]))
	start := int(offset) + 2
	if count > 256 || start+count*12 > len(data) {
		return nil
	}
	entries := make(map[uint16][]byte, count)
	for i := 0; i < count; i++ {
		entry := data[start+i*12 : start+(i+1)*12]
		entries[order.Uint16(entry[:2])] = entry
	}
	return entries
}

// gpsRef accepts only TIFF ASCII direction bytes stored inline.
func gpsRef(entry []byte, first, second byte, order binary.ByteOrder) byte {
	if len(entry) != 12 || order.Uint16(entry[2:4]) != 2 || order.Uint32(entry[4:8]) < 2 {
		return 0
	}
	if entry[8] == first || entry[8] == second {
		return entry[8]
	}
	return 0
}

// gpsDegrees decodes three checked unsigned rationals into decimal degrees.
func gpsDegrees(data, entry []byte, order binary.ByteOrder, maximum float64) (float64, bool) {
	if len(entry) != 12 || order.Uint16(entry[2:4]) != 5 || order.Uint32(entry[4:8]) != 3 {
		return 0, false
	}
	offset := uint64(order.Uint32(entry[8:12]))
	if offset+24 > uint64(len(data)) {
		return 0, false
	}
	parts := [3]float64{}
	for i := range parts {
		start := int(offset) + i*8
		denominator := order.Uint32(data[start+4 : start+8])
		if denominator == 0 {
			return 0, false
		}
		parts[i] = float64(order.Uint32(data[start:start+4])) / float64(denominator)
	}
	value := parts[0] + parts[1]/60 + parts[2]/3600
	return value, parts[1] < 60 && parts[2] < 60 && value <= maximum && !math.IsNaN(value)
}

// tiffGPSLocation validates the primary and GPS directories before using coordinates.
func tiffGPSLocation(data []byte) *gpsLocation {
	if len(data) < 8 {
		return nil
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return nil
	}
	if order.Uint16(data[2:4]) != 42 {
		return nil
	}
	primary := gpsIFD(data, order, order.Uint32(data[4:8]))
	pointer := primary[0x8825]
	if len(pointer) != 12 || order.Uint16(pointer[2:4]) != 4 || order.Uint32(pointer[4:8]) != 1 {
		return nil
	}
	gps := gpsIFD(data, order, order.Uint32(pointer[8:12]))
	latRef := gpsRef(gps[1], 'N', 'S', order)
	lonRef := gpsRef(gps[3], 'E', 'W', order)
	latitude, latOK := gpsDegrees(data, gps[2], order, 90)
	longitude, lonOK := gpsDegrees(data, gps[4], order, 180)
	if latRef == 0 || lonRef == 0 || !latOK || !lonOK {
		return nil
	}
	if latRef == 'S' {
		latitude = -latitude
	}
	if lonRef == 'W' {
		longitude = -longitude
	}
	return &gpsLocation{Latitude: latitude, Longitude: longitude}
}
