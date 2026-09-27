package mobile

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"howett.net/plist"
)

type IOSApp struct {
	BundleID        string   `json:"bundle_id"`
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Build           string   `json:"build"`
	MinOS           string   `json:"minimum_os"`
	Executable      string   `json:"executable"`
	URLSchemes      []string `json:"url_schemes,omitempty"`
	BackgroundModes []string `json:"background_modes,omitempty"`
	UsageKeys       []string `json:"usage_descriptions,omitempty"`
	DetectedAPIs    []string `json:"detected_sensitive_apis,omitempty"`
	Frameworks      []string `json:"embedded_frameworks,omitempty"`
	Provisioning    string   `json:"provisioning,omitempty"`
	SizeMB          float64  `json:"size_mb"`
	Files           int      `json:"files"`
}

var apiRules = []struct {
	symbol string
	keys   []string
	label  string
}{
	{"AVCaptureDevice", []string{"NSCameraUsageDescription", "NSMicrophoneUsageDescription"}, "camera/microphone"},
	{"PHPhotoLibrary", []string{"NSPhotoLibraryUsageDescription", "NSPhotoLibraryAddUsageDescription"}, "photo library"},
	{"CLLocationManager", []string{"NSLocationWhenInUseUsageDescription", "NSLocationAlwaysAndWhenInUseUsageDescription", "NSLocationAlwaysUsageDescription"}, "location"},
	{"CNContactStore", []string{"NSContactsUsageDescription"}, "contacts"},
	{"EKEventStore", []string{"NSCalendarsUsageDescription", "NSCalendarsFullAccessUsageDescription", "NSCalendarsWriteOnlyAccessUsageDescription", "NSRemindersUsageDescription", "NSRemindersFullAccessUsageDescription"}, "calendar/reminders"},
	{"CBCentralManager", []string{"NSBluetoothAlwaysUsageDescription", "NSBluetoothPeripheralUsageDescription"}, "bluetooth"},
	{"CBPeripheralManager", []string{"NSBluetoothAlwaysUsageDescription", "NSBluetoothPeripheralUsageDescription"}, "bluetooth"},
	{"LAContext", []string{"NSFaceIDUsageDescription"}, "Face ID"},
	{"requestRecordPermission", []string{"NSMicrophoneUsageDescription"}, "microphone"},
	{"SFSpeechRecognizer", []string{"NSSpeechRecognitionUsageDescription"}, "speech recognition"},
	{"MPMediaLibrary", []string{"NSAppleMusicUsageDescription"}, "media library"},
	{"HKHealthStore", []string{"NSHealthShareUsageDescription", "NSHealthUpdateUsageDescription"}, "HealthKit"},
	{"CMMotionActivityManager", []string{"NSMotionUsageDescription"}, "motion"},
	{"ATTrackingManager", []string{"NSUserTrackingUsageDescription"}, "app tracking"},
	{"NFCNDEFReaderSession", []string{"NFCReaderUsageDescription"}, "NFC"},
}

var provXML = regexp.MustCompile(`(?s)<\?xml.*?</plist>`)

func AnalyzeIPA(data []byte) (*StaticReport, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a valid IPA/zip: %w", err)
	}
	if len(zr.File) > maxEntries {
		return nil, errors.New("archive has too many entries")
	}
	var appDir string
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "Payload/") && strings.HasSuffix(f.Name, ".app/Info.plist") && strings.Count(f.Name, "/") == 2 {
			appDir = strings.TrimSuffix(f.Name, "Info.plist")
			break
		}
	}
	if appDir == "" {
		return nil, errors.New("Payload/<name>.app/Info.plist not found: not an IPA")
	}
	read := func(name string, limit int64) []byte {
		for _, f := range zr.File {
			if f.Name == name {
				rc, err := f.Open()
				if err != nil {
					return nil
				}
				defer rc.Close()
				b, _ := io.ReadAll(io.LimitReader(rc, limit))
				return b
			}
		}
		return nil
	}

	var info map[string]any
	if _, err := plist.Unmarshal(read(appDir+"Info.plist", 8<<20), &info); err != nil {
		return nil, fmt.Errorf("cannot parse Info.plist: %w", err)
	}
	str := func(k string) string { s, _ := info[k].(string); return s }
	app := &IOSApp{
		BundleID: str("CFBundleIdentifier"), Name: orStr(str("CFBundleDisplayName"), str("CFBundleName")),
		Version: str("CFBundleShortVersionString"), Build: str("CFBundleVersion"), MinOS: str("MinimumOSVersion"),
		Executable: str("CFBundleExecutable"), SizeMB: float64(int(float64(len(data))/1024/1024*10+0.5)) / 10, Files: len(zr.File),
	}
	c := &collector{}

	for k := range info {
		if strings.HasPrefix(k, "NS") && strings.HasSuffix(k, "UsageDescription") || k == "NFCReaderUsageDescription" {
			app.UsageKeys = append(app.UsageKeys, k)
		}
	}
	sort.Strings(app.UsageKeys)
	if modes, ok := info["UIBackgroundModes"].([]any); ok {
		for _, m := range modes {
			if s, ok := m.(string); ok {
				app.BackgroundModes = append(app.BackgroundModes, s)
			}
		}
	}
	if types, ok := info["CFBundleURLTypes"].([]any); ok {
		for _, t := range types {
			if m, ok := t.(map[string]any); ok {
				if ss, ok := m["CFBundleURLSchemes"].([]any); ok {
					for _, s := range ss {
						if v, ok := s.(string); ok {
							app.URLSchemes = append(app.URLSchemes, v)
						}
					}
				}
			}
		}
	}

	if ats, ok := info["NSAppTransportSecurity"].(map[string]any); ok {
		if b, _ := ats["NSAllowsArbitraryLoads"].(bool); b {
			c.add("MOB-IOS-ATS-001", High, "App Transport Security is disabled (NSAllowsArbitraryLoads)", "All HTTP traffic is allowed. Restrict it to specific exception domains.")
		}
		if b, _ := ats["NSAllowsLocalNetworking"].(bool); b {
			c.add("MOB-IOS-ATS-003", Info, "local networking exempted from ATS", "")
		}
		if ex, ok := ats["NSExceptionDomains"].(map[string]any); ok {
			var doms []string
			for d, v := range ex {
				if m, ok := v.(map[string]any); ok {
					if b, _ := m["NSExceptionAllowsInsecureHTTPLoads"].(bool); b {
						doms = append(doms, d)
					}
				}
			}
			sort.Strings(doms)
			if len(doms) > 0 {
				c.add("MOB-IOS-ATS-002", Medium, "insecure HTTP allowed for specific domains", strings.Join(doms, ", "))
			}
		}
	}
	if b, _ := info["UIFileSharingEnabled"].(bool); b {
		c.add("MOB-IOS-FS-001", Low, "documents folder is exposed through file sharing", "UIFileSharingEnabled=true")
	}
	if app.MinOS != "" && majorVersion(app.MinOS) > 0 && majorVersion(app.MinOS) < 14 {
		c.add("MOB-IOS-OS-001", Low, fmt.Sprintf("supports very old iOS (%s)", app.MinOS), "Old releases lack current security defaults.")
	}
	if app.Build == "" {
		c.add("MOB-IOS-VER-001", Medium, "CFBundleVersion is missing", "App Store Connect requires a build number.")
	}

	if app.Executable != "" {
		bin := read(appDir+app.Executable, 320<<20)
		have := map[string]bool{}
		for _, k := range app.UsageKeys {
			have[k] = true
		}
		seen := map[string]bool{}
		for _, r := range apiRules {
			if !bytes.Contains(bin, []byte(r.symbol)) {
				continue
			}
			app.DetectedAPIs = append(app.DetectedAPIs, r.symbol)
			ok := false
			for _, k := range r.keys {
				ok = ok || have[k]
			}
			if !ok && !seen[r.label] {
				seen[r.label] = true
				c.add("MOB-IOS-PRIV-001", High, fmt.Sprintf("uses %s APIs but Info.plist has no usage description", r.label),
					fmt.Sprintf("Add one of: %s. App Store review rejects (ITMS-90683) and the app crashes on first access.", strings.Join(r.keys, ", ")))
			}
		}
		sort.Strings(app.DetectedAPIs)
		if len(bin) == 0 {
			c.add("MOB-IOS-BIN-001", Low, "executable could not be read for API analysis", app.Executable)
		}
	}

	frameworks := map[string]bool{}
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, appDir+"Frameworks/") {
			if p := strings.Split(strings.TrimPrefix(f.Name, appDir+"Frameworks/"), "/"); len(p) > 0 && p[0] != "" {
				frameworks[p[0]] = true
			}
		}
	}
	for f := range frameworks {
		app.Frameworks = append(app.Frameworks, f)
	}
	sort.Strings(app.Frameworks)

	if read(appDir+"PrivacyInfo.xcprivacy", 1<<20) == nil {
		c.add("MOB-IOS-PRIV-002", Medium, "no PrivacyInfo.xcprivacy privacy manifest", "Apple requires a privacy manifest declaring required-reason API usage.")
	}
	provisioning(read(appDir+"embedded.mobileprovision", 4<<20), app, c)

	sc := scanArchive(zr, func(n string) bool {
		return strings.HasPrefix(n, "Payload/") && (strings.Contains(n, "/_CodeSignature/") || strings.HasSuffix(n, ".car") || strings.HasSuffix(n, ".png") || strings.HasSuffix(n, ".mobileprovision"))
	})
	for _, f := range sc.findings {
		c.list = append(c.list, f)
	}
	if len(sc.cleartext) > 0 {
		shown := sc.cleartext
		if len(shown) > 10 {
			shown = shown[:10]
		}
		c.add("MOB-IOS-NET-001", Low, fmt.Sprintf("%d plain-http host(s) referenced in the bundle", len(sc.cleartext)), strings.Join(shown, ", "))
	}
	return &StaticReport{Platform: "ios", IOS: app, Findings: c.sorted(), Scanned: sc.filesRead}, nil
}

func provisioning(raw []byte, app *IOSApp, c *collector) {
	if len(raw) == 0 {
		app.Provisioning = "none embedded (App Store or simulator build)"
		return
	}
	xmlPart := provXML.Find(raw)
	if xmlPart == nil {
		return
	}
	var p map[string]any
	if _, err := plist.Unmarshal(xmlPart, &p); err != nil {
		return
	}
	kind := "development"
	if _, ok := p["ProvisionedDevices"]; !ok {
		kind = "app store / enterprise"
	}
	if b, _ := p["ProvisionsAllDevices"].(bool); b {
		kind = "enterprise"
	}
	app.Provisioning = kind
	if ent, ok := p["Entitlements"].(map[string]any); ok {
		if b, _ := ent["get-task-allow"].(bool); b {
			c.add("MOB-IOS-DBG-001", High, "debug build: get-task-allow entitlement is true", "Debuggers can attach. Release builds must be signed with a distribution profile.")
		}
		if env, _ := ent["aps-environment"].(string); env == "development" {
			c.add("MOB-IOS-PUSH-001", Medium, "push notifications use the development APNs environment", "")
		}
	}
	if exp, ok := p["ExpirationDate"].(time.Time); ok {
		days := int(time.Until(exp).Hours() / 24)
		switch {
		case days < 0:
			c.add("MOB-IOS-PROV-001", High, "provisioning profile has expired", exp.Format("2006-01-02"))
		case days < 30:
			c.add("MOB-IOS-PROV-001", Medium, fmt.Sprintf("provisioning profile expires in %d days", days), exp.Format("2006-01-02"))
		}
	}
}

func majorVersion(v string) int {
	n := 0
	for _, r := range v {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func orStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
