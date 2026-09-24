package application

import (
	"archive/zip"
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"path"
	"strings"
	"unicode"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
)

const (
	maxNameRunes   = 200
	maxZipEntries  = 20000
	pdfHeaderRange = 1024
)

func detectFormat(data []byte) (string, bool) {
	switch {
	case bytes.Contains(data[:min(len(data), pdfHeaderRange)], []byte("%PDF-")):
		return model.FormatPDF, true
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return model.FormatJPEG, imageHeaderOK(data, "jpeg")
	case bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return model.FormatPNG, imageHeaderOK(data, "png")
	case bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}):
		return "ole", true
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return officeFormat(data)
	}
	return "", false
}

func imageHeaderOK(data []byte, want string) bool {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	return err == nil && format == want && cfg.Width > 0 && cfg.Height > 0
}

func officeFormat(data []byte) (string, bool) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(zr.File) > maxZipEntries {
		return "", false
	}
	has := map[string]bool{}
	for _, f := range zr.File {
		has[f.Name] = true
	}
	switch {
	case has["[Content_Types].xml"] && has["word/document.xml"]:
		return model.FormatDOCX, true
	case has["[Content_Types].xml"] && has["xl/workbook.xml"]:
		return model.FormatXLSX, true
	}
	return "", false
}

func matches(detected string, ext model.FileFormat) bool {
	if detected == "ole" {
		return ext.Name == model.FormatDOC || ext.Name == model.FormatXLS
	}
	return detected == ext.Name
}

func cleanName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	if name == "." || name == "/" {
		return ""
	}
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '"' || r == '<' || r == '>' || r == '|' || r == ':' || r == '*' || r == '?' {
			return -1
		}
		return r
	}, name)
	name = strings.Trim(name, " .")

	if r := []rune(name); len(r) > maxNameRunes {
		ext := path.Ext(name)
		keep := maxNameRunes - len([]rune(ext))
		if keep < 1 {
			keep = 1
		}
		name = string(r[:keep]) + ext
	}
	return name
}
