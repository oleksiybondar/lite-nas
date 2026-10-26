package main

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// jpegEXIFTags returns selected textual fields actually embedded in JPEG EXIF.
func jpegEXIFTags(data []byte) []mediaTag {
	tags := []mediaTag{}
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return tags
	}
	for offset := 2; offset+4 <= len(data); {
		if data[offset] != 0xff {
			break
		}
		marker := data[offset+1]
		if marker == 0xda || marker == 0xd9 {
			break
		}
		length := int(binary.BigEndian.Uint16(data[offset+2:]))
		if length < 2 || offset+2+length > len(data) {
			break
		}
		segment := data[offset+4 : offset+2+length]
		if marker == 0xe1 && bytes.HasPrefix(segment, []byte("Exif\x00\x00")) {
			return tiffEXIFTags(segment[6:])
		}
		offset += 2 + length
	}
	return tags
}

// tiffEXIFTags safely follows the primary and EXIF IFD offsets in a JPEG payload.
func tiffEXIFTags(data []byte) []mediaTag {
	if len(data) < 8 {
		return []mediaTag{}
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return []mediaTag{}
	}
	if order.Uint16(data[2:4]) != 42 {
		return []mediaTag{}
	}
	tags := []mediaTag{}
	readIFD := func(offset uint32, includeSubIFD bool) uint32 {
		if offset > uint32(len(data)-2) {
			return 0
		}
		count := int(order.Uint16(data[offset:]))
		start := int(offset) + 2
		if count > 256 || start+count*12 > len(data) {
			return 0
		}
		var subIFD uint32
		for i := 0; i < count; i++ {
			entry := data[start+i*12 : start+(i+1)*12]
			field := order.Uint16(entry[:2])
			if field == 0x8769 && includeSubIFD {
				subIFD = order.Uint32(entry[8:12])
				continue
			}
			category := ""
			switch field {
			case 0x010e:
				category = "DESCRIPTION"
			case 0x010f:
				category = "CAMERA_MAKE"
			case 0x0110:
				category = "CAMERA_MODEL"
			case 0x9003:
				category = "TAKEN_AT"
			}
			if category == "" || order.Uint16(entry[2:4]) != 2 {
				continue
			}
			size := order.Uint32(entry[4:8])
			if size == 0 || size > 256 {
				continue
			}
			var raw []byte
			if size <= 4 {
				raw = entry[8 : 8+size]
			} else {
				valueOffset := order.Uint32(entry[8:12])
				if valueOffset > uint32(len(data)) || size > uint32(len(data))-valueOffset {
					continue
				}
				raw = data[valueOffset : valueOffset+size]
			}
			value := strings.TrimSpace(strings.TrimRight(string(raw), "\x00"))
			if value != "" && !strings.ContainsAny(value, "\x00\r\n") {
				tags = append(tags, mediaTag{Category: category, Value: value})
			}
		}
		return subIFD
	}
	subIFD := readIFD(order.Uint32(data[4:8]), true)
	if subIFD != 0 {
		readIFD(subIFD, false)
	}
	return tags
}
