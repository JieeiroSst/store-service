package domain

import (
	"fmt"
	"strings"
	"time"
)

type LinkSafetyOutcome string

const (
	SafetyAllow LinkSafetyOutcome = "allow"
	SafetyWarn  LinkSafetyOutcome = "warn"
	SafetyBlock LinkSafetyOutcome = "block"
)

// LinkSafetyInput mirrors LinkSafetyInput in link-safety.ts.
type LinkSafetyInput struct {
	IsActive         *bool
	WarnAt           *time.Time
	OwnerSuspendedAt *time.Time
}

func EvaluateLinkSafety(in LinkSafetyInput) LinkSafetyOutcome {
	if in.OwnerSuspendedAt != nil {
		return SafetyBlock
	}
	if in.IsActive != nil && !*in.IsActive {
		return SafetyBlock
	}
	if in.WarnAt != nil {
		return SafetyWarn
	}
	return SafetyAllow
}

func SafeHref(destination string) string {
	d := strings.TrimSpace(destination)
	if !strings.HasPrefix(d, "http://") && !strings.HasPrefix(d, "https://") {
		return ""
	}
	return destination
}

func EscapeHTML(value string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(value)
}

func GenerateWarningLinkHTML(destination, reportURL string) string {
	shown := "(no destination recorded)"
	if strings.TrimSpace(destination) != "" {
		shown = EscapeHTML(destination)
	}

	continueButton := `<p class="inert">This link has no usable web destination, so there is nothing to continue to.</p>`
	if href := SafeHref(destination); href != "" {
		continueButton = fmt.Sprintf(
			`<a class="go" href="%s" rel="nofollow noopener noreferrer">Continue anyway</a>`,
			EscapeHTML(href),
		)
	}

	reportLink := ""
	if reportURL != "" {
		reportLink = fmt.Sprintf(
			`<p class="report"><a href="%s" rel="nofollow noopener">Report this link</a></p>`,
			EscapeHTML(reportURL),
		)
	}

	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex,nofollow">
<title>Check this link before continuing</title>
<style>
  :root { color-scheme: light dark; }
  body { margin:0; min-height:100vh; display:flex; align-items:center; justify-content:center;
         background:#f6f7f8; color:#16191d; padding:24px;
         font:16px/1.55 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif; }
  .card { max-width:34rem; width:100%; background:#fff; border:1px solid #e3e6ea;
          border-radius:10px; padding:28px; }
  h1 { margin:0 0 12px; font-size:1.35rem; line-height:1.25; }
  p { margin:0 0 14px; }
  .dest { display:block; word-break:break-all; background:#f0f2f4; border-radius:6px;
          padding:10px 12px; font-family:ui-monospace,SFMono-Regular,Menlo,monospace;
          font-size:.9rem; margin-bottom:18px; }
  a.go { display:inline-block; text-decoration:none; border:1px solid #c9ced4; color:#16191d;
         border-radius:6px; padding:9px 15px; font-size:.95rem; font-weight:600; }
  .inert { margin:0; font-size:.9rem; color:#5b636d; }
  .report { margin:16px 0 0; font-size:.85rem; }
  .report a { color:#5b636d; }
  @media (prefers-color-scheme: dark) {
    body { background:#14171a; color:#e8eaed; }
    .card { background:#1d2126; border-color:#2c3238; }
    .dest { background:#14171a; }
    a.go { border-color:#3a4149; color:#e8eaed; }
    .report a { color:#98a1ab; }
    .inert { color:#98a1ab; }
  }
</style>
</head>
<body>
  <main class="card">
    <h1>Check this link before continuing</h1>
    <p>This short link has been flagged as possibly unsafe, so we have not sent you
       straight there. It leads to:</p>
    <span class="dest">` + shown + `</span>
    <p>If you were not expecting this link, or it claims to be from a bank, a government
       service, or a company you do business with, close this page. Do not enter any
       password or personal details.</p>
    ` + continueButton + `
    ` + reportLink + `
  </main>
</body>
</html>`
}
