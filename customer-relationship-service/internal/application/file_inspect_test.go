package application

import (
	"bytes"
	"strings"
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/go-pdf/fpdf"
)

func plainPDF(t *testing.T, mutate func(*fpdf.Fpdf)) []byte {
	t.Helper()
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(40, 10, "Hop dong")
	if mutate != nil {
		mutate(pdf)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func pdfWithJS(t *testing.T) []byte {
	return plainPDF(t, func(p *fpdf.Fpdf) { p.SetJavascript("app.alert('hi');") })
}

func pdfWithAttachment(t *testing.T) []byte {
	return plainPDF(t, func(p *fpdf.Fpdf) {
		p.SetAttachments([]fpdf.Attachment{{Content: []byte("MZ payload"), Filename: "invoice.exe", Description: "x"}})
	})
}

func TestInspectPDF(t *testing.T) {
	clean := plainPDF(t, nil)
	js, attached := pdfWithJS(t), pdfWithAttachment(t)

	for name, c := range map[string]struct {
		data   []byte
		policy string
		reject string // substring of the reason; "" = accepted
	}{
		"a plain PDF":              {clean, "strict", ""},
		"JavaScript, strict":       {js, "strict", "JavaScript"},
		"JavaScript, lenient":      {js, "lenient", "JavaScript"},
		"JavaScript, checks off":   {js, "off", ""},
		"an embedded file":         {attached, "strict", "embedded file"},
		"garbage after the header": {[]byte("%PDF-1.7\nthis is not a pdf at all"), "strict", "cannot be inspected"},
		"the same, lenient":        {[]byte("%PDF-1.7\nthis is not a pdf at all"), "lenient", ""},
		"an empty-looking PDF":     {[]byte("%PDF-1.4\n%%EOF"), "strict", "cannot be inspected"},
	} {
		got := inspectPDF(c.data, c.policy)
		if (c.reject == "") != (got == "") || !strings.Contains(got, c.reject) {
			t.Errorf("%s: reason %q, want one containing %q", name, got, c.reject)
		}
	}
}

func TestWalkPDF_ParserFindsJavaScriptAndAttachments(t *testing.T) {
	if what, problem := walkPDF(pdfWithJS(t)); what == "" {
		t.Fatalf("the parser missed JavaScript (problem: %q)", problem)
	}
	if what, problem := walkPDF(pdfWithAttachment(t)); what == "" {
		t.Fatalf("the parser missed the embedded file (problem: %q)", problem)
	}
	if what, problem := walkPDF(plainPDF(t, nil)); what != "" || problem != "" {
		t.Fatalf("a plain PDF: %q, %q", what, problem)
	}
}

func TestInspectOffice(t *testing.T) {
	ok := zipOf(t, "[Content_Types].xml", "word/document.xml")
	macro := zipOf(t, "[Content_Types].xml", "word/document.xml", "word/vbaProject.bin")
	activex := zipOf(t, "[Content_Types].xml", "word/document.xml", "word/activeX/activeX1.xml")
	sheetMacro := zipOf(t, "[Content_Types].xml", "xl/workbook.xml", "xl/VBAPROJECT.BIN")

	if got := inspect(model.FormatDOCX, ok, "strict"); got != "" {
		t.Errorf("a clean docx: %q", got)
	}
	for name, data := range map[string][]byte{"macro": macro, "activex": activex, "xlsx macro": sheetMacro} {
		format := model.FormatDOCX
		if name == "xlsx macro" {
			format = model.FormatXLSX
		}
		if got := inspect(format, data, "strict"); got == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	// The PDF policy does not switch Office checks off.
	if got := inspect(model.FormatDOCX, macro, "off"); got == "" {
		t.Error("macros accepted when the PDF policy is off")
	}
	// Images and legacy formats have nothing to inspect here.
	if got := inspect(model.FormatPNG, pngBytes(), "strict"); got != "" {
		t.Errorf("png: %q", got)
	}
}
