package application

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
	"github.com/go-pdf/fpdf"
	"image"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

func zipOf(t *testing.T, names ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range names {
		w, _ := zw.Create(n)
		w.Write([]byte("<x/>"))
	}
	zw.Close()
	return buf.Bytes()
}

var (
	pdfBytes  = validPDF()
	oleBytes  = append([]byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, make([]byte, 64)...)
	testImage = image.NewRGBA(image.Rect(0, 0, 4, 4))
)

func pngBytes() []byte  { var b bytes.Buffer; _ = png.Encode(&b, testImage); return b.Bytes() }
func jpegBytes() []byte { var b bytes.Buffer; _ = jpeg.Encode(&b, testImage, nil); return b.Bytes() }

func TestDetectFormat(t *testing.T) {
	docx := zipOf(t, "[Content_Types].xml", "word/document.xml")
	xlsx := zipOf(t, "[Content_Types].xml", "xl/workbook.xml")

	for name, c := range map[string]struct {
		data []byte
		want string
		ok   bool
	}{
		"pdf":                     {pdfBytes, model.FormatPDF, true},
		"pdf with a leading byte": {append([]byte("\n"), pdfBytes...), model.FormatPDF, true},
		"png":                     {pngBytes(), model.FormatPNG, true},
		"jpeg":                    {jpegBytes(), model.FormatJPEG, true},
		"docx":                    {docx, model.FormatDOCX, true},
		"xlsx":                    {xlsx, model.FormatXLSX, true},
		"legacy office":           {oleBytes, "ole", true},
		"a zip that is neither":   {zipOf(t, "a.txt"), "", false},
		"a zip with no manifest":  {zipOf(t, "word/document.xml"), "", false},
		"a truncated png":         {pngBytes()[:20], "", false},
		"a truncated jpeg":        {jpegBytes()[:10], "", false},
		"text":                    {[]byte("hello"), "", false},
		"an executable":           {[]byte("MZ\x90\x00\x03"), "", false},
		"empty":                   {nil, "", false},
	} {
		got, ok := detectFormat(c.data)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("%s: got %q, %v; want %q, %v", name, got, ok, c.want, c.ok)
		}
	}
}

func TestCleanName(t *testing.T) {
	for in, want := range map[string]string{
		"hop-dong.pdf":                    "hop-dong.pdf",
		"Hợp đồng số 01/2026.pdf":         "2026.pdf", // a path: only the last element survives
		"Hợp đồng số 01-2026.pdf":         "Hợp đồng số 01-2026.pdf",
		`C:\Users\a\scan.jpg`:             "scan.jpg",
		"../../etc/passwd":                "passwd",
		"a\x00b\r\n.pdf":                  "ab.pdf",
		`na"me<>|:*?.pdf`:                 "name.pdf",
		"  .hidden.pdf ":                  "hidden.pdf",
		"":                                "",
		"..":                              "",
		"/":                               "",
		strings.Repeat("ạ", 300) + ".pdf": strings.Repeat("ạ", 196) + ".pdf",
	} {
		if got := cleanName(in); got != want {
			t.Errorf("cleanName(%q) = %q, want %q", in, got, want)
		}
	}
}

type fakeScanner struct {
	on    bool
	res   port.ScanResult
	err   error
	calls int
}

func (f *fakeScanner) Enabled() bool { return f.on }
func (f *fakeScanner) Scan(context.Context, string, []byte) (port.ScanResult, error) {
	f.calls++
	return f.res, f.err
}

type filesFixture struct {
	env       *fileEnv
	svc       *ContractFiles
	contracts *memRepo[model.Contract]
	store     *testsupport.MemStore
	scanner   *fakeScanner
}

func newFilesFixture(t *testing.T, mutate ...func(*config.FilesConfig)) *filesFixture {
	t.Helper()
	f := &filesFixture{
		env: newFileEnv(t), contracts: newMemRepo[model.Contract](), store: testsupport.NewMemStore(),
		scanner: &fakeScanner{res: port.ScanResult{Clean: true, Engine: "clamav"}},
	}
	f.contracts.items[1] = &model.Contract{Base: model.Base{ID: 1}}
	f.contracts.items[2] = &model.Contract{Base: model.Base{ID: 2}}
	f.contracts.next = 2
	cfg := config.FilesConfig{MaxBytes: 1 << 20, PDFPolicy: "strict", ScanRequired: true}
	for _, m := range mutate {
		m(&cfg)
	}
	f.svc = f.env.buildWith(f.contracts, f.store, repository.NewTxRunner(f.env.db), f.scanner, cfg)
	return f
}

func up(kind, name string, data []byte) port.FileUpload {
	return port.FileUpload{ContractID: 1, Kind: kind, Name: name, Description: "d", Data: data}
}

func (f *filesFixture) eventsOf(t *testing.T, action string) []model.ContractFileEvent {
	t.Helper()
	rows, _, err := f.env.events.List(context.Background(), port.ListQuery{Limit: 1000, Equals: map[string]any{"action": action}})
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func (f *filesFixture) mustUpload(t *testing.T, p auth.Principal, in port.FileUpload) *model.ContractFile {
	t.Helper()
	got, err := f.svc.Upload(as(p), in)
	if err != nil {
		t.Fatalf("upload %s: %v", in.Name, err)
	}
	return got
}

func TestUpload_AcceptsEveryKindItsFormats(t *testing.T) {
	docx := zipOf(t, "[Content_Types].xml", "word/document.xml")
	xlsx := zipOf(t, "[Content_Types].xml", "xl/workbook.xml")
	f := newFilesFixture(t)
	f.scanner.on = true

	// A staff member may not upload "legal" files; a manager does those.
	for _, c := range []struct {
		who        auth.Principal
		kind, name string
		data       []byte
		ct         string
	}{
		{staff, model.FileKindContract, "hd.pdf", pdfBytes, "application/pdf"},
		{staff, model.FileKindContract, "hd.docx", docx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{staff, model.FileKindContract, "hd.doc", oleBytes, "application/msword"},
		{staff, model.FileKindAppendix, "pl.PDF", pdfBytes, "application/pdf"},
		{staff, model.FileKindScan, "scan.jpg", jpegBytes(), "image/jpeg"},
		{staff, model.FileKindScan, "scan.jpeg", jpegBytes(), "image/jpeg"},
		{manager, model.FileKindLegal, "dkkd.png", pngBytes(), "image/png"},
		{staff, model.FileKindPayment, "ck.xlsx", xlsx, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{staff, model.FileKindPayment, "ck.xls", oleBytes, "application/vnd.ms-excel"},
		{staff, model.FileKindAcceptance, "nt.pdf", pdfBytes, "application/pdf"},
		{staff, model.FileKindAttachment, "a.xlsx", xlsx, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{staff, model.FileKindOther, "o.png", pngBytes(), "image/png"},
	} {
		got, err := f.svc.Upload(as(c.who), up(c.kind, c.name, c.data))
		if err != nil {
			t.Errorf("%s %s: %v", c.kind, c.name, err)
			continue
		}
		if got.ContentType != c.ct || got.Kind != c.kind || got.Size != int64(len(c.data)) || len(got.SHA256) != 64 ||
			got.Source != model.FileSourceUpload || got.UploadedBy != c.who.Name || got.UploaderID != c.who.Subject ||
			got.Version != 1 || !got.Latest || got.ScanStatus != model.ScanClean || got.ScanEngine != "clamav" || got.ScannedAt == nil {
			t.Errorf("%s %s: row = %+v", c.kind, c.name, got)
		}
	}
}

func TestUpload_Rejects(t *testing.T) {
	docx := zipOf(t, "[Content_Types].xml", "word/document.xml")
	f := newFilesFixture(t, func(c *config.FilesConfig) { c.MaxBytes = 4 << 10 })

	rejected := 0
	for name, c := range map[string]struct {
		in   port.FileUpload
		want error
	}{
		"unknown kind":                 {up("nonsense", "a.pdf", pdfBytes), common.ErrInvalidRequest},
		"empty file":                   {up(model.FileKindContract, "a.pdf", nil), common.ErrInvalidRequest},
		"no name":                      {up(model.FileKindContract, "", pdfBytes), common.ErrInvalidRequest},
		"a system kind":                {up(model.FileKindRenewal, "a.pdf", pdfBytes), common.ErrForbidden},
		"too large":                    {up(model.FileKindContract, "a.pdf", append(append([]byte{}, pdfBytes...), make([]byte, 8<<10)...)), common.ErrTooLarge},
		"extension not accepted":       {up(model.FileKindAttachment, "run.exe", pdfBytes), common.ErrUnsupportedMedia},
		"no extension":                 {up(model.FileKindAttachment, "readme", pdfBytes), common.ErrUnsupportedMedia},
		"executable renamed .pdf":      {up(model.FileKindContract, "a.pdf", []byte("MZ\x90\x00 not a pdf")), common.ErrUnsupportedMedia},
		"pdf renamed .png":             {up(model.FileKindAttachment, "a.png", pdfBytes), common.ErrUnsupportedMedia},
		"docx renamed .xlsx":           {up(model.FileKindAttachment, "a.xlsx", docx), common.ErrUnsupportedMedia},
		"png in a documents-only kind": {up(model.FileKindContract, "a.png", pngBytes()), common.ErrUnsupportedMedia},
		"spreadsheet as a scan":        {up(model.FileKindScan, "a.xls", oleBytes), common.ErrUnsupportedMedia},
	} {
		if _, err := f.svc.Upload(as(staff), c.in); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", name, err, c.want)
		}
		if c.in.Kind != "nonsense" { // an unknown kind is refused before anything is recorded
			rejected++
		}
	}
	if f.env.docs.count() != 0 || len(f.store.Objects) != 0 {
		t.Fatalf("rejected uploads left %d rows and %d objects", f.env.docs.count(), len(f.store.Objects))
	}
	// Each refusal is on the audit trail.
	if got := len(f.eventsOf(t, model.FileEventRejected)); got != rejected {
		t.Fatalf("rejected events = %d, want %d", got, rejected)
	}

	in := up(model.FileKindContract, "a.pdf", pdfBytes)
	in.ContractID = 99
	if _, err := f.svc.Upload(as(staff), in); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("unknown contract: %v, want ErrNotFound", err)
	}
}

func TestUpload_StorageBackends(t *testing.T) {
	f := newFilesFixture(t)
	a := f.mustUpload(t, staff, up(model.FileKindContract, "Hợp đồng.pdf", pdfBytes))
	b := f.mustUpload(t, staff, up(model.FileKindContract, "Hợp đồng.pdf", pdfBytes)) // same bytes again
	if a.Data != nil || !strings.HasPrefix(a.ObjectKey, "contracts/1/contract/") || !strings.HasSuffix(a.ObjectKey, ".pdf") {
		t.Fatalf("row = %+v", a)
	}
	if a.ObjectKey == b.ObjectKey || len(f.store.Objects) != 2 {
		t.Fatal("two uploads of the same bytes must not share an object")
	}
	if strings.Contains(a.ObjectKey, "Hợp") {
		t.Fatal("the client's file name leaked into the object key")
	}

	f = newFilesFixture(t)
	f.store.Off = true
	row := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))
	if row.ObjectKey != "" || !bytes.Equal(row.Data, pdfBytes) || len(f.store.Objects) != 0 {
		t.Fatalf("row = %+v", row)
	}
}

func TestPrepare_TheSizeLimitAppliesToUploadsNotToSystemFiles(t *testing.T) {
	f := newFilesFixture(t, func(c *config.FilesConfig) { c.MaxBytes = 16 })
	f.scanner.on = true

	if _, err := f.svc.Upload(as(staff), up(model.FileKindContract, "a.pdf", pdfBytes)); !errors.Is(err, common.ErrTooLarge) {
		t.Fatalf("upload: %v, want ErrTooLarge", err)
	}
	got, err := f.svc.Prepare(context.Background(), port.FileSpec{
		ContractID: 1, Kind: model.FileKindRenewal, Name: "Phu-luc.pdf", Source: model.FileSourceSystem, Data: pdfBytes,
	})
	if err != nil || got.Source != model.FileSourceSystem || got.ScanStatus != model.ScanSystem || got.UploaderID != "system" {
		t.Fatalf("system file: %+v, %v", got, err)
	}
	if f.scanner.calls != 0 {
		t.Fatal("a file the service produced was sent to the virus scanner")
	}
}

// ---- who may do what -------------------------------------------------------

func TestAuthorization(t *testing.T) {
	f := newFilesFixture(t)
	anon := context.Background()
	staffFile := f.mustUpload(t, staff, up(model.FileKindContract, "staff.pdf", pdfBytes))
	legal := f.mustUpload(t, manager, up(model.FileKindLegal, "dkkd.pdf", pdfBytes))

	t.Run("no principal", func(t *testing.T) {
		if _, err := f.svc.Upload(anon, up(model.FileKindContract, "a.pdf", pdfBytes)); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("upload: %v", err)
		}
		if _, _, err := f.svc.List(anon, 1, port.FileListFilter{Limit: 10}); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("list: %v", err)
		}
		if _, _, err := f.svc.Download(anon, 1, staffFile.ID); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("download: %v", err)
		}
	})

	t.Run("an account with no role", func(t *testing.T) {
		if _, _, err := f.svc.List(as(nobody), 1, port.FileListFilter{Limit: 10}); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("list: %v", err)
		}
		if _, err := f.svc.Get(as(nobody), 1, staffFile.ID); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("get: %v", err)
		}
		if _, err := f.svc.Upload(as(nobody), up(model.FileKindContract, "a.pdf", pdfBytes)); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("upload: %v", err)
		}
	})

	t.Run("a viewer reads but does not write", func(t *testing.T) {
		if _, _, err := f.svc.Download(as(viewer), 1, staffFile.ID); err != nil {
			t.Fatalf("download: %v", err)
		}
		if _, err := f.svc.Upload(as(viewer), up(model.FileKindContract, "a.pdf", pdfBytes)); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("upload: %v", err)
		}
		if err := f.svc.Delete(as(viewer), 1, staffFile.ID); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("delete: %v", err)
		}
	})

	t.Run("restricted kinds are invisible below manager", func(t *testing.T) {
		for _, p := range []auth.Principal{viewer, staff} {
			rows, total, err := f.svc.List(as(p), 1, port.FileListFilter{Limit: 50})
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range rows {
				if r.Kind == model.FileKindLegal {
					t.Fatalf("%s saw a legal file in the list", p.Name)
				}
			}
			if total != 1 {
				t.Fatalf("%s: total = %d, want 1 (the legal file is hidden from the count too)", p.Name, total)
			}
			// Hidden, not forbidden: its existence is not revealed.
			if _, err := f.svc.Get(as(p), 1, legal.ID); !errors.Is(err, common.ErrNotFound) {
				t.Fatalf("%s get legal: %v, want ErrNotFound", p.Name, err)
			}
			if _, _, err := f.svc.Download(as(p), 1, legal.ID); !errors.Is(err, common.ErrNotFound) {
				t.Fatalf("%s download legal: %v, want ErrNotFound", p.Name, err)
			}
			if err := f.svc.Delete(as(p), 1, legal.ID); !errors.Is(err, common.ErrNotFound) {
				t.Fatalf("%s delete legal: %v, want ErrNotFound", p.Name, err)
			}
		}
		if _, err := f.svc.Upload(as(staff), up(model.FileKindLegal, "x.pdf", pdfBytes)); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("staff uploading a legal file: %v", err)
		}
		if _, err := f.svc.Get(as(manager), 1, legal.ID); err != nil {
			t.Fatalf("manager get legal: %v", err)
		}
	})

	t.Run("staff delete their own files only", func(t *testing.T) {
		if err := f.svc.Delete(as(staff2), 1, staffFile.ID); !errors.Is(err, common.ErrForbidden) {
			t.Fatalf("another staff member deleting: %v", err)
		}
		mine := f.mustUpload(t, staff2, up(model.FileKindAttachment, "mine.pdf", pdfBytes))
		if err := f.svc.Delete(as(staff2), 1, mine.ID); err != nil {
			t.Fatalf("deleting their own: %v", err)
		}
	})

	t.Run("a manager deletes anyone's", func(t *testing.T) {
		theirs := f.mustUpload(t, staff, up(model.FileKindAttachment, "theirs.pdf", pdfBytes))
		if err := f.svc.Delete(as(manager), 1, theirs.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("a file of another contract is invisible", func(t *testing.T) {
		if _, err := f.svc.Get(as(manager), 2, staffFile.ID); !errors.Is(err, common.ErrNotFound) {
			t.Fatalf("get via the wrong contract: %v", err)
		}
	})

	t.Run("refusals are audited", func(t *testing.T) {
		if n := len(f.eventsOf(t, model.FileEventDenied)); n < 3 {
			t.Fatalf("denied events = %d, want at least 3", n)
		}
	})
}

func TestAudit_OnlyManagersRead(t *testing.T) {
	f := newFilesFixture(t)
	f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))

	for _, p := range []auth.Principal{viewer, staff, nobody} {
		if _, _, err := f.svc.Audit(as(p), 1, port.FileAuditFilter{Limit: 50}); !errors.Is(err, common.ErrForbidden) {
			t.Errorf("%s: %v, want ErrForbidden", p.Name, err)
		}
	}
	rows, total, err := f.svc.Audit(as(manager), 1, port.FileAuditFilter{Limit: 50})
	if err != nil || total < 1 || rows[0].Action != model.FileEventUploaded {
		t.Fatalf("audit: %+v, total %d, err %v", rows, total, err)
	}
	if _, _, err := f.svc.Audit(as(manager), 9, port.FileAuditFilter{Limit: 50}); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("unknown contract: %v", err)
	}
}

// ---- history ---------------------------------------------------------------

func TestVersions(t *testing.T) {
	f := newFilesFixture(t)
	v1 := f.mustUpload(t, staff, up(model.FileKindContract, "hd.pdf", pdfBytes))
	in := up(model.FileKindContract, "hd-v2.pdf", append([]byte{}, pdfBytes...))
	in.ReplacesID = v1.ID
	v2 := f.mustUpload(t, staff, in)

	if v2.Version != 2 || v2.LineageID != v1.ID || v2.ReplacesID == nil || *v2.ReplacesID != v1.ID || !v2.Latest {
		t.Fatalf("v2 = %+v", v2)
	}
	old, _ := f.svc.Get(as(staff), 1, v1.ID)
	if old.Latest {
		t.Fatal("the replaced version is still marked latest")
	}

	// Only the latest is listed unless asked for all.
	rows, total, _ := f.svc.List(as(staff), 1, port.FileListFilter{Limit: 50})
	if total != 1 || rows[0].ID != v2.ID {
		t.Fatalf("list = %+v (total %d)", rows, total)
	}
	if _, total, _ = f.svc.List(as(staff), 1, port.FileListFilter{Limit: 50, AllVersions: true}); total != 2 {
		t.Fatalf("all versions: total %d", total)
	}
	// An old version can still be downloaded.
	if _, _, err := f.svc.Download(as(staff), 1, v1.ID); err != nil {
		t.Fatal(err)
	}

	// A third version chains on.
	in3 := up(model.FileKindContract, "hd-v3.pdf", pdfBytes)
	in3.ReplacesID = v2.ID
	v3 := f.mustUpload(t, staff, in3)
	if v3.Version != 3 || v3.LineageID != v1.ID {
		t.Fatalf("v3 = %+v", v3)
	}

	// What cannot be replaced.
	stale := up(model.FileKindContract, "again.pdf", pdfBytes)
	stale.ReplacesID = v1.ID
	if _, err := f.svc.Upload(as(staff), stale); !errors.Is(err, common.ErrInvalidTransition) {
		t.Errorf("replacing an old version: %v, want ErrInvalidTransition", err)
	}
	other := up(model.FileKindContract, "x.pdf", pdfBytes)
	other.ReplacesID = v3.ID
	if _, err := f.svc.Upload(as(staff2), other); !errors.Is(err, common.ErrForbidden) {
		t.Errorf("replacing someone else's file: %v, want ErrForbidden", err)
	}
	if _, err := f.svc.Upload(as(manager), other); err != nil {
		t.Errorf("a manager replacing it: %v", err)
	}
	wrongKind := up(model.FileKindAppendix, "x.pdf", pdfBytes)
	wrongKind.ReplacesID = v3.ID
	if _, err := f.svc.Upload(as(staff), wrongKind); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("replacing across kinds: %v, want ErrInvalidRequest", err)
	}
	wrongContract := up(model.FileKindContract, "x.pdf", pdfBytes)
	wrongContract.ContractID, wrongContract.ReplacesID = 2, v3.ID
	if _, err := f.svc.Upload(as(staff), wrongContract); !errors.Is(err, common.ErrNotFound) {
		t.Errorf("replacing across contracts: %v, want ErrNotFound", err)
	}
}

func TestHistory(t *testing.T) {
	f := newFilesFixture(t)
	v1 := f.mustUpload(t, staff, up(model.FileKindContract, "hd.pdf", pdfBytes))
	in := up(model.FileKindContract, "hd-v2.pdf", pdfBytes)
	in.ReplacesID = v1.ID
	v2 := f.mustUpload(t, staff, in)
	if _, _, err := f.svc.Download(as(viewer), 1, v2.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.svc.Delete(as(manager), 1, v2.ID); err != nil {
		t.Fatal(err)
	}

	// Ask through the surviving version: the deleted one is part of the history.
	h, err := f.svc.History(as(viewer), 1, v1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Versions) != 2 || h.Versions[0].ID != v1.ID || h.Versions[1].ID != v2.ID || !h.Versions[1].DeletedAt.Valid || h.Versions[0].DeletedAt.Valid {
		t.Fatalf("versions = %+v", h.Versions)
	}
	var actions []string
	for _, e := range h.Events {
		actions = append(actions, e.Action)
	}
	want := []string{model.FileEventUploaded, model.FileEventNewVersion, model.FileEventDownloaded, model.FileEventDeleted}
	if strings.Join(actions, ",") != strings.Join(want, ",") {
		t.Fatalf("events = %v, want %v", actions, want)
	}

	// Who did it, from where.
	up1, dl, del := h.Events[0], h.Events[2], h.Events[3]
	if up1.ActorID != staff.Subject || up1.ActorName != "Sam" || up1.IP != "10.0.0.5" || up1.ActorRoles != "staff" || up1.FileID != v1.ID || up1.Version != 1 {
		t.Fatalf("upload event = %+v", up1)
	}
	if !strings.Contains(up1.Detail, "sha256") || !strings.Contains(h.Events[1].Detail, "replaces file") {
		t.Fatalf("details: %q / %q", up1.Detail, h.Events[1].Detail)
	}
	if dl.ActorID != viewer.Subject || del.ActorID != manager.Subject || dl.LineageID != v1.ID || del.LineageID != v1.ID {
		t.Fatalf("download %+v, delete %+v", dl, del)
	}

	// Deleting the newest version brought the previous one back.
	cur, _ := f.svc.Get(as(viewer), 1, v1.ID)
	if !cur.Latest {
		t.Fatal("the previous version did not become the latest again")
	}
	if rows, total, _ := f.svc.List(as(viewer), 1, port.FileListFilter{Limit: 10}); total != 1 || rows[0].ID != v1.ID {
		t.Fatalf("list after deleting v2: %+v", rows)
	}

	// The history of a file the caller cannot read is hidden.
	legal := f.mustUpload(t, manager, up(model.FileKindLegal, "l.pdf", pdfBytes))
	if _, err := f.svc.History(as(staff), 1, legal.ID); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("history of a legal file as staff: %v", err)
	}
}

func TestAudit_FiltersAndRecordsRejections(t *testing.T) {
	f := newFilesFixture(t)
	a := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))
	f.mustUpload(t, staff, up(model.FileKindContract, "b.pdf", pdfBytes))
	_, _ = f.svc.Upload(as(staff), up(model.FileKindContract, "bad.pdf", []byte("MZ nope")))
	_, _, _ = f.svc.Download(as(viewer), 1, a.ID)

	all, total, _ := f.svc.Audit(as(manager), 1, port.FileAuditFilter{Limit: 100})
	if total != 4 {
		t.Fatalf("entries = %d, want 4 (2 uploads, 1 rejection, 1 download): %+v", total, all)
	}
	byFile, _, _ := f.svc.Audit(as(manager), 1, port.FileAuditFilter{FileID: a.ID, Limit: 100})
	if len(byFile) != 2 {
		t.Fatalf("entries for file %d: %+v", a.ID, byFile)
	}
	rej, _, _ := f.svc.Audit(as(manager), 1, port.FileAuditFilter{Action: model.FileEventRejected, Limit: 100})
	if len(rej) != 1 || rej[0].FileID != 0 || !strings.Contains(rej[0].Detail, "bad.pdf") || rej[0].ActorID != staff.Subject {
		t.Fatalf("rejection entry = %+v", rej)
	}
}

// failEvents refuses to write audit entries.
type failEvents struct {
	port.Repository[model.ContractFileEvent]
}

func (failEvents) Create(context.Context, *model.ContractFileEvent) error { return common.ErrDBFailed }

func TestUpload_IsAtomicWithItsAuditEntry(t *testing.T) {
	f := newFilesFixture(t)
	f.svc.events = failEvents{f.env.events}

	if _, err := f.svc.Upload(as(staff), up(model.FileKindContract, "a.pdf", pdfBytes)); !errors.Is(err, common.ErrDBFailed) {
		t.Fatalf("err = %v", err)
	}
	if f.env.docs.count() != 0 {
		t.Fatal("the file was recorded although its audit entry was not")
	}
	if len(f.store.Objects) != 0 {
		t.Fatalf("the uploaded bytes were left behind: %v", f.store.Objects)
	}
}

func TestDownload_ChecksTheHash(t *testing.T) {
	f := newFilesFixture(t)
	row := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))

	if _, data, err := f.svc.Download(as(viewer), 1, row.ID); err != nil || !bytes.Equal(data, pdfBytes) {
		t.Fatalf("download: %v", err)
	}
	f.store.Tamper(row.ObjectKey, []byte("%PDF-1.7 forged"))
	if _, _, err := f.svc.Download(as(viewer), 1, row.ID); !errors.Is(err, common.ErrIntegrity) {
		t.Fatalf("tampered: %v, want ErrIntegrity", err)
	}
	if n := len(f.eventsOf(t, model.FileEventDownloaded)); n != 1 {
		t.Fatalf("downloads recorded = %d, want only the successful one", n)
	}
}

func TestDelete_SoftDeletesAndProtectsSystemFiles(t *testing.T) {
	f := newFilesFixture(t)
	mine := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))

	evidence, err := f.svc.Prepare(context.Background(), port.FileSpec{
		ContractID: 1, Kind: model.FileKindRenewal, Name: "r.pdf", Source: model.FileSourceSystem, Data: pdfBytes,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.svc.Persist(context.Background(), evidence); err != nil {
		t.Fatal(err)
	}

	if err := f.svc.Delete(as(manager), 1, evidence.ID); !errors.Is(err, common.ErrForbidden) {
		t.Fatalf("deleting evidence, even as a manager: %v", err)
	}
	if err := f.svc.Delete(as(staff), 1, mine.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Get(as(staff), 1, mine.ID); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
	if len(f.store.Deleted) != 0 {
		t.Fatal("a soft delete removed the stored bytes")
	}
	// The row is still there for the history.
	if h, err := f.svc.History(as(manager), 1, evidence.ID); err != nil || len(h.Versions) != 1 {
		t.Fatalf("history: %v", err)
	}
}

func TestPersist_RenewalsChainAsVersions(t *testing.T) {
	f := newFilesFixture(t)
	ctx := context.Background()
	var ids []uint
	for i := 0; i < 3; i++ {
		row, err := f.svc.Prepare(ctx, port.FileSpec{
			ContractID: 1, Kind: model.FileKindRenewal, Name: "r.pdf", Source: model.FileSourceSystem, Data: pdfBytes,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := f.svc.Persist(ctx, row); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, row.ID)
	}
	latest, err := f.svc.Latest(ctx, 1, model.FileKindRenewal)
	if err != nil || latest.ID != ids[2] || latest.Version != 3 || latest.LineageID != ids[0] {
		t.Fatalf("latest = %+v, %v", latest, err)
	}
	h, _ := f.svc.History(as(manager), 1, ids[2])
	if len(h.Versions) != 3 || len(h.Events) != 3 || h.Events[2].Action != model.FileEventNewVersion || h.Events[2].ActorID != "system" {
		t.Fatalf("history = %+v", h)
	}
}

// ---- scanning and inspection ----------------------------------------------

func TestScan(t *testing.T) {
	t.Run("an infected file is refused and leaves nothing behind", func(t *testing.T) {
		f := newFilesFixture(t)
		f.scanner.on = true
		f.scanner.res = port.ScanResult{Clean: false, Signature: "Win.Test.EICAR_HDB-1", Engine: "clamav"}

		_, err := f.svc.Upload(as(staff), up(model.FileKindContract, "a.pdf", pdfBytes))
		if !errors.Is(err, common.ErrMalicious) || !strings.Contains(err.Error(), "EICAR") {
			t.Fatalf("err = %v, want ErrMalicious naming the signature", err)
		}
		if f.env.docs.count() != 0 || len(f.store.Objects) != 0 {
			t.Fatal("an infected file was stored")
		}
		if ev := f.eventsOf(t, model.FileEventRejected); len(ev) != 1 || !strings.Contains(ev[0].Detail, "EICAR") {
			t.Fatalf("rejection not audited: %+v", ev)
		}
	})

	t.Run("required scanner down: uploads are refused", func(t *testing.T) {
		f := newFilesFixture(t)
		f.scanner.on, f.scanner.err = true, errors.New("connection refused")
		if _, err := f.svc.Upload(as(staff), up(model.FileKindContract, "a.pdf", pdfBytes)); !errors.Is(err, common.ErrUpstream) {
			t.Fatalf("err = %v, want ErrUpstream", err)
		}
		if f.env.docs.count() != 0 {
			t.Fatal("a file was stored without being scanned")
		}
	})

	t.Run("optional scanner down: accepted and marked unscanned", func(t *testing.T) {
		f := newFilesFixture(t, func(c *config.FilesConfig) { c.ScanRequired = false })
		f.scanner.on, f.scanner.err = true, errors.New("connection refused")
		row := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))
		if row.ScanStatus != model.ScanSkipped || row.ScannedAt != nil {
			t.Fatalf("row = %+v", row)
		}
	})

	t.Run("no scanner configured", func(t *testing.T) {
		f := newFilesFixture(t)
		if row := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes)); row.ScanStatus != model.ScanSkipped {
			t.Fatalf("row = %+v", row)
		}
	})

	t.Run("a clean file is scanned once", func(t *testing.T) {
		f := newFilesFixture(t)
		f.scanner.on = true
		row := f.mustUpload(t, staff, up(model.FileKindContract, "a.pdf", pdfBytes))
		if f.scanner.calls != 1 || row.ScanStatus != model.ScanClean {
			t.Fatalf("calls %d, row %+v", f.scanner.calls, row)
		}
	})

	t.Run("active content is refused before the scanner is asked", func(t *testing.T) {
		f := newFilesFixture(t)
		f.scanner.on = true
		if _, err := f.svc.Upload(as(staff), up(model.FileKindContract, "js.pdf", pdfWithJS(t))); !errors.Is(err, common.ErrMalicious) {
			t.Fatalf("err = %v", err)
		}
		if f.scanner.calls != 0 {
			t.Fatal("the scanner was called for a file already refused")
		}
	})
}

// validPDF is a small well-formed PDF, as a scanner or word processor would
// produce (the file inspection must accept ordinary documents).
func validPDF() []byte {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(40, 10, "Hop dong")
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
