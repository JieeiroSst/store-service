package mobile

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/shogo82148/androidbinary"
)

type Component struct {
	Kind          string   `json:"kind"`
	Name          string   `json:"name"`
	Exported      *bool    `json:"exported,omitempty"`
	Permission    string   `json:"permission,omitempty"`
	IntentFilters int      `json:"intent_filters"`
	Schemes       []string `json:"deep_link_schemes,omitempty"`
}

type AndroidApp struct {
	Package     string      `json:"package"`
	VersionName string      `json:"version_name"`
	VersionCode int         `json:"version_code"`
	MinSDK      int         `json:"min_sdk"`
	TargetSDK   int         `json:"target_sdk"`
	Permissions []string    `json:"permissions"`
	Components  []Component `json:"components"`
	Launcher    string      `json:"launcher_activity,omitempty"`
	ABIs        []string    `json:"native_abis,omitempty"`
	SizeMB      float64     `json:"size_mb"`
	Files       int         `json:"files"`
	Signed      bool        `json:"signed"`
	SignedV2    bool        `json:"signed_v2_or_later"`
}

type StaticReport struct {
	Platform string      `json:"platform"`
	Android  *AndroidApp `json:"android,omitempty"`
	IOS      *IOSApp     `json:"ios,omitempty"`
	Findings []Finding   `json:"findings"`
	Scanned  int         `json:"files_scanned"`
}

type node struct {
	name  string
	attrs map[string]string
	kids  []*node
}

func parseTree(x []byte) (*node, error) {
	dec := xml.NewDecoder(bytes.NewReader(x))
	var root *node
	var stack []*node
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &node{name: t.Name.Local, attrs: map[string]string{}}
			for _, a := range t.Attr {
				if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
					continue
				}
				n.attrs[a.Name.Local] = a.Value
			}
			if len(stack) > 0 {
				p := stack[len(stack)-1]
				p.kids = append(p.kids, n)
			} else {
				root = n
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if root == nil {
		return nil, errors.New("empty manifest")
	}
	return root, nil
}

func (n *node) find(name string) []*node {
	var out []*node
	for _, k := range n.kids {
		if k.name == name {
			out = append(out, k)
		}
	}
	return out
}

func (n *node) first(name string) *node {
	if l := n.find(name); len(l) > 0 {
		return l[0]
	}
	return nil
}

func boolAttr(n *node, key string) *bool {
	if n == nil {
		return nil
	}
	v, ok := n.attrs[key]
	if !ok {
		return nil
	}
	b := v == "true" || v == "0xffffffff" || v == "-1" || v == "1"
	return &b
}

func intAttr(n *node, key string) int {
	if n == nil {
		return 0
	}
	v := n.attrs[key]
	i, _ := strconv.ParseInt(strings.TrimPrefix(v, "0x"), map[bool]int{true: 16, false: 10}[strings.HasPrefix(v, "0x")], 64)
	return int(i)
}

func qualify(pkg, name string) string {
	switch {
	case name == "":
		return ""
	case strings.HasPrefix(name, "."):
		return pkg + name
	case !strings.Contains(name, "."):
		return pkg + "." + name
	}
	return name
}

var highRiskPerms = map[string]bool{
	"MANAGE_EXTERNAL_STORAGE": true, "SYSTEM_ALERT_WINDOW": true, "REQUEST_INSTALL_PACKAGES": true, "READ_SMS": true,
	"RECEIVE_SMS": true, "SEND_SMS": true, "READ_CALL_LOG": true, "WRITE_CALL_LOG": true, "PROCESS_OUTGOING_CALLS": true,
	"QUERY_ALL_PACKAGES": true, "BIND_ACCESSIBILITY_SERVICE": true, "BIND_DEVICE_ADMIN": true, "WRITE_SETTINGS": true,
}

var sensitivePerms = map[string]bool{
	"CAMERA": true, "RECORD_AUDIO": true, "ACCESS_FINE_LOCATION": true, "ACCESS_COARSE_LOCATION": true, "ACCESS_BACKGROUND_LOCATION": true,
	"READ_CONTACTS": true, "WRITE_CONTACTS": true, "READ_CALENDAR": true, "READ_PHONE_STATE": true, "READ_MEDIA_IMAGES": true,
	"READ_MEDIA_VIDEO": true, "READ_EXTERNAL_STORAGE": true, "BODY_SENSORS": true, "CALL_PHONE": true, "GET_ACCOUNTS": true,
}

func manifestText(zr *zip.Reader) ([]byte, error) {
	for _, f := range zr.File {
		if f.Name != "AndroidManifest.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		b, err := io.ReadAll(io.LimitReader(rc, 8<<20))
		if err != nil {
			return nil, err
		}
		if bytes.HasPrefix(bytes.TrimSpace(b), []byte("<")) {
			return b, nil
		}
		x, err := androidbinary.NewXMLFile(bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("cannot decode binary AndroidManifest.xml: %w", err)
		}
		return io.ReadAll(x.Reader())
	}
	return nil, errors.New("AndroidManifest.xml not found: not an APK")
}

func AnalyzeAPK(data []byte) (*StaticReport, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a valid APK/zip: %w", err)
	}
	if len(zr.File) > maxEntries {
		return nil, errors.New("archive has too many entries")
	}
	text, err := manifestText(zr)
	if err != nil {
		return nil, err
	}
	app, c, err := analyzeManifest(text)
	if err != nil {
		return nil, err
	}
	app.SizeMB = float64(int(float64(len(data))/1024/1024*10+0.5)) / 10
	app.Files = len(zr.File)

	abis := map[string]bool{}
	for _, f := range zr.File {
		switch {
		case strings.HasPrefix(f.Name, "lib/"):
			if p := strings.Split(f.Name, "/"); len(p) > 2 {
				abis[p[1]] = true
			}
		case strings.HasPrefix(f.Name, "META-INF/") && (strings.HasSuffix(f.Name, ".RSA") || strings.HasSuffix(f.Name, ".DSA") || strings.HasSuffix(f.Name, ".EC")):
			app.Signed = true
		}
	}
	if bytes.Contains(data, []byte("APK Sig Block 42")) {
		app.Signed, app.SignedV2 = true, true
	}
	for a := range abis {
		app.ABIs = append(app.ABIs, a)
	}
	sort.Strings(app.ABIs)

	switch {
	case !app.Signed:
		c.add("MOB-AND-SIGN-001", High, "APK is not signed", "Android refuses to install unsigned packages; release builds must be signed.")
	case !app.SignedV2 && app.MinSDK >= 24:
		c.add("MOB-AND-SIGN-002", Low, "only the legacy v1 (JAR) signature was found", "Use APK Signature Scheme v2 or later.")
	}
	if abis["armeabi-v7a"] && !abis["arm64-v8a"] {
		c.add("MOB-AND-ABI-001", Medium, "native libraries lack a 64-bit (arm64-v8a) build", "Google Play requires 64-bit support for apps with native code.")
	}

	sc := scanArchive(zr, func(n string) bool {
		return strings.HasPrefix(n, "META-INF/") || strings.HasPrefix(n, "res/drawable") || strings.HasPrefix(n, "res/mipmap")
	})
	for _, f := range sc.findings {
		c.list = append(c.list, f)
	}
	if len(sc.cleartext) > 0 {
		shown := sc.cleartext
		if len(shown) > 10 {
			shown = shown[:10]
		}
		c.add("MOB-AND-NET-003", Low, fmt.Sprintf("%d plain-http host(s) referenced in code or assets", len(sc.cleartext)), strings.Join(shown, ", "))
	}
	return &StaticReport{Platform: "android", Android: app, Findings: c.sorted(), Scanned: sc.filesRead}, nil
}

func analyzeManifest(x []byte) (*AndroidApp, *collector, error) {
	root, err := parseTree(x)
	if err != nil {
		return nil, nil, fmt.Errorf("manifest is not valid XML: %w", err)
	}
	c := &collector{}
	app := &AndroidApp{Package: root.attrs["package"], VersionName: root.attrs["versionName"], VersionCode: intAttr(root, "versionCode")}
	if sdk := root.first("uses-sdk"); sdk != nil {
		app.MinSDK, app.TargetSDK = intAttr(sdk, "minSdkVersion"), intAttr(sdk, "targetSdkVersion")
	}
	if app.TargetSDK == 0 {
		app.TargetSDK = app.MinSDK
	}
	for _, k := range []string{"uses-permission", "uses-permission-sdk-23"} {
		for _, p := range root.find(k) {
			app.Permissions = append(app.Permissions, p.attrs["name"])
		}
	}
	sort.Strings(app.Permissions)

	a := root.first("application")
	if a == nil {
		return nil, nil, errors.New("manifest has no <application> element")
	}

	if b := boolAttr(a, "debuggable"); b != nil && *b {
		c.add("MOB-AND-DBG-001", High, "app is debuggable", "android:debuggable=true lets anyone attach a debugger and read app data. Never ship it.")
	}
	if b := boolAttr(root, "testOnly"); (b != nil && *b) || (boolAttr(a, "testOnly") != nil && *boolAttr(a, "testOnly")) {
		c.add("MOB-AND-DBG-002", High, "app is marked testOnly", "testOnly packages cannot be distributed through stores.")
	}
	if b := boolAttr(a, "allowBackup"); (b == nil || *b) && a.attrs["dataExtractionRules"] == "" && a.attrs["fullBackupContent"] == "" {
		c.add("MOB-AND-BKP-001", Medium, "app data can be extracted with adb backup", "Set android:allowBackup=false or define backup rules that exclude sensitive data.")
	}
	cleartext := boolAttr(a, "usesCleartextTraffic")
	switch {
	case cleartext != nil && *cleartext:
		c.add("MOB-AND-NET-001", Medium, "cleartext (HTTP) traffic is allowed", "android:usesCleartextTraffic=true")
	case cleartext == nil && app.TargetSDK > 0 && app.TargetSDK < 28:
		c.add("MOB-AND-NET-002", Medium, "cleartext (HTTP) traffic is allowed by default", fmt.Sprintf("targetSdk %d is below 28, where cleartext became blocked by default.", app.TargetSDK))
	}
	if a.attrs["networkSecurityConfig"] != "" {
		c.add("MOB-AND-NET-004", Info, "custom network security config in use", "Review it for trusted user CAs and cleartext exceptions.")
	}
	if app.TargetSDK > 0 && app.TargetSDK < 34 {
		c.add("MOB-AND-SDK-001", Medium, fmt.Sprintf("targetSdkVersion %d is older than Google Play currently accepts for new releases", app.TargetSDK), "Check the current Play requirement and raise the target level.")
	}
	if app.MinSDK > 0 && app.MinSDK < 23 {
		c.add("MOB-AND-SDK-002", Low, fmt.Sprintf("minSdkVersion %d supports very old Android versions", app.MinSDK), "Old versions lack runtime permissions and modern security defaults.")
	}

	var high, sens []string
	for _, p := range app.Permissions {
		short := p[strings.LastIndex(p, ".")+1:]
		switch {
		case highRiskPerms[short]:
			high = append(high, short)
		case sensitivePerms[short]:
			sens = append(sens, short)
		}
	}
	if len(high) > 0 {
		c.add("MOB-AND-PERM-001", Medium, "high-risk or Play-restricted permissions requested", strings.Join(high, ", "))
	}
	if len(sens) > 0 {
		c.add("MOB-AND-PERM-002", Low, "sensitive runtime permissions requested", strings.Join(sens, ", "))
	}

	for _, kind := range []string{"activity", "activity-alias", "service", "receiver", "provider"} {
		for _, n := range a.find(kind) {
			comp := Component{Kind: kind, Name: qualify(app.Package, n.attrs["name"]), Exported: boolAttr(n, "exported"), Permission: n.attrs["permission"]}
			isLauncher := false
			for _, f := range n.find("intent-filter") {
				comp.IntentFilters++
				var mainAct, launcherCat bool
				for _, x := range f.find("action") {
					mainAct = mainAct || x.attrs["name"] == "android.intent.action.MAIN"
				}
				for _, x := range f.find("category") {
					launcherCat = launcherCat || x.attrs["name"] == "android.intent.category.LAUNCHER"
				}
				isLauncher = isLauncher || (mainAct && launcherCat)
				for _, d := range f.find("data") {
					if s := d.attrs["scheme"]; s != "" {
						comp.Schemes = append(comp.Schemes, s)
					}
				}
			}
			if isLauncher && app.Launcher == "" {
				app.Launcher = comp.Name
			}
			app.Components = append(app.Components, comp)

			exported := (comp.Exported != nil && *comp.Exported) || (comp.Exported == nil && comp.IntentFilters > 0 && app.TargetSDK < 31)
			if !exported || isLauncher || comp.Permission != "" {
				continue
			}
			switch kind {
			case "provider":
				c.add("MOB-AND-EXP-001", High, "content provider is exported without a permission", comp.Name)
			case "service", "receiver":
				c.add("MOB-AND-EXP-002", Medium, kind+" is exported without a permission", comp.Name+" can be invoked by any app.")
			}
		}
	}
	return app, c, nil
}
