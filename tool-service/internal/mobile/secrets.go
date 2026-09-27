package mobile

import (
	"archive/zip"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	maxEntries      = 60000
	maxFileScan     = 64 << 20
	maxTotalScan    = 512 << 20
	maxSecretsShown = 15
)

type secretRule struct {
	id, title, sev string
	re             *regexp.Regexp
}

var secretRules = []secretRule{
	{"MOB-SEC-AWS", "AWS access key id embedded in the app", High, regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"MOB-SEC-PRIVKEY", "private key embedded in the app", High, regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`)},
	{"MOB-SEC-STRIPE", "Stripe live secret key embedded in the app", High, regexp.MustCompile(`sk_live_[0-9a-zA-Z]{24,}`)},
	{"MOB-SEC-SLACK", "Slack token embedded in the app", High, regexp.MustCompile(`xox[baprs]-[0-9A-Za-z-]{10,}`)},
	{"MOB-SEC-GITHUB", "GitHub token embedded in the app", High, regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`)},
	{"MOB-SEC-GOOGLE", "Google API key embedded (verify it is restricted to your app)", Medium, regexp.MustCompile(`AIza[0-9A-Za-z_\-]{35}`)},
	{"MOB-SEC-GENERIC", "hard-coded credential assignment", Medium, regexp.MustCompile(`(?i)(?:api[_-]?key|secret|passwd|password|access[_-]?token|auth[_-]?token)["']?\s*[:=]\s*["'][A-Za-z0-9/+_\-]{12,}["']`)},
}

var scannableExt = regexp.MustCompile(`(?i)(\.dex|\.json|\.xml|\.properties|\.plist|\.txt|\.js|\.html|\.yml|\.yaml|\.cfg|\.conf|\.env|\.strings|\.so|\.pem|\.key)$|^assets/|^res/raw/|(^|/)[^/.]+$`)

var benignURLHost = regexp.MustCompile(`(?i)(schemas\.android\.com|w3\.org|xmlpull\.org|apache\.org|localhost|127\.0\.0\.1|10\.0\.2\.2|example\.(com|org)|developer\.android\.com|ns\.adobe\.com|purl\.org|xml\.org|java\.sun\.com|www\.apple\.com/DTDs)`)
var cleartextURL = regexp.MustCompile(`http://[A-Za-z0-9.\-]+[A-Za-z0-9](?::\d+)?(?:/[\w\-./?%&=]*)?`)

type scanResult struct {
	findings   []Finding
	cleartext  []string
	filesRead  int
	bytesRead  int64
	skippedBig int
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + strings.Repeat("*", 6) + s[len(s)-2:]
}

func scanArchive(zr *zip.Reader, skipName func(string) bool) scanResult {
	var res scanResult
	type hit struct {
		rule         secretRule
		file, sample string
	}
	var hits []hit
	seenKey := map[string]bool{}
	hosts := map[string]bool{}

	for _, f := range zr.File {
		if f.FileInfo().IsDir() || skipName(f.Name) || !scannableExt.MatchString(f.Name) {
			continue
		}
		if f.UncompressedSize64 > maxFileScan {
			res.skippedBig++
			continue
		}
		if res.bytesRead+int64(f.UncompressedSize64) > maxTotalScan {
			break
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		b, err := io.ReadAll(io.LimitReader(rc, maxFileScan))
		rc.Close()
		if err != nil {
			continue
		}
		res.filesRead++
		res.bytesRead += int64(len(b))
		text := string(b)
		for _, r := range secretRules {
			for _, m := range r.re.FindAllString(text, 3) {
				if k := r.id + m; !seenKey[k] {
					seenKey[k] = true
					hits = append(hits, hit{r, f.Name, maskSecret(m)})
				}
			}
		}
		if strings.HasSuffix(f.Name, ".dex") || strings.HasPrefix(f.Name, "assets/") || strings.HasSuffix(f.Name, ".json") {
			for _, u := range cleartextURL.FindAllString(text, 200) {
				h := strings.TrimPrefix(u, "http://")
				h = strings.SplitN(h, "/", 2)[0]
				if !benignURLHost.MatchString(h) && strings.Contains(h, ".") {
					hosts[h] = true
				}
			}
		}
	}

	for _, h := range hits {
		if len(res.findings) >= maxSecretsShown {
			break
		}
		res.findings = append(res.findings, Finding{h.rule.id, h.rule.sev, h.rule.title, fmt.Sprintf("%s in %s", h.sample, h.file)})
	}
	for h := range hosts {
		res.cleartext = append(res.cleartext, h)
	}
	sort.Strings(res.cleartext)
	return res
}
