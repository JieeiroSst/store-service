package nfcreader

import "fmt"

type TLV struct {
	Tag      uint32
	Value    []byte
	Children []TLV
}

func (t TLV) IsConstructed() bool {
	return t.Tag != 0 && (firstTagByte(t.Tag)&0x20) != 0
}

func firstTagByte(tag uint32) byte {
	for shift := 24; shift >= 0; shift -= 8 {
		b := byte(tag >> shift)
		if b != 0 || shift == 0 {
			return b
		}
	}
	return 0
}

func ParseTLV(data []byte) (TLV, int, error) {
	tag, tagLen, err := parseTag(data)
	if err != nil {
		return TLV{}, 0, err
	}
	length, lenLen, err := parseLength(data[tagLen:])
	if err != nil {
		return TLV{}, 0, err
	}
	valueStart := tagLen + lenLen
	valueEnd := valueStart + length
	if valueEnd > len(data) {
		return TLV{}, 0, fmt.Errorf("nfcreader: tag %#x declares length %d beyond available %d bytes", tag, length, len(data)-valueStart)
	}
	value := data[valueStart:valueEnd]

	node := TLV{Tag: tag, Value: value}
	if node.IsConstructed() {
		children, err := ParseTLVSequence(value)
		if err != nil {
			return TLV{}, 0, fmt.Errorf("nfcreader: tag %#x children: %w", tag, err)
		}
		node.Children = children
	}
	return node, valueEnd, nil
}

func ParseTLVSequence(data []byte) ([]TLV, error) {
	var nodes []TLV
	for len(data) > 0 {
		node, consumed, err := ParseTLV(data)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
		data = data[consumed:]
	}
	return nodes, nil
}

func (t TLV) Find(tag uint32) (TLV, bool) {
	for _, c := range t.Children {
		if c.Tag == tag {
			return c, true
		}
	}
	return TLV{}, false
}

func parseTag(data []byte) (tag uint32, consumed int, err error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("nfcreader: empty tag")
	}
	first := data[0]
	tag = uint32(first)
	consumed = 1
	if first&0x1F != 0x1F {
		return tag, consumed, nil
	}
	for {
		if consumed >= len(data) {
			return 0, 0, fmt.Errorf("nfcreader: truncated multi-byte tag")
		}
		b := data[consumed]
		tag = tag<<8 | uint32(b)
		consumed++
		if b&0x80 == 0 {
			break
		}
	}
	return tag, consumed, nil
}

func parseLength(data []byte) (length int, consumed int, err error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("nfcreader: empty length")
	}
	first := data[0]
	if first&0x80 == 0 {
		return int(first), 1, nil
	}
	numBytes := int(first & 0x7F)
	if numBytes == 0 {
		return 0, 0, fmt.Errorf("nfcreader: indefinite-form BER length is not supported for LDS files")
	}
	if numBytes > 4 || len(data) < 1+numBytes {
		return 0, 0, fmt.Errorf("nfcreader: unsupported or truncated long-form length")
	}
	length = 0
	for i := 0; i < numBytes; i++ {
		length = length<<8 | int(data[1+i])
	}
	return length, 1 + numBytes, nil
}
