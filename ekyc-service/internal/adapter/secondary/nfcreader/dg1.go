package nfcreader

import (
	"fmt"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/mrz"
)

func ParseDG1(data []byte) (mrz.TD1Fields, string, error) {
	root, _, err := ParseTLV(data)
	if err != nil {
		return mrz.TD1Fields{}, "", fmt.Errorf("nfcreader: parse DG1: %w", err)
	}
	if root.Tag != tagDG1 {
		return mrz.TD1Fields{}, "", fmt.Errorf("nfcreader: DG1: expected outer tag %#x, got %#x", tagDG1, root.Tag)
	}
	mrzField, ok := root.Find(tagMRZData)
	if !ok {
		return mrz.TD1Fields{}, "", fmt.Errorf("nfcreader: DG1: missing MRZ data element (tag %#x)", tagMRZData)
	}

	text := string(mrzField.Value)
	if len(text) != 90 {
		return mrz.TD1Fields{}, text, fmt.Errorf("nfcreader: DG1 MRZ length = %d, only the 90-character (3x30 TD1) layout is supported", len(text))
	}

	fields := mrz.ParseTD1(text[0:30], text[30:60], text[60:90])
	return fields, text, nil
}
