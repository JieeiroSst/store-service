package handler

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"howett.net/plist"

	"github.com/JIeeiroSst/tool-service/internal/mobile"
)

const (
	maxMobileUpload = 220 << 20
	maxUnzipTotal   = 1 << 30
	maxUnzipEntries = 100000
)

func (h *Handler) registerMobile(q *gin.RouterGroup) {
	m := q.Group("/mobile")
	m.GET("/devices", h.mobileDevices)
	m.POST("/static", h.mobileStatic)
	m.POST("/android", h.mobileAndroid)
	m.POST("/ios", h.mobileIOS)
}

func mobileGate(fs []mobile.Finding, blockUnreviewed bool) *gate {
	var fail, review []string
	for _, f := range fs {
		switch f.Severity {
		case mobile.High:
			fail = append(fail, f.ID+": "+f.Title)
		case mobile.Medium:
			review = append(review, f.ID+": "+f.Title)
		}
	}
	return combine(nil, blockUnreviewed, capList(fail, 20), capList(review, 20))
}

func (h *Handler) mobileDevices(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	out := gin.H{}
	if bin, err := mobile.FindADB(); err != nil {
		out["android"] = gin.H{"available": false, "error": err.Error()}
	} else if ds, err := mobile.ListAndroidDevices(ctx, mobile.ExecADB{Bin: bin}); err != nil {
		out["android"] = gin.H{"available": false, "error": err.Error()}
	} else {
		out["android"] = gin.H{"available": true, "devices": ds}
	}
	if err := mobile.FindSimctl(); err != nil {
		out["ios"] = gin.H{"available": false, "error": err.Error()}
	} else if sims, err := mobile.ListSimulators(ctx, mobile.ExecSimctl{}); err != nil {
		out["ios"] = gin.H{"available": false, "error": err.Error()}
	} else {
		out["ios"] = gin.H{"available": true, "simulators": sims}
	}
	c.JSON(http.StatusOK, out)
}

func readUpload(c *gin.Context, field string) (name string, data []byte, err error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return "", nil, fmt.Errorf("upload the app in form field %q", field)
	}
	f, err := fh.Open()
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	data, err = io.ReadAll(io.LimitReader(f, maxMobileUpload+1))
	if err != nil {
		return "", nil, err
	}
	if len(data) > maxMobileUpload {
		return "", nil, fmt.Errorf("file larger than %d MB", maxMobileUpload>>20)
	}
	return fh.Filename, data, nil
}

func (h *Handler) mobileStatic(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMobileUpload+(1<<20))
	name, data, err := readUpload(c, "file")
	if err != nil {
		badRequest(c, err)
		return
	}
	var rep *mobile.StaticReport
	switch strings.ToLower(filepath.Ext(name)) {
	case ".apk":
		rep, err = mobile.AnalyzeAPK(data)
	case ".ipa":
		rep, err = mobile.AnalyzeIPA(data)
	case ".aab":
		err = fmt.Errorf("Android App Bundles cannot be analysed directly: build a universal APK (bundletool build-apks --mode=universal) and upload that")
	default:
		err = fmt.Errorf("unsupported file type %q (use .apk or .ipa)", filepath.Ext(name))
	}
	if err != nil {
		badRequest(c, err)
		return
	}
	status := http.StatusOK
	g := mobileGate(rep.Findings, c.Query("unreviewed") != "warn")
	if !g.Passed && c.Query("fail_http") == "true" {
		status = http.StatusUnprocessableEntity
	}
	c.JSON(status, gin.H{"report": rep, "gate": g})
}

func formList(c *gin.Context, key string) []string {
	var out []string
	for _, s := range strings.Split(c.PostForm(key), ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func (h *Handler) mobileAndroid(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMobileUpload+(1<<20))
	bin, err := mobile.FindADB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	opt := mobile.AndroidOptions{
		Package: strings.TrimSpace(c.PostForm("package")), Component: strings.TrimSpace(c.PostForm("component")),
		Scenarios: formList(c, "scenarios"), Uninstall: c.PostForm("uninstall") == "true",
		MonkeyEvents: 300, ArtifactDir: filepath.Join(h.infra.DataDir, "artifacts"),
	}
	if v := c.PostForm("monkey_events"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			badRequest(c, fmt.Errorf("monkey_events must be a number"))
			return
		}
		opt.MonkeyEvents = n
	}
	if v := c.PostForm("seed"); v != "" {
		opt.Seed, _ = strconv.ParseInt(v, 10, 64)
	}
	serial := strings.TrimSpace(c.PostForm("serial"))
	if serial != "" && !mobile.ValidSerial(serial) {
		badRequest(c, fmt.Errorf("invalid device serial"))
		return
	}
	blockUnreviewed := c.PostForm("unreviewed") != "warn"

	var static *mobile.StaticReport
	var apkPath string
	{
		if fh, err := c.FormFile("file"); err == nil && fh != nil {
			_, data, err := readUpload(c, "file")
			if err != nil {
				badRequest(c, err)
				return
			}
			static, err = mobile.AnalyzeAPK(data)
			if err != nil {
				badRequest(c, err)
				return
			}
			if opt.Package == "" {
				opt.Package = static.Android.Package
			}
			if opt.Component == "" && static.Android.Launcher != "" && opt.Package == static.Android.Package {
				opt.Component = opt.Package + "/" + static.Android.Launcher
			}
			tmp, err := os.CreateTemp("", "qa-*.apk")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			tmp.Write(data)
			tmp.Close()
			apkPath = tmp.Name()
			opt.APKPath = "/data/local/tmp/qa-upload.apk"
		}
	}
	if !mobile.ValidPackage(opt.Package) {
		if apkPath != "" {
			os.Remove(apkPath)
		}
		badRequest(c, fmt.Errorf("package is required (e.g. com.example.app) unless an APK is uploaded"))
		return
	}

	h.runQA(c, "android", 25*time.Minute, func(ctx context.Context, progress func(string)) (gin.H, error) {
		if apkPath != "" {
			defer os.Remove(apkPath)
		}
		opt.Progress = progress
		base := mobile.ExecADB{Bin: bin}
		devs, err := mobile.ListAndroidDevices(ctx, base)
		if err != nil {
			return nil, err
		}
		var chosen string
		for _, d := range devs {
			if d.State == "device" && (serial == "" || d.Serial == serial) {
				chosen = d.Serial
				break
			}
		}
		if chosen == "" {
			return nil, fmt.Errorf("no Android device or emulator in state 'device' (adb sees %d); set ANDROID_ADB_SERVER_ADDRESS to reach a remote adb server", len(devs))
		}
		r := mobile.ExecADB{Bin: bin, Serial: chosen}
		if apkPath != "" {
			if out, err := r.Run(ctx, "push", apkPath, opt.APKPath); err != nil {
				return nil, fmt.Errorf("cannot copy APK to the device: %s", strings.TrimSpace(string(out)))
			}
			defer func() { _, _ = r.Run(context.Background(), "shell", "rm", "-f", opt.APKPath) }()
		}
		rep, err := mobile.RunAndroid(ctx, r, opt)
		if err != nil {
			return nil, inputError{err}
		}
		fs := rep.Findings
		if static != nil {
			fs = append(append([]mobile.Finding(nil), static.Findings...), fs...)
		}
		return gin.H{"report": rep, "static": static, "gate": mobileGate(fs, blockUnreviewed)}, nil
	})
}

func (h *Handler) mobileIOS(c *gin.Context) {
	if !h.needsToken(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMobileUpload+(1<<20))
	if err := mobile.FindSimctl(); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
	opt := mobile.IOSOptions{
		BundleID: strings.TrimSpace(c.PostForm("bundle_id")), Device: strings.TrimSpace(c.PostForm("device")),
		Scenarios: formList(c, "scenarios"), DeepLinks: formList(c, "deep_links"),
		ArtifactDir: filepath.Join(h.infra.DataDir, "artifacts"),
	}
	blockUnreviewed := c.PostForm("unreviewed") != "warn"

	var workDir string
	if fh, err := c.FormFile("file"); err == nil && fh != nil {
		_, data, err := readUpload(c, "file")
		if err != nil {
			badRequest(c, err)
			return
		}
		workDir, err = os.MkdirTemp("", "qa-ios-")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		appDir, err := unzipApp(data, workDir)
		if err != nil {
			os.RemoveAll(workDir)
			badRequest(c, err)
			return
		}
		info, err := readInfoPlist(appDir)
		if err != nil {
			os.RemoveAll(workDir)
			badRequest(c, err)
			return
		}
		opt.AppPath = appDir
		if opt.BundleID == "" {
			opt.BundleID = info.bundle
		}
		opt.ExecName = info.exec
	}
	if !mobile.ValidBundleID(opt.BundleID) {
		if workDir != "" {
			os.RemoveAll(workDir)
		}
		badRequest(c, fmt.Errorf("bundle_id is required (e.g. com.example.app) unless a simulator .app is uploaded"))
		return
	}

	h.runQA(c, "ios", 25*time.Minute, func(ctx context.Context, progress func(string)) (gin.H, error) {
		if workDir != "" {
			defer os.RemoveAll(workDir)
		}
		opt.Progress = progress
		rep, err := mobile.RunIOS(ctx, mobile.ExecSimctl{}, opt)
		if err != nil {
			return nil, inputError{err}
		}
		return gin.H{"report": rep, "gate": mobileGate(rep.Findings, blockUnreviewed)}, nil
	})
}

type appInfo struct{ bundle, exec string }

func readInfoPlist(appDir string) (appInfo, error) {
	b, err := os.ReadFile(filepath.Join(appDir, "Info.plist"))
	if err != nil {
		return appInfo{}, fmt.Errorf("Info.plist not found in the .app bundle")
	}
	var info map[string]any
	if _, err := plist.Unmarshal(b, &info); err != nil {
		return appInfo{}, fmt.Errorf("cannot parse Info.plist: %w", err)
	}
	bundle, _ := info["CFBundleIdentifier"].(string)
	exe, _ := info["CFBundleExecutable"].(string)
	if bundle == "" {
		return appInfo{}, fmt.Errorf("Info.plist has no CFBundleIdentifier")
	}
	return appInfo{bundle, exe}, nil
}

func unzipApp(data []byte, dest string) (string, error) {
	zr, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid zip: %w", err)
	}
	if len(zr.File) > maxUnzipEntries {
		return "", fmt.Errorf("archive has too many entries")
	}
	root := filepath.Clean(dest) + string(os.PathSeparator)
	var total int64
	var appDir string
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "Payload/") && strings.HasSuffix(strings.SplitN(f.Name, "/", 3)[1], ".app") {
			return "", fmt.Errorf("this is a device .ipa; the simulator needs a .app built for the simulator, zipped")
		}
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, root) {
			return "", fmt.Errorf("archive entry escapes the extraction directory: %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
			continue
		}
		if f.Mode()&os.ModeSymlink != 0 {
			continue
		}
		total += int64(f.UncompressedSize64)
		if total > maxUnzipTotal {
			return "", fmt.Errorf("archive expands beyond %d MB", maxUnzipTotal>>20)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode().Perm()|0o600)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, err = io.Copy(out, io.LimitReader(rc, int64(f.UncompressedSize64)+1))
		rc.Close()
		out.Close()
		if err != nil {
			return "", err
		}
		if appDir == "" {
			if d := appRoot(f.Name); d != "" {
				appDir = filepath.Join(dest, d)
			}
		}
	}
	if appDir == "" {
		return "", fmt.Errorf("no .app bundle found in the archive")
	}
	return appDir, nil
}

func appRoot(name string) string {
	parts := strings.Split(filepath.ToSlash(name), "/")
	for i, p := range parts {
		if strings.HasSuffix(p, ".app") {
			return filepath.Join(parts[:i+1]...)
		}
	}
	return ""
}
