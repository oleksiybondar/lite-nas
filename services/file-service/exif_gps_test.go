package main

import (
	"encoding/binary"
	"testing"
)

// jpegGPSFixture builds a complete bounded TIFF GPS directory inside JPEG APP1.
func jpegGPSFixture() []byte {
	tiff := make([]byte, 292)
	copy(tiff, "II")
	put16 := func(offset int, value uint16) { binary.LittleEndian.PutUint16(tiff[offset:], value) }
	put32 := func(offset int, value uint32) { binary.LittleEndian.PutUint32(tiff[offset:], value) }
	entry := func(offset int, tag, kind uint16, count, value uint32) {
		put16(offset, tag)
		put16(offset+2, kind)
		put32(offset+4, count)
		put32(offset+8, value)
	}
	put16(2, 42)
	put32(4, 8)
	put16(8, 1)
	entry(10, 0x8825, 4, 1, 120)
	put16(120, 4)
	entry(122, 1, 2, 2, 'N')
	entry(134, 2, 5, 3, 180)
	entry(146, 3, 2, 2, 'E')
	entry(158, 4, 5, 3, 204)
	for _, part := range []struct {
		offset int
		values [3]uint32
	}{{180, [3]uint32{55, 45, 0}}, {204, [3]uint32{12, 30, 0}}} {
		for index, value := range part.values {
			put32(part.offset+index*8, value)
			put32(part.offset+index*8+4, 1)
		}
	}
	segment := append([]byte("Exif\x00\x00"), tiff...)
	jpeg := []byte{0xff, 0xd8, 0xff, 0xe1, 0x01, 0x2c}
	jpeg = append(jpeg, segment...)
	return jpeg
}

// TestJPEGGPSLocationReadsCoordinates covers moments-backend/FR-008 GPS extraction.
func TestJPEGGPSLocationReadsCoordinates(t *testing.T) {
	location := jpegGPSLocation(jpegGPSFixture())
	if location == nil || location.Latitude != 55.75 || location.Longitude != 12.5 {
		t.Fatalf("GPS location = %+v", location)
	}
}

// TestJPEGGPSLocationRejectsMalformedReference avoids indexing a guessed hemisphere.
func TestJPEGGPSLocationRejectsMalformedReference(t *testing.T) {
	jpeg := jpegGPSFixture()
	jpeg[12+122+8] = 'X'
	if location := jpegGPSLocation(jpeg); location != nil {
		t.Fatalf("malformed GPS location = %+v", location)
	}
}
