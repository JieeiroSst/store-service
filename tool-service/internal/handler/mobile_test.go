package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func upload(r *gin.Engine, path, token, field, filename string, data []byte, fields map[string]string) *httptest.ResponseRecorder {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		w.WriteField(k, v)
	}
	if data != nil {
		fw, _ := w.CreateFormFile(field, filename)
		fw.Write(data)
	}
	w.Close()
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestMobileEndpointsNeedToken(t *testing.T) {
	r := newTestHandler(t, "", nil)
	for _, p := range []string{"/api/v1/qa/mobile/static", "/api/v1/qa/mobile/android", "/api/v1/qa/mobile/ios"} {
		if w := upload(r, p, "", "file", "a.apk", []byte("x"), nil); w.Code != http.StatusForbidden {
			t.Errorf("%s without API_TOKEN: %d", p, w.Code)
		}
	}
	rt := newTestHandler(t, "tok", nil)
	if w := upload(rt, "/api/v1/qa/mobile/static", "wrong", "file", "a.apk", []byte("x"), nil); w.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: %d", w.Code)
	}
}

func TestStaticAPKUpload(t *testing.T) {
	apk, err := os.ReadFile("../mobile/testdata/helloworld.apk")
	if err != nil {
		t.Fatal(err)
	}
	r := newTestHandler(t, "tok", nil)
	w := upload(r, "/api/v1/qa/mobile/static", "tok", "file", "app.apk", apk, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var out struct {
		Report struct {
			Platform string `json:"platform"`
			Android  struct {
				Package string `json:"package"`
			} `json:"android"`
		} `json:"report"`
		Gate struct {
			Status string `json:"status"`
		} `json:"gate"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	if out.Report.Platform != "android" || out.Report.Android.Package != "com.example.helloworld" || out.Gate.Status == "" {
		t.Fatalf("%s", w.Body)
	}

	for name, tc := range map[string]struct {
		file string
		data []byte
	}{"aab": {"a.aab", []byte("x")}, "exe": {"a.exe", []byte("x")}, "corrupt": {"a.apk", []byte("junk")}, "ipa corrupt": {"a.ipa", []byte("junk")}} {
		if w := upload(r, "/api/v1/qa/mobile/static", "tok", "file", tc.file, tc.data, nil); w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", name, w.Code, w.Body)
		}
	}
	if w := upload(r, "/api/v1/qa/mobile/static", "tok", "file", "", nil, nil); w.Code != http.StatusBadRequest {
		t.Errorf("missing file: %d", w.Code)
	}
}

func TestStaticGateFailOnDebuggableApp(t *testing.T) {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	w, _ := zw.Create("AndroidManifest.xml")
	w.Write([]byte(`<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.acme.dbg"><uses-sdk android:minSdkVersion="26" android:targetSdkVersion="35"/><application android:debuggable="true" android:allowBackup="false"><activity android:name=".M"/></application></manifest>`))
	zw.Close()
	r := newTestHandler(t, "tok", nil)
	rec := upload(r, "/api/v1/qa/mobile/static?fail_http=true", "tok", "file", "dbg.apk", b.Bytes(), nil)
	if rec.Code != http.StatusUnprocessableEntity || !strings.Contains(rec.Body.String(), "MOB-AND-DBG-001") {
		t.Fatalf("debuggable release must fail the gate: %d %s", rec.Code, rec.Body)
	}
}

func TestAndroidDynamicRequiresADB(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("ADB_PATH", "")
	r := newTestHandler(t, "tok", nil)
	if w := upload(r, "/api/v1/qa/mobile/android", "tok", "", "", nil, map[string]string{"package": "com.acme.app"}); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("without adb: %d %s", w.Code, w.Body)
	}
}

func TestUnzipAppRejectsUnsafeArchives(t *testing.T) {
	mk := func(entries map[string]string) []byte {
		var b bytes.Buffer
		zw := zip.NewWriter(&b)
		for n, c := range entries {
			w, _ := zw.Create(n)
			w.Write([]byte(c))
		}
		zw.Close()
		return b.Bytes()
	}
	dest := t.TempDir()
	if _, err := unzipApp(mk(map[string]string{"../evil.txt": "x", "Shop.app/Info.plist": "x"}), dest); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Errorf("zip-slip must be refused: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dest), "evil.txt")); err == nil {
		t.Error("file was written outside the extraction directory")
	}
	if _, err := unzipApp(mk(map[string]string{"Payload/Shop.app/Info.plist": "x"}), t.TempDir()); err == nil || !strings.Contains(err.Error(), "device .ipa") {
		t.Errorf("device ipa: %v", err)
	}
	if _, err := unzipApp(mk(map[string]string{"readme.txt": "x"}), t.TempDir()); err == nil {
		t.Error("archive without .app must fail")
	}
	d2 := t.TempDir()
	app, err := unzipApp(mk(map[string]string{"build/Shop.app/Info.plist": "x", "build/Shop.app/Shop": "bin"}), d2)
	if err != nil || filepath.Base(app) != "Shop.app" {
		t.Fatalf("%q %v", app, err)
	}
	if _, err := os.Stat(filepath.Join(app, "Shop")); err != nil {
		t.Error("files not extracted")
	}
}
