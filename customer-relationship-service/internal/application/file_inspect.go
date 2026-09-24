package application

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/digitorus/pdf"
)

const (
	policyStrict  = "strict"
	policyLenient = "lenient"
	policyOff     = "off"

	// maxZipBytes bounds what an Office file may claim to expand to.
	maxZipBytes = 512 << 20
	// maxWalkDepth bounds how deep inspectValue follows nested objects.
	maxWalkDepth = 4
)

// inspect looks for active content in a file whose format is known. It returns
// a reason when the file must be refused, or an error-free empty string.
func inspect(format string, data []byte, pdfPolicy string) string {
	switch format {
	case model.FormatPDF:
		return inspectPDF(data, pdfPolicy)
	case model.FormatDOCX, model.FormatXLSX:
		return inspectOffice(data)
	}
	return ""
}

// Names that make a PDF actively do something: run script, launch programs,
// talk to the network, play media, or carry other files.
var dangerousNames = map[string]string{
	"JavaScript": "JavaScript", "Launch": "a launch action", "SubmitForm": "a form submission",
	"ImportData": "a data import action", "GoToR": "a link to another file", "GoToE": "a link to an embedded file",
	"Rendition": "media playback", "RichMedia": "rich media", "3D": "3D content", "Movie": "a movie",
	"Sound": "sound", "Screen": "a screen annotation", "EmbeddedFile": "an embedded file", "XFA": "XFA forms",
}

// Dictionary keys with the same meaning.
var dangerousKeys = map[string]string{
	"JS": "JavaScript", "JavaScript": "JavaScript",
	"XFA": "XFA forms", "RichMediaContent": "rich media",
}

// Raw byte patterns, for the parts of the file the parser cannot see.
var rawPatterns = map[string]string{
	"/JavaScript": "JavaScript", "/JS(": "JavaScript", "/JS ": "JavaScript", "/JS<": "JavaScript",
	"/Launch": "a launch action", "/RichMedia": "rich media",
	"/SubmitForm": "a form submission", "/ImportData": "a data import action",
}

func inspectPDF(data []byte, policy string) string {
	if policy == policyOff {
		return ""
	}
	if found := scanRaw(data); found != "" {
		return "the PDF contains " + found
	}

	found, err := walkPDF(data)
	switch {
	case found != "":
		return "the PDF contains " + found
	case err != "" && policy != policyLenient:
		return "the PDF cannot be inspected for active content (" + err + ")"
	}
	return ""
}

func scanRaw(data []byte) string {
	for pattern, what := range rawPatterns {
		if bytes.Contains(data, []byte(pattern)) {
			return what
		}
	}
	// "/EmbeddedFile" is an embedded file; "/EmbeddedFiles" is only the name
	// tree, which some generators write empty into every document.
	const embedded = "/EmbeddedFile"
	for rest := data; ; {
		i := bytes.Index(rest, []byte(embedded))
		if i < 0 {
			break
		}
		rest = rest[i+len(embedded):]
		if len(rest) == 0 || rest[0] != 's' {
			return "an embedded file"
		}
	}
	return ""
}

// walkPDF parses the file and looks at every object. It returns what it found,
// or why it could not look.
func walkPDF(data []byte) (found string, problem string) {
	defer func() {
		// The PDF parser panics on malformed input.
		if r := recover(); r != nil {
			found, problem = "", "malformed file"
		}
	}()

	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err.Error()
	}
	if !r.Trailer().Key("Encrypt").IsNull() {
		return "", "encrypted"
	}
	for _, x := range r.Xref() {
		if what := inspectValue(r.Resolve(x.Ptr(), x.Ptr()), 0); what != "" {
			return what, ""
		}
	}
	// Objects inside compressed object streams are reached through the trailer.
	if what := inspectValue(r.Trailer(), 0); what != "" {
		return what, ""
	}
	return "", ""
}

func inspectValue(v pdf.Value, depth int) string {
	if depth > maxWalkDepth {
		return ""
	}
	switch v.Kind() {
	case pdf.Name:
		if what, bad := dangerousNames[v.Name()]; bad {
			return what
		}
	case pdf.Array:
		for i := 0; i < v.Len(); i++ {
			if what := inspectValue(v.Index(i), depth+1); what != "" {
				return what
			}
		}
	case pdf.Dict, pdf.Stream:
		for _, k := range v.Keys() {
			if what, bad := dangerousKeys[k]; bad {
				return what
			}
			if k == "EmbeddedFiles" {
				// Only a tree with entries carries files.
				tree := v.Key(k)
				if tree.Key("Names").Len() > 0 || tree.Key("Kids").Len() > 0 {
					return "embedded files"
				}
				continue
			}
			if what := inspectValue(v.Key(k), depth+1); what != "" {
				return what
			}
		}
	}
	return ""
}

// inspectOffice refuses Office Open XML files that carry macros or ActiveX
// controls, and zip archives that expand absurdly.
func inspectOffice(data []byte) string {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "the archive cannot be read"
	}
	var total uint64
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		switch {
		case strings.HasSuffix(name, "vbaproject.bin"):
			return "the document contains macros"
		case strings.Contains(name, "/activex/") || strings.HasPrefix(name, "activex/"):
			return "the document contains ActiveX controls"
		}
		total += f.UncompressedSize64
	}
	if total > maxZipBytes {
		return fmt.Sprintf("the document expands to more than %d MB", maxZipBytes>>20)
	}
	return ""
}
