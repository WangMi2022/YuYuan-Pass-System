package provider

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
)

const exifOrientationTag = 0x0112

func jpegEXIFOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 1
	}
	for offset := 2; offset+4 <= len(data); {
		if data[offset] != 0xff {
			offset++
			continue
		}
		for offset < len(data) && data[offset] == 0xff {
			offset++
		}
		if offset >= len(data) {
			break
		}
		marker := data[offset]
		offset++
		if marker == 0xd9 || marker == 0xda {
			break
		}
		if marker == 0x01 || marker >= 0xd0 && marker <= 0xd7 {
			continue
		}
		if offset+2 > len(data) {
			break
		}
		segmentSize := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if segmentSize < 2 || offset+segmentSize > len(data) {
			break
		}
		payload := data[offset+2 : offset+segmentSize]
		if marker == 0xe1 && bytes.HasPrefix(payload, []byte("Exif\x00\x00")) {
			return tiffOrientation(payload[6:])
		}
		offset += segmentSize
	}
	return 1
}

func tiffOrientation(data []byte) int {
	if len(data) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	if order.Uint16(data[2:4]) != 42 {
		return 1
	}
	ifdOffset := uint64(order.Uint32(data[4:8]))
	if ifdOffset+2 > uint64(len(data)) {
		return 1
	}
	entryCount := uint64(order.Uint16(data[ifdOffset : ifdOffset+2]))
	entriesStart := ifdOffset + 2
	if entryCount > (uint64(len(data))-entriesStart)/12 {
		return 1
	}
	for index := uint64(0); index < entryCount; index++ {
		offset := entriesStart + index*12
		entry := data[offset : offset+12]
		if order.Uint16(entry[:2]) != exifOrientationTag ||
			order.Uint16(entry[2:4]) != 3 || order.Uint32(entry[4:8]) != 1 {
			continue
		}
		orientation := int(order.Uint16(entry[8:10]))
		if orientation >= 1 && orientation <= 8 {
			return orientation
		}
		return 1
	}
	return 1
}

type orientedImage struct {
	source       image.Image
	sourceBounds image.Rectangle
	orientation  int
	bounds       image.Rectangle
}

func applyEXIFOrientation(source image.Image, orientation int) image.Image {
	if orientation < 2 || orientation > 8 {
		return source
	}
	sourceBounds := source.Bounds()
	width, height := sourceBounds.Dx(), sourceBounds.Dy()
	if orientation >= 5 {
		width, height = height, width
	}
	return &orientedImage{
		source: source, sourceBounds: sourceBounds, orientation: orientation,
		bounds: image.Rect(0, 0, width, height),
	}
}

func (img *orientedImage) ColorModel() color.Model {
	return img.source.ColorModel()
}

func (img *orientedImage) Bounds() image.Rectangle {
	return img.bounds
}

func (img *orientedImage) At(x, y int) color.Color {
	if !image.Pt(x, y).In(img.bounds) {
		return color.RGBA{}
	}
	width, height := img.sourceBounds.Dx(), img.sourceBounds.Dy()
	var sourceX, sourceY int
	switch img.orientation {
	case 2:
		sourceX, sourceY = width-1-x, y
	case 3:
		sourceX, sourceY = width-1-x, height-1-y
	case 4:
		sourceX, sourceY = x, height-1-y
	case 5:
		sourceX, sourceY = y, x
	case 6:
		sourceX, sourceY = y, height-1-x
	case 7:
		sourceX, sourceY = width-1-y, height-1-x
	case 8:
		sourceX, sourceY = width-1-y, x
	default:
		sourceX, sourceY = x, y
	}
	return img.source.At(img.sourceBounds.Min.X+sourceX, img.sourceBounds.Min.Y+sourceY)
}
