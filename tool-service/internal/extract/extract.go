package extract

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/dslipak/pdf"
)

const (
	MaxFileSize = 10 << 20
	MaxText     = 60000
)

func Text(filename string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("empty file")
	}
	if len(data) > MaxFileSize {
		return "", fmt.Errorf("file larger than %d MB", MaxFileSize>>20)
	}
	var (
		text string
		err  error
	)
	switch ext := strings.ToLower(filepath.Ext(filename)); ext {
	case ".docx":
		text, err = docx(data)
	case ".pdf":
		text, err = pdfText(data)
	case ".txt", ".md", ".markdown", ".yaml", ".yml", ".json":
		if !utf8.Valid(data) {
			return "", errors.New("text file is not valid UTF-8")
		}
		text = string(data)
	case ".doc":
		return "", errors.New("legacy .doc is not supported, please save as .docx")
	default:
		return "", fmt.Errorf("unsupported file type %q (use .pdf, .docx, .txt or .md)", ext)
	}
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("no readable text found (scanned PDFs need OCR first)")
	}
	if r := []rune(text); len(r) > MaxText {
		text = string(r[:MaxText])
	}
	return text, nil
}

func docx(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid .docx: %w", err)
	}
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		return docxXML(io.LimitReader(rc, 50<<20))
	}
	return "", errors.New("word/document.xml not found in .docx")
}

func docxXML(r io.Reader) (string, error) {
	var b strings.Builder
	dec := xml.NewDecoder(r)
	inText := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab":
				b.WriteByte('\t')
			case "br":
				b.WriteByte('\n')
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				b.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				b.Write(t)
			}
		}
	}
	return b.String(), nil
}

func pdfText(data []byte) (text string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot parse pdf: %v", r)
		}
	}()
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("cannot parse pdf: %w", err)
	}
	pr, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(io.LimitReader(pr, 20<<20))
	return string(out), err
}
