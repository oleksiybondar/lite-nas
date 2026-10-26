package main

import (
	"encoding/binary"
	"testing"
)

// TestJPEGEXIFTagsReadsDescription covers moments-backend/FR-003 source-backed tags.
func TestJPEGEXIFTagsReadsDescription(t *testing.T) {
	tiff := make([]byte, 32)
	copy(tiff, "II")
	binary.LittleEndian.PutUint16(tiff[2:4], 42)
	binary.LittleEndian.PutUint32(tiff[4:8], 8)
	binary.LittleEndian.PutUint16(tiff[8:10], 1)
	binary.LittleEndian.PutUint16(tiff[10:12], 0x010e)
	binary.LittleEndian.PutUint16(tiff[12:14], 2)
	binary.LittleEndian.PutUint32(tiff[14:18], 6)
	binary.LittleEndian.PutUint32(tiff[18:22], 26)
	copy(tiff[26:], "Hello\x00")
	segment := append([]byte("Exif\x00\x00"), tiff...)
	jpeg := []byte{0xff, 0xd8, 0xff, 0xe1, 0, byte(len(segment) + 2)}
	jpeg = append(jpeg, segment...)
	jpeg = append(jpeg, 0xff, 0xd9)
	tags := jpegEXIFTags(jpeg)
	if len(tags) != 1 || tags[0] != (mediaTag{Category: "DESCRIPTION", Value: "Hello"}) {
		t.Fatalf("EXIF tags = %#v", tags)
	}
}
