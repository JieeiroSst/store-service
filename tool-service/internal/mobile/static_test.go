package mobile

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"howett.net/plist"
)

func ids(fs []Finding) map[string]string {
	m := map[string]string{}
	for _, f := range fs {
		m[f.ID] = f.Severity
	}
	return m
}

func TestRealAPK(t *testing.T) {
	data, err := os.ReadFile("testdata/helloworld.apk")
	if err != nil {
		t.Fatal(err)
	}
	rep, err := AnalyzeAPK(data)
	if err != nil {
		t.Fatal(err)
	}
	a := rep.Android
	if a.Package != "com.example.helloworld" || a.MinSDK == 0 || a.Launcher == "" || a.Files < 100 {
		t.Fatalf("manifest not decoded: %+v", a)
	}
	t.Logf("package=%s min=%d target=%d launcher=%s findings=%v", a.Package, a.MinSDK, a.TargetSDK, a.Launcher, ids(rep.Findings))
}

func TestRealLargeAPKIfPresent(t *testing.T) {
	p := os.Getenv("TEST_LARGE_APK")
	if p == "" {
		t.Skip("TEST_LARGE_APK not set")
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := AnalyzeAPK(data)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Android.Package == "" || len(rep.Android.Permissions) == 0 || rep.Android.Launcher == "" {
		t.Fatalf("%+v", rep.Android)
	}
	t.Logf("%s v%s target=%d perms=%d components=%d files_scanned=%d", rep.Android.Package, rep.Android.VersionName, rep.Android.TargetSDK, len(rep.Android.Permissions), len(rep.Android.Components), rep.Scanned)
	for _, f := range rep.Findings {
		t.Logf("  [%s] %s %s | %s", f.Severity, f.ID, f.Title, f.Detail)
	}
}

const riskyManifest = `<?xml version="1.0"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.acme.pay" android:versionCode="7" android:versionName="1.2">
<uses-sdk android:minSdkVersion="21" android:targetSdkVersion="28"/>
<uses-permission android:name="android.permission.INTERNET"/>
<uses-permission android:name="android.permission.SYSTEM_ALERT_WINDOW"/>
<uses-permission android:name="android.permission.READ_SMS"/>
<uses-permission android:name="android.permission.CAMERA"/>
<application android:debuggable="true" android:usesCleartextTraffic="true">
 <activity android:name=".Main"><intent-filter><action android:name="android.intent.action.MAIN"/><category android:name="android.intent.category.LAUNCHER"/></intent-filter></activity>
 <activity android:name=".Deep" android:exported="true"><intent-filter><action android:name="android.intent.action.VIEW"/><data android:scheme="acme"/></intent-filter></activity>
 <service android:name="com.acme.pay.SyncService" android:exported="true"/>
 <receiver android:name=".BootReceiver"><intent-filter><action android:name="android.intent.action.BOOT_COMPLETED"/></intent-filter></receiver>
 <provider android:name=".Db" android:authorities="com.acme.db" android:exported="true"/>
 <service android:name=".Safe" android:exported="true" android:permission="com.acme.PRIVATE"/>
</application></manifest>`

const cleanManifest = `<?xml version="1.0"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.acme.ok">
<uses-sdk android:minSdkVersion="26" android:targetSdkVersion="35"/>
<application android:allowBackup="false" android:usesCleartextTraffic="false">
 <activity android:name=".Main" android:exported="true"><intent-filter><action android:name="android.intent.action.MAIN"/><category android:name="android.intent.category.LAUNCHER"/></intent-filter></activity>
 <service android:name=".Internal" android:exported="false"/>
</application></manifest>`

func TestManifestRules(t *testing.T) {
	app, c, err := analyzeManifest([]byte(riskyManifest))
	if err != nil {
		t.Fatal(err)
	}
	got := ids(c.sorted())
	for id, sev := range map[string]string{
		"MOB-AND-DBG-001": High, "MOB-AND-BKP-001": Medium, "MOB-AND-NET-001": Medium, "MOB-AND-SDK-001": Medium,
		"MOB-AND-SDK-002": Low, "MOB-AND-PERM-001": Medium, "MOB-AND-PERM-002": Low, "MOB-AND-EXP-001": High, "MOB-AND-EXP-002": Medium,
	} {
		if got[id] != sev {
			t.Errorf("%s: want %s got %q (%v)", id, sev, got[id], got)
		}
	}
	if app.Launcher != "com.acme.pay.Main" || app.Package != "com.acme.pay" || app.VersionCode != 7 || app.TargetSDK != 28 {
		t.Fatalf("%+v", app)
	}
	var svc, rcv, safe int
	for _, f := range c.list {
		if f.ID == "MOB-AND-EXP-002" {
			svc++
			if strings.Contains(f.Detail, "Safe") {
				safe++
			}
			if strings.Contains(f.Detail, "BootReceiver") {
				rcv++
			}
		}
	}
	if svc != 2 || safe != 0 || rcv != 1 {
		t.Errorf("exported components: %d findings, permission-protected flagged=%d, implicit-export receiver=%d", svc, safe, rcv)
	}
	if len(app.Components[1].Schemes) != 1 || app.Components[1].Schemes[0] != "acme" {
		t.Errorf("deep link scheme not captured: %+v", app.Components[1])
	}
}

func TestCleanManifestHasNoFindings(t *testing.T) {
	_, c, err := analyzeManifest([]byte(cleanManifest))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.list) != 0 {
		t.Fatalf("well-configured app flagged: %+v", c.list)
	}
}

func TestManifestErrors(t *testing.T) {
	if _, _, err := analyzeManifest([]byte("not xml <")); err == nil {
		t.Error("garbage must fail")
	}
	if _, _, err := analyzeManifest([]byte(`<manifest package="x"/>`)); err == nil {
		t.Error("missing application must fail")
	}
	if _, err := AnalyzeAPK([]byte("junk")); err == nil {
		t.Error("non-zip must fail")
	}
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	zw.Create("readme.txt")
	zw.Close()
	if _, err := AnalyzeAPK(b.Bytes()); err == nil || !strings.Contains(err.Error(), "AndroidManifest") {
		t.Errorf("zip without manifest: %v", err)
	}
}

func makeZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	for name, data := range files {
		w, _ := zw.Create(name)
		w.Write(data)
	}
	zw.Close()
	return b.Bytes()
}

func TestAPKFromZipWithSecretsAndSignature(t *testing.T) {
	awsKey := "AKIA" + "IOSFODNN7EXAMPLE"
	pem := "-----BEGIN RSA " + "PRIVATE KEY-----"
	data := makeZip(t, map[string][]byte{
		"AndroidManifest.xml":  []byte(cleanManifest),
		"assets/config.json":   []byte(`{"aws":"` + awsKey + `","backend":"http://api.legacy-shop.io/v1"}`),
		"assets/id_rsa.pem":    []byte(pem + "\nMIIE..."),
		"classes.dex":          []byte("...http://schemas.android.com/apk/res/android... http://localhost:8080/x"),
		"lib/armeabi-v7a/a.so": []byte("x"),
	})
	rep, err := AnalyzeAPK(data)
	if err != nil {
		t.Fatal(err)
	}
	got := ids(rep.Findings)
	for id, sev := range map[string]string{"MOB-SEC-AWS": High, "MOB-SEC-PRIVKEY": High, "MOB-AND-SIGN-001": High, "MOB-AND-ABI-001": Medium, "MOB-AND-NET-003": Low} {
		if got[id] != sev {
			t.Errorf("%s: want %s got %q (%v)", id, sev, got[id], got)
		}
	}
	for _, f := range rep.Findings {
		if strings.Contains(f.Detail, awsKey) {
			t.Errorf("secret must be masked in the report: %+v", f)
		}
		if f.ID == "MOB-AND-NET-003" && (strings.Contains(f.Detail, "schemas.android") || strings.Contains(f.Detail, "localhost")) {
			t.Errorf("benign hosts must be ignored: %s", f.Detail)
		}
	}
	signed := makeZip(t, map[string][]byte{"AndroidManifest.xml": []byte(cleanManifest), "META-INF/CERT.RSA": []byte("x")})
	rep, _ = AnalyzeAPK(signed)
	if !rep.Android.Signed || ids(rep.Findings)["MOB-AND-SIGN-001"] != "" {
		t.Errorf("signed apk misreported: %+v %v", rep.Android, ids(rep.Findings))
	}
}

func plistBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, err := plist.Marshal(v, plist.XMLFormat)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func ipa(t *testing.T, info map[string]any, binary string, extra map[string][]byte) []byte {
	files := map[string][]byte{
		"Payload/Shop.app/Info.plist": plistBytes(t, info),
		"Payload/Shop.app/Shop":       []byte("\xcf\xfa\xed\xfe" + binary),
	}
	for k, v := range extra {
		files["Payload/Shop.app/"+k] = v
	}
	return makeZip(t, files)
}

func TestIPARules(t *testing.T) {
	info := map[string]any{
		"CFBundleIdentifier": "com.acme.shop", "CFBundleExecutable": "Shop", "CFBundleName": "Shop",
		"CFBundleShortVersionString": "3.1", "MinimumOSVersion": "12.0", "UIFileSharingEnabled": true,
		"NSAppTransportSecurity":   map[string]any{"NSAllowsArbitraryLoads": true, "NSExceptionDomains": map[string]any{"old.acme.io": map[string]any{"NSExceptionAllowsInsecureHTTPLoads": true}}},
		"NSCameraUsageDescription": "scan barcodes",
		"CFBundleURLTypes":         []any{map[string]any{"CFBundleURLSchemes": []any{"acme"}}},
		"UIBackgroundModes":        []any{"fetch"},
	}
	prov := plistBytes(t, map[string]any{
		"ProvisionedDevices": []any{"abc"}, "ExpirationDate": time.Now().Add(-48 * time.Hour),
		"Entitlements": map[string]any{"get-task-allow": true, "aps-environment": "development"},
	})
	rep, err := AnalyzeIPA(ipa(t, info, "AVCaptureDevice CLLocationManager LAContext CNContactStore", map[string][]byte{
		"embedded.mobileprovision": append([]byte("\x30\x82garbage"), prov...),
		"config.json":              []byte(`{"stripe":"sk_live_` + strings.Repeat("a", 24) + `"}`),
	}))
	if err != nil {
		t.Fatal(err)
	}
	got := ids(rep.Findings)
	for id, sev := range map[string]string{
		"MOB-IOS-ATS-001": High, "MOB-IOS-ATS-002": Medium, "MOB-IOS-FS-001": Low, "MOB-IOS-OS-001": Low, "MOB-IOS-VER-001": Medium,
		"MOB-IOS-PRIV-001": High, "MOB-IOS-PRIV-002": Medium, "MOB-IOS-DBG-001": High, "MOB-IOS-PUSH-001": Medium, "MOB-IOS-PROV-001": High, "MOB-SEC-STRIPE": High,
	} {
		if got[id] != sev {
			t.Errorf("%s: want %s got %q (%v)", id, sev, got[id], got)
		}
	}
	missing := map[string]bool{}
	for _, f := range rep.Findings {
		if f.ID == "MOB-IOS-PRIV-001" {
			missing[strings.Fields(f.Title)[1]] = true
		}
	}
	if !missing["location"] || !missing["Face"] || !missing["contacts"] || missing["camera/microphone"] {
		t.Errorf("usage-description check: %v (camera has a description and must not be flagged)", missing)
	}
	if rep.IOS.BundleID != "com.acme.shop" || len(rep.IOS.URLSchemes) != 1 || rep.IOS.Provisioning != "development" {
		t.Errorf("%+v", rep.IOS)
	}
}

func TestCleanIPA(t *testing.T) {
	info := map[string]any{"CFBundleIdentifier": "com.acme.ok", "CFBundleExecutable": "Shop", "CFBundleVersion": "12", "MinimumOSVersion": "16.0"}
	rep, err := AnalyzeIPA(ipa(t, info, "harmless", map[string][]byte{"PrivacyInfo.xcprivacy": []byte("<plist/>")}))
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Findings) != 0 {
		t.Fatalf("clean app flagged: %+v", rep.Findings)
	}
	if _, err := AnalyzeIPA([]byte("nope")); err == nil {
		t.Error("garbage must fail")
	}
	if _, err := AnalyzeIPA(makeZip(t, map[string][]byte{"x.txt": nil})); err == nil {
		t.Error("zip without Payload must fail")
	}
}
