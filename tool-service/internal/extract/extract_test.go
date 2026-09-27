package extract

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func makeDocx(t *testing.T, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("word/document.xml")
	w.Write([]byte(`<?xml version="1.0"?><w:document xmlns:w="x"><w:body>` + body + `</w:body></w:document>`))
	zw.Close()
	return buf.Bytes()
}

func TestDocx(t *testing.T) {
	data := makeDocx(t, `<w:p><w:r><w:t>POST /orders</w:t></w:r></w:p><w:p><w:r><w:t>amount must be</w:t></w:r><w:r><w:t> &gt; 0</w:t></w:r></w:p>`)
	got, err := Text("task.DOCX", data)
	if err != nil {
		t.Fatal(err)
	}
	if got != "POST /orders\namount must be > 0" {
		t.Fatalf("got %q", got)
	}
}

func TestErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		file string
		data []byte
		want string
	}{
		"empty":       {"a.pdf", nil, "empty"},
		"unsupported": {"a.exe", []byte("x"), "unsupported"},
		"legacy doc":  {"a.doc", []byte("x"), ".docx"},
		"bad docx":    {"a.docx", []byte("not a zip"), "not a valid"},
		"bad pdf":     {"a.pdf", []byte("not a pdf"), "pdf"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Text(tc.file, tc.data)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want contains %q", err, tc.want)
			}
		})
	}
}

func TestPlainText(t *testing.T) {
	if got, err := Text("a.md", []byte("  hello  ")); err != nil || got != "hello" {
		t.Fatalf("got %q, %v", got, err)
	}
}
