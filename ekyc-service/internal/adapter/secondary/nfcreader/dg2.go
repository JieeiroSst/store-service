package nfcreader

import (
	"bytes"
	"fmt"
)

var (
	jpegSOI = []byte{0xFF, 0xD8, 0xFF}
	jp2Sig  = []byte{0x00, 0x00, 0x00, 0x0C, 0x6A, 0x50, 0x20, 0x20, 0x0D, 0x0A, 0x87, 0x0A}
)

func ParseDG2Photo(data []byte) ([]byte, error) {
	if img := findImageMarker(navigateToBiometricDataBlock(data)); img != nil {
		return img, nil
	}
	if img := findImageMarker(data); img != nil {
		return img, nil
	}
	return nil, fmt.Errorf("nfcreader: DG2: no embedded JPEG/JP2 image found")
}

func navigateToBiometricDataBlock(data []byte) []byte {
	root, _, err := ParseTLV(data)
	if err != nil || root.Tag != tagDG2 {
		return nil
	}
	group, ok := root.Find(tagBIOInfoGroup)
	if !ok {
		return nil
	}
	tmpl, ok := group.Find(tagBIOInfoTmpl)
	if !ok {
		return nil
	}
	block, ok := tmpl.Find(tagBioDataBlock)
	if !ok {
		return nil
	}
	return block.Value
}

func findImageMarker(data []byte) []byte {
	if data == nil {
		return nil
	}
	if i := bytes.Index(data, jpegSOI); i >= 0 {
		return data[i:]
	}
	if i := bytes.Index(data, jp2Sig); i >= 0 {
		return data[i:]
	}
	return nil
}
