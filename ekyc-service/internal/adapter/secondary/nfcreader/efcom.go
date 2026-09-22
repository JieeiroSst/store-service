package nfcreader

import "fmt"

const (
	tagEFCOM        uint32 = 0x60
	tagLDSVersion   uint32 = 0x5F01
	tagUnicodeVer   uint32 = 0x5F36
	tagDGTagList    uint32 = 0x5C
	tagDG1          uint32 = 0x61
	tagDG2          uint32 = 0x75
	tagMRZData      uint32 = 0x5F1F
	tagBIOInfoGroup uint32 = 0x7F61
	tagBIOInfoTmpl  uint32 = 0x7F60
	tagBioDataBlock uint32 = 0x5F2E
)

type EFCOM struct {
	LDSVersion        string
	UnicodeVersion    string
	PresentDataGroups []int
}

func ParseEFCOM(data []byte) (*EFCOM, error) {
	root, _, err := ParseTLV(data)
	if err != nil {
		return nil, fmt.Errorf("nfcreader: parse EF.COM: %w", err)
	}
	if root.Tag != tagEFCOM {
		return nil, fmt.Errorf("nfcreader: EF.COM: expected outer tag %#x, got %#x", tagEFCOM, root.Tag)
	}

	com := &EFCOM{}
	if v, ok := root.Find(tagLDSVersion); ok {
		com.LDSVersion = string(v.Value)
	}
	if v, ok := root.Find(tagUnicodeVer); ok {
		com.UnicodeVersion = string(v.Value)
	}
	if v, ok := root.Find(tagDGTagList); ok {
		com.PresentDataGroups = decodeDGTagList(v.Value)
	}
	return com, nil
}

func decodeDGTagList(tags []byte) []int {
	byTag := map[byte]int{
		0x61: 1, 0x75: 2, 0x63: 3, 0x76: 4, 0x65: 5, 0x66: 6,
		0x67: 7, 0x68: 8, 0x69: 9, 0x6A: 10, 0x6B: 11, 0x6C: 12,
		0x6D: 13, 0x6E: 14, 0x6F: 15, 0x70: 16,
	}
	var groups []int
	for _, t := range tags {
		if n, ok := byTag[t]; ok {
			groups = append(groups, n)
		}
	}
	return groups
}
