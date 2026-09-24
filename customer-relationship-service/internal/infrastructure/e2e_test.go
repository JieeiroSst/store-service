package infrastructure

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-pdf/fpdf"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// env is the real fx graph (use cases, workflows, router) on an in-memory
// SQLite database, with a test CA, OCSP responder and time-stamp authority.
type env struct {
	idp    *testsupport.IDP
	engine *gin.Engine
	expiry port.ContractExpiryUsecase
	ca     *testsupport.CA
	rev    *testsupport.Revocation
	tsa    *testsupport.TSA
}

const apiKey = "e2e-key"

func newEngine(t *testing.T) *gin.Engine { return newApp(t).engine }

func newApp(t *testing.T, opts ...fx.Option) *env {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(model.Models()...); err != nil {
		t.Fatal(err)
	}

	ca := testsupport.NewCA(t)
	rev := testsupport.NewRevocation(t, ca)
	tsa := testsupport.NewTSA(t, ca, rev.OCSPURL())
	keyPath, certPath := testsupport.WriteSeal(t, ca, rev.OCSPURL())
	cfg := &config.Config{Company: config.CompanyConfig{
		Name: "Công ty Cổ phần Giải pháp Số Việt", Place: "Hà Nội", SignKeyFile: keyPath, SignCertFile: certPath,
	}, Files: config.FilesConfig{MaxBytes: 64 << 10, ScanRequired: true}, Signature: config.SignatureConfig{
		TrustRootsFile: ca.WriteRoots(t),
		Revocation:     "hard",
		TSAURL:         tsa.URL(),
	}}

	idp := testsupport.NewIDP(t, "https://sso.test/realms/crm", "crm-api")
	cfg.Server.APIKey = apiKey
	cfg.Auth = config.AuthConfig{
		APIKeyRole: "admin", OIDCIssuer: idp.Issuer, OIDCJWKSURL: idp.JWKSURL(), OIDCAudience: idp.Audience,
		RolesClaim: "realm_access.roles", RolePrefix: "crm-",
	}

	e := &env{ca: ca, rev: rev, tsa: tsa, idp: idp}
	app := fx.New(append([]fx.Option{
		fx.NopLogger,
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() *gorm.DB { return db }),
		Core,
		fx.Populate(&e.engine, &e.expiry),
	}, opts...)...)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
	return e
}

type client struct {
	t *testing.T
	h http.Handler
	// bearer is the token requests carry; without one they use the API key
	// (an administrator).
	bearer string
}

func (c client) as(token string) client { c.bearer = token; return c }

// do sends the request with the client's credentials.
func (c client) do(req *http.Request) *httptest.ResponseRecorder {
	if c.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearer)
	} else {
		req.Header.Set("X-API-Key", apiKey)
	}
	w := httptest.NewRecorder()
	c.h.ServeHTTP(w, req)
	return w
}

func (c client) call(method, path, body string, wantStatus int) map[string]any {
	c.t.Helper()
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := c.do(req)
	if w.Code != wantStatus {
		c.t.Fatalf("%s %s: status %d, want %d, body %s", method, path, w.Code, wantStatus, w.Body)
	}
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out
}

func (c client) raw(method, path, body string, wantStatus int, wantType string) []byte {
	c.t.Helper()
	req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
	w := c.do(req)
	if w.Code != wantStatus {
		c.t.Fatalf("%s %s: status %d, want %d, body %.200s", method, path, w.Code, wantStatus, w.Body)
	}
	if wantType != "" && w.Header().Get("Content-Type") != wantType {
		c.t.Fatalf("%s %s: content type %q, want %q", method, path, w.Header().Get("Content-Type"), wantType)
	}
	return w.Body.Bytes()
}

func (c client) list(path string) (rows []map[string]any, total string) {
	c.t.Helper()
	req := httptest.NewRequest("GET", "/api/v1"+path, nil)
	w := c.do(req)
	if w.Code != http.StatusOK {
		c.t.Fatalf("GET %s: status %d, body %s", path, w.Code, w.Body)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &rows)
	return rows, w.Header().Get("X-Total-Count")
}

func TestLeadConversion(t *testing.T) {
	c := client{t: t, h: newEngine(t)}

	// A client cannot create a lead that is already converted.
	lead := c.call("POST", "/leads", `{"lead_firstname":"Ada","lead_surname":"Lovelace","email":"ada@x.io","status":"converted"}`, 201)
	if lead["status"] != "new" {
		t.Fatalf("status = %v, want new", lead["status"])
	}

	res := c.call("POST", "/leads/1/convert", `{"create_opportunity":true,"opportunity_amount":1500}`, 200)
	account, contact := res["account"].(map[string]any), res["contact"].(map[string]any)
	if account["account_name"] != "Ada Lovelace" || contact["email"] != "ada@x.io" || res["opportunity"] == nil {
		t.Fatalf("conversion result = %v", res)
	}

	got := c.call("GET", "/leads/1", "", 200)
	if got["status"] != "converted" || got["converted_account_id"] != account["id"] {
		t.Fatalf("lead after conversion = %v", got)
	}

	// A PUT cannot un-convert it.
	got = c.call("PUT", "/leads/1", `{"lead_firstname":"Ada","lead_surname":"Byron","status":"new"}`, 200)
	if got["status"] != "converted" || got["lead_surname"] != "Byron" {
		t.Fatalf("lead after PUT = %v", got)
	}

	c.call("POST", "/leads/1/convert", "", 409)
	c.call("POST", "/leads/99/convert", "", 404)

	overview := c.call("GET", "/accounts/1/overview", "", 200)
	if len(overview["contacts"].([]any)) != 1 || len(overview["opportunities"].([]any)) != 1 {
		t.Fatalf("overview = %v", overview)
	}
}

func TestOpportunityPipelineAndSummary(t *testing.T) {
	c := client{t: t, h: newEngine(t)}
	c.call("POST", "/accounts", `{"account_name":"acme"}`, 201)
	opp := c.call("POST", "/opportunities", `{"account_id":1,"amount":100,"opportunity_stage":"won"}`, 201)
	if opp["opportunity_stage"] != "prospecting" {
		t.Fatalf("stage = %v, want prospecting", opp["opportunity_stage"])
	}

	c.call("POST", "/opportunities/1/stage", `{"stage":"nonsense"}`, 400)
	c.call("POST", "/opportunities/1/stage", `{"stage":"proposal"}`, 200)
	c.call("POST", "/opportunities/1/stage", `{"stage":"won"}`, 200)
	c.call("POST", "/opportunities/1/stage", `{"stage":"lost"}`, 409) // final

	summary := c.call("GET", "/reports/summary", "", 200)
	pipeline := summary["pipeline"].([]any)
	if len(pipeline) != len(model.Stages) {
		t.Fatalf("pipeline = %v", pipeline)
	}
	won := pipeline[4].(map[string]any)
	if won["stage"] != "won" || won["count"] != float64(1) || won["amount"] != float64(100) {
		t.Fatalf("won row = %v", won)
	}
}

func TestContractWorkflow(t *testing.T) {
	c := client{t: t, h: newEngine(t)}
	c.call("POST", "/contracts", `{"account_id":1,"contract_status":"approved"}`, 201)

	c.call("POST", "/contracts/1/approve", "", 409) // still a draft
	c.call("POST", "/contracts/1/submit", "", 200)
	c.call("POST", "/contracts/1/submit", "", 409)
	got := c.call("POST", "/contracts/1/approve", "", 200)
	if got["contract_status"] != "approved" || got["contract_approval"] != "approved" {
		t.Fatalf("contract = %v", got)
	}
	c.call("POST", "/contracts/1/reject", "", 409)
}

func TestCaseCloseAndListFilters(t *testing.T) {
	c := client{t: t, h: newEngine(t)}
	c.call("POST", "/cases", `{"contact_id":1,"subject":"Login broken"}`, 201)
	c.call("POST", "/cases", `{"contact_id":2,"subject":"Refund"}`, 201)

	if closed := c.call("POST", "/cases/1/close", "", 200); closed["status"] != "closed" || closed["closed_at"] == nil {
		t.Fatalf("case = %v", closed)
	}
	c.call("POST", "/cases/1/close", "", 409)

	rows, total := c.list("/cases?status=open")
	if len(rows) != 1 || total != "1" || rows[0]["subject"] != "Refund" {
		t.Fatalf("open cases = %v (total %s)", rows, total)
	}
	rows, _ = c.list("/cases?q=login")
	if len(rows) != 1 || rows[0]["subject"] != "Login broken" {
		t.Fatalf("search = %v", rows)
	}
	if s := c.call("GET", "/reports/summary", "", 200); s["open_cases"] != float64(1) {
		t.Fatalf("open_cases = %v", s["open_cases"])
	}
}

// expiredContract creates contract 1, approves it, lets its end date pass and
// runs the expiry job. It returns RFC 3339 timestamps for a future end date.
func expiredContract(t *testing.T, e *env) (c client, future string) {
	t.Helper()
	c = client{t: t, h: e.engine}
	c.call("POST", "/accounts", `{"account_name":"Công ty TNHH Acme","billing_address":"1 Lê Lợi, Hà Nội","tax_code":"0101234567","representative_title":"Giám đốc"}`, 201)
	future = time.Now().Add(48 * time.Hour).UTC().Format(time.RFC3339)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)

	c.call("POST", "/contracts", `{"account_id":1,"end_date":"`+future+`"}`, 201)
	c.call("POST", "/contracts/1/submit", "", 200)
	c.call("POST", "/contracts/1/approve", "", 200)
	if n, err := e.expiry.ExpireOverdue(context.Background()); err != nil || n != 0 {
		t.Fatalf("expired %d, err %v; want 0", n, err)
	}

	c.call("PUT", "/contracts/1", `{"account_id":1,"end_date":"`+past+`"}`, 200)
	if n, err := e.expiry.ExpireOverdue(context.Background()); err != nil || n != 1 {
		t.Fatalf("expired %d, err %v; want 1", n, err)
	}
	if got := c.call("GET", "/contracts/1", "", 200); got["contract_status"] != "expired" {
		t.Fatalf("status = %v, want expired", got["contract_status"])
	}
	return c, future
}

func TestContractExpiryAndResign(t *testing.T) {
	e := newApp(t)
	c, future := expiredContract(t, e)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)

	c.call("POST", "/contracts/1/submit", "", 409)
	c.call("POST", "/contracts/1/approve", "", 409)
	c.call("PUT", "/contracts/1", `{"account_id":1,"end_date":"`+future+`","contract_status":"approved"}`, 200)
	if got := c.call("GET", "/contracts/1", "", 200); got["contract_status"] != "expired" {
		t.Fatalf("PUT reopened the contract: %v", got["contract_status"])
	}

	key := testsupport.ECKey(t)
	cert := testsupport.Issue(t, e.ca, key, testsupport.Options{OCSPURL: e.rev.OCSPURL()})
	signBody := func(payload map[string]any, signature string) string {
		b, _ := json.Marshal(map[string]any{
			"signed_by": "Ada", "end_date": future, "signing_time": payload["signing_time"],
			"algorithm": "ECDSA-SHA256", "signature": signature, "certificate": cert,
		})
		return string(b)
	}

	payload := c.call("GET", "/contracts/1/signing-payload?signed_by=Ada&end_date="+future, "", 200)
	good := testsupport.Sign(t, key, "ECDSA-SHA256", []byte(payload["payload"].(string)))

	c.call("POST", "/contracts/1/sign", `{"signed_by":"Ada","end_date":"`+future+`"}`, 400)
	c.call("POST", "/contracts/1/sign", signBody(payload, testsupport.Sign(t, key, "ECDSA-SHA256", []byte("something else"))), 422)
	c.call("POST", "/contracts/1/sign", signBody(payload, testsupport.Sign(t, testsupport.ECKey(t), "ECDSA-SHA256", []byte(payload["payload"].(string)))), 422)
	_ = past

	// No time-stamp, no renewal: the authority being down is a 502, not a silent downgrade.
	e.tsa.SetDown(true)
	c.call("POST", "/contracts/1/sign", signBody(payload, good), 502)
	e.tsa.SetDown(false)
	if got := c.call("GET", "/contracts/1", "", 200); got["contract_status"] != "expired" {
		t.Fatalf("failed signatures reopened the contract: %v", got["contract_status"])
	}

	got := c.call("POST", "/contracts/1/sign", signBody(payload, good), 200)
	if got["contract_status"] != "approved" || got["signed_by"] != "Ada" || got["signed_at"] == nil ||
		got["signature_algorithm"] != "ECDSA-SHA256" || got["signature"] != good || got["signature_format"] != "detached-x509" ||
		got["revocation_status"] != "good" || got["timestamp_token"] == nil || got["timestamp_at"] == nil || got["timestamp_authority"] == nil ||
		!strings.Contains(got["signer_subject"].(string), "Ada Lovelace") || len(got["signer_fingerprint"].(string)) != 64 {
		t.Fatalf("signed contract = %v", got)
	}

	got = c.call("PUT", "/contracts/1", `{"account_id":1,"end_date":"`+future+`"}`, 200)
	if got["signature"] != good || got["timestamp_token"] == nil {
		t.Fatalf("PUT dropped the signature evidence: %v", got)
	}
	c.call("POST", "/contracts/1/sign", signBody(payload, good), 409)
}

func TestContractResignWithRevokedCertificate(t *testing.T) {
	e := newApp(t)
	c, future := expiredContract(t, e)

	key := testsupport.ECKey(t)
	cert := testsupport.Issue(t, e.ca, key, testsupport.Options{OCSPURL: e.rev.OCSPURL()})
	e.rev.Revoke(testsupport.ParseCert(t, cert).SerialNumber)

	payload := c.call("GET", "/contracts/1/signing-payload?signed_by=Ada&end_date="+future, "", 200)
	body, _ := json.Marshal(map[string]any{
		"signed_by": "Ada", "end_date": future, "signing_time": payload["signing_time"], "algorithm": "ECDSA-SHA256",
		"signature": testsupport.Sign(t, key, "ECDSA-SHA256", []byte(payload["payload"].(string))), "certificate": cert,
	})
	if got := c.call("POST", "/contracts/1/sign", string(body), 422); !strings.Contains(got["error"].(string), "revoked") {
		t.Fatalf("error = %v", got["error"])
	}
}

func TestContractResignWithPAdES(t *testing.T) {
	t.Run("stored in the database", func(t *testing.T) { resignWithPAdES(t, nil) })
	t.Run("stored in MinIO", func(t *testing.T) { resignWithPAdES(t, testsupport.NewMemStore()) })
}

// resignWithPAdES runs the PAdES renewal; a non-nil store replaces MinIO.
func resignWithPAdES(t *testing.T, store *testsupport.MemStore) {
	var opts []fx.Option
	if store != nil {
		opts = append(opts, fx.Decorate(func(port.DocumentStore) port.DocumentStore { return store }))
	}
	e := newApp(t, opts...)
	c, future := expiredContract(t, e)

	key := testsupport.ECKey(t)
	cert := testsupport.ParseCert(t, testsupport.Issue(t, e.ca, key, testsupport.Options{OCSPURL: e.rev.OCSPURL()}))
	signingTime := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)

	// The document to sign.
	q := "?signed_by=Ada&end_date=" + future + "&signing_time=" + signingTime
	issued := c.raw("GET", "/contracts/1/document"+q, "", 200, "application/pdf")
	if !bytes.HasPrefix(issued, []byte("%PDF")) {
		t.Fatalf("not a PDF: %.20q", issued)
	}
	c.call("GET", "/contracts/1/signed-document", "", 404)

	body := func(pdf []byte) string {
		b, _ := json.Marshal(map[string]any{
			"signed_by": "Ada", "end_date": future, "signing_time": signingTime, "signed_pdf": base64.StdEncoding.EncodeToString(pdf),
		})
		return string(b)
	}

	// Unsigned, or signed for someone else, is refused.
	c.call("POST", "/contracts/1/sign", body(issued), 422)
	other := c.raw("GET", "/contracts/1/document?signed_by=Eve&end_date="+future+"&signing_time="+signingTime, "", 200, "application/pdf")
	c.call("POST", "/contracts/1/sign", body(testsupport.SignPDF(t, other, key, cert, []*x509.Certificate{e.ca.Cert}, e.tsa.URL())), 422)

	signed := testsupport.SignPDF(t, issued, key, cert, []*x509.Certificate{e.ca.Cert}, e.tsa.URL())
	got := c.call("POST", "/contracts/1/sign", body(signed), 200)
	if got["contract_status"] != "approved" || got["signature_format"] != "pades" || got["revocation_status"] != "good" ||
		got["timestamp_token"] == nil || got["signed_by"] != "Ada" {
		t.Fatalf("signed contract = %v", got)
	}
	// Bên A countersigned after Bên B: its evidence is recorded next to Bên B's.
	if !strings.Contains(got["countersigner_subject"].(string), "Giải pháp Số Việt") ||
		len(got["countersigner_fingerprint"].(string)) != 64 || got["countersigned_at"] == nil ||
		got["countersigner_fingerprint"] == got["signer_fingerprint"] {
		t.Fatalf("countersignature evidence = %v", got)
	}

	// The kept file is Bên B's signed PDF with Bên A's signature appended.
	stored := c.raw("GET", "/contracts/1/signed-document", "", 200, "application/pdf")
	if !bytes.HasPrefix(stored, signed) || bytes.Count(stored, []byte("/ByteRange")) != 2 {
		t.Fatalf("stored document: %d bytes, %d signatures, extends Bên B's file: %v",
			len(stored), bytes.Count(stored, []byte("/ByteRange")), bytes.HasPrefix(stored, signed))
	}

	if store == nil {
		return
	}
	// The file lives in the bucket, not the database, and tampering is caught.
	if len(store.Objects) != 1 {
		t.Fatalf("objects in the bucket: %d", len(store.Objects))
	}
	for key := range store.Objects {
		if !strings.HasPrefix(key, "contracts/1/renewal/") || store.Types[key] != "application/pdf" {
			t.Fatalf("object %q (%s)", key, store.Types[key])
		}
		store.Tamper(key, append(append([]byte{}, stored...), '!'))
	}
	c.raw("GET", "/contracts/1/signed-document", "", 500, "")
}

// upload sends a multipart form to a contract's file endpoint.
func (c client) upload(contractID int, fields map[string]string, filename string, data []byte, wantStatus int) map[string]any {
	c.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	if filename != "" {
		part, _ := mw.CreateFormFile("file", filename)
		part.Write(data)
	}
	mw.Close()

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/contracts/%d/files", contractID), &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := c.do(req)
	if w.Code != wantStatus {
		c.t.Fatalf("upload: status %d, want %d, body %.300s", w.Code, wantStatus, w.Body)
	}
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return out
}

func (c client) download(path string, wantStatus int) *httptest.ResponseRecorder {
	c.t.Helper()
	req := httptest.NewRequest("GET", "/api/v1"+path, nil)
	w := c.do(req)
	if w.Code != wantStatus {
		c.t.Fatalf("GET %s: status %d, want %d", path, w.Code, wantStatus)
	}
	return w
}

// people are the users the file tests act as.
type people struct{ alice, bob, vera, max, nia string }

func newPeople(t *testing.T, e *env) people {
	return people{
		alice: e.idp.Token(t, "u-alice", "alice", "crm-staff", "offline_access"),
		bob:   e.idp.Token(t, "u-bob", "bob", "crm-staff"),
		vera:  e.idp.Token(t, "u-vera", "vera", "crm-viewer"),
		max:   e.idp.Token(t, "u-max", "max", "crm-manager"),
		nia:   e.idp.Token(t, "u-nia", "nia", "unrelated-role"),
	}
}

func fileMeta(t *testing.T, rows []map[string]any, name string) map[string]any {
	t.Helper()
	for _, r := range rows {
		if r["name"] == name {
			return r
		}
	}
	t.Fatalf("no file named %q in %v", name, rows)
	return nil
}

func TestContractFiles(t *testing.T) {
	e := newApp(t)
	admin := client{t: t, h: e.engine}
	who := newPeople(t, e)
	alice, bob, vera, max, nia := admin.as(who.alice), admin.as(who.bob), admin.as(who.vera), admin.as(who.max), admin.as(who.nia)

	admin.call("POST", "/accounts", `{"account_name":"Acme"}`, 201)
	admin.call("POST", "/contracts", `{"account_id":1,"contract_number":"HĐ-01/2026"}`, 201)
	admin.call("POST", "/contracts", `{"account_id":1}`, 201)
	pdf := validPDF(t)

	// The API needs credentials at all.
	anon := httptest.NewRecorder()
	e.engine.ServeHTTP(anon, httptest.NewRequest("GET", "/api/v1/contracts/1/files", nil))
	if anon.Code != 401 {
		t.Fatalf("no credentials: %d", anon.Code)
	}

	// The kinds a client can choose from.
	var kinds []map[string]any
	_ = json.Unmarshal(admin.raw("GET", "/contract-file-kinds", "", 200, ""), &kinds)
	if len(kinds) != len(model.FileKinds) || kinds[0]["kind"] != "contract" || kinds[2]["system"] != true {
		t.Fatalf("kinds = %v", kinds)
	}

	// Upload as a staff member: the uploader is who the token says, not what the form claims.
	up := alice.upload(1, map[string]string{"kind": "contract", "description": "Bản chính", "uploaded_by": "someone else"}, "Hợp đồng số 01.pdf", pdf, 201)
	if up["name"] != "Hợp đồng số 01.pdf" || up["kind"] != "contract" || up["content_type"] != "application/pdf" || up["size"] != float64(len(pdf)) ||
		up["source"] != "upload" || up["object_key"] != nil || len(up["sha256"].(string)) != 64 ||
		up["uploaded_by"] != "alice" || up["uploader_id"] != "u-alice" || up["version"] != float64(1) || up["latest"] != true || up["scan_status"] != "skipped" {
		t.Fatalf("uploaded = %v", up)
	}
	alice.upload(1, map[string]string{"kind": "scan"}, "scan.png", []byte("\x89PNG\r\n\x1a\n"), 415) // not a real PNG
	alice.upload(1, map[string]string{"kind": "appendix"}, "pl.pdf", pdf, 201)
	alice.upload(2, map[string]string{"kind": "contract"}, "other.pdf", pdf, 201)
	max.upload(1, map[string]string{"kind": "legal"}, "dkkd.pdf", pdf, 201)

	// A viewer reads; a staff member sees the same; the legal file is hidden below manager.
	for name, c := range map[string]client{"viewer": vera, "staff": alice} {
		rows, total := c.list("/contracts/1/files")
		if len(rows) != 2 || total != "2" {
			t.Fatalf("%s sees %v (total %s)", name, rows, total)
		}
	}
	if rows, total := max.list("/contracts/1/files"); len(rows) != 3 || total != "3" {
		t.Fatalf("manager sees %v", rows)
	}
	if rows, _ := alice.list("/contracts/1/files?kind=appendix"); len(rows) != 1 || rows[0]["name"] != "pl.pdf" {
		t.Fatalf("appendices = %v", rows)
	}
	alice.call("GET", "/contracts/1/files?kind=nonsense", "", 400)
	alice.call("GET", "/contracts/99/files", "", 404)
	nia.call("GET", "/contracts/1/files", "", 403) // authenticated, but no role of ours

	legalID := int(fileMeta(t, mustList(t, max, "/contracts/1/files"), "dkkd.pdf")["id"].(float64))
	vera.call("GET", fmt.Sprintf("/contracts/1/files/%d", legalID), "", 404)
	alice.call("GET", fmt.Sprintf("/contracts/1/files/%d/download", legalID), "", 404)
	max.call("GET", fmt.Sprintf("/contracts/1/files/%d", legalID), "", 200)

	// Download.
	dl := vera.download("/contracts/1/files/1/download", 200)
	if !bytes.Equal(dl.Body.Bytes(), pdf) || dl.Header().Get("Content-Type") != "application/pdf" ||
		dl.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.HasPrefix(dl.Header().Get("Content-Disposition"), "attachment;") ||
		!strings.Contains(dl.Header().Get("Content-Disposition"), "filename*=") {
		t.Fatalf("download headers: %v", dl.Header())
	}
	admin.download("/contracts/2/files/1/download", 404)

	// What is refused.
	vera.upload(1, map[string]string{"kind": "contract"}, "v.pdf", pdf, 403)
	alice.upload(1, map[string]string{"kind": "legal"}, "l.pdf", pdf, 403)
	alice.upload(1, map[string]string{"kind": "renewal"}, "r.pdf", pdf, 403)
	alice.upload(1, map[string]string{"kind": "contract"}, "virus.pdf", []byte("MZ\x90\x00 nope"), 415)
	alice.upload(1, map[string]string{"kind": "contract"}, "hd.exe", pdf, 415)
	alice.upload(1, map[string]string{"kind": "nonsense"}, "hd.pdf", pdf, 400)
	alice.upload(9, map[string]string{"kind": "contract"}, "hd.pdf", pdf, 404)
	alice.upload(1, map[string]string{"kind": "contract"}, "", nil, 400) // no file part
	alice.upload(1, map[string]string{"kind": "contract"}, "big.pdf", append(append([]byte{}, pdf...), make([]byte, 100<<10)...), 413)
	alice.upload(1, map[string]string{"kind": "contract"}, "huge.pdf", append(append([]byte{}, pdf...), make([]byte, 2<<20)...), 413) // past the request limit
	alice.upload(1, map[string]string{"kind": "contract"}, "js.pdf", pdfWithJavaScript(t), 422)                                       // active content

	// Deleting: only the uploader or a manager.
	bob.call("DELETE", "/contracts/1/files/2", "", 403)
	vera.call("DELETE", "/contracts/1/files/2", "", 403)
	alice.call("DELETE", "/contracts/1/files/2", "", 204)
	alice.call("GET", "/contracts/1/files/2", "", 404)
	if rows, total := alice.list("/contracts/1/files"); len(rows) != 1 || total != "1" {
		t.Fatalf("after delete: %v (total %s)", rows, total)
	}
	alice.call("DELETE", "/contracts/1/files/2", "", 404)
	max.call("DELETE", "/contracts/1/files/1", "", 204) // a manager deletes anyone's
}

func mustList(t *testing.T, c client, path string) []map[string]any {
	t.Helper()
	rows, _ := c.list(path)
	return rows
}

func TestContractFileVersionsAndHistory(t *testing.T) {
	e := newApp(t)
	admin := client{t: t, h: e.engine}
	who := newPeople(t, e)
	alice, bob, vera, max := admin.as(who.alice), admin.as(who.bob), admin.as(who.vera), admin.as(who.max)
	admin.call("POST", "/accounts", `{"account_name":"Acme"}`, 201)
	admin.call("POST", "/contracts", `{"account_id":1}`, 201)
	pdf := validPDF(t)

	v1 := alice.upload(1, map[string]string{"kind": "contract"}, "hd-v1.pdf", pdf, 201)
	id1 := int(v1["id"].(float64))

	// Someone else cannot replace it; the uploader can; so can a manager.
	bob.upload(1, map[string]string{"kind": "contract", "replaces": fmt.Sprint(id1)}, "hd-v2.pdf", pdf, 403)
	alice.upload(1, map[string]string{"kind": "contract", "replaces": "abc"}, "hd-v2.pdf", pdf, 400)
	alice.upload(1, map[string]string{"kind": "appendix", "replaces": fmt.Sprint(id1)}, "hd-v2.pdf", pdf, 400)
	v2 := alice.upload(1, map[string]string{"kind": "contract", "replaces": fmt.Sprint(id1)}, "hd-v2.pdf", pdf, 201)
	id2 := int(v2["id"].(float64))
	if v2["version"] != float64(2) || v2["lineage_id"] != float64(id1) || v2["replaces_id"] != float64(id1) || v2["latest"] != true {
		t.Fatalf("v2 = %v", v2)
	}
	alice.upload(1, map[string]string{"kind": "contract", "replaces": fmt.Sprint(id1)}, "again.pdf", pdf, 409) // v1 is no longer the latest
	v3 := max.upload(1, map[string]string{"kind": "contract", "replaces": fmt.Sprint(id2)}, "hd-v3.pdf", pdf, 201)
	if v3["version"] != float64(3) {
		t.Fatalf("v3 = %v", v3)
	}
	id3 := int(v3["id"].(float64))

	// The list shows the current version; ?all_versions=true shows the chain.
	rows, total := vera.list("/contracts/1/files")
	if total != "1" || rows[0]["name"] != "hd-v3.pdf" {
		t.Fatalf("current = %v", rows)
	}
	if rows, total = vera.list("/contracts/1/files?all_versions=true"); total != "3" || len(rows) != 3 {
		t.Fatalf("all versions = %v", rows)
	}
	vera.download(fmt.Sprintf("/contracts/1/files/%d/download", id1), 200) // an old version is still downloadable

	// Delete the newest: v2 is current again.
	max.call("DELETE", fmt.Sprintf("/contracts/1/files/%d", id3), "", 204)
	if rows, _ = vera.list("/contracts/1/files"); len(rows) != 1 || rows[0]["name"] != "hd-v2.pdf" || rows[0]["latest"] != true {
		t.Fatalf("after deleting v3: %v", rows)
	}

	// The history, asked through any version, includes the deleted one and who did what.
	h := vera.call("GET", fmt.Sprintf("/contracts/1/files/%d/history", id1), "", 200)
	versions, events := h["versions"].([]any), h["events"].([]any)
	if len(versions) != 3 || versions[2].(map[string]any)["deleted_at"] == nil {
		t.Fatalf("versions = %v", versions)
	}
	var actions []string
	for _, ev := range events {
		m := ev.(map[string]any)
		actions = append(actions, m["action"].(string)+":"+m["actor_name"].(string))
	}
	// uploaded(v1), bob's refused attempt to replace it, v2, v3, vera's download of v1, max deleting v3.
	want := "uploaded:alice,denied:bob,new_version:alice,new_version:max,downloaded:vera,deleted:max"
	if strings.Join(actions, ",") != want {
		t.Fatalf("events = %v\nwant   %s", actions, want)
	}
	if first := events[0].(map[string]any); first["actor_id"] != "u-alice" || first["actor_roles"] != "staff" || first["actor_service"] != false || first["ip"] == "" {
		t.Fatalf("first event = %v", first)
	}

	// The contract-wide audit trail: managers only, filterable.
	for name, c := range map[string]client{"staff": alice, "viewer": vera} {
		c.call("GET", "/contracts/1/file-events", "", 403)
		_ = name
	}
	all, total := max.list("/contracts/1/file-events")
	if total == "" || len(all) < 5 {
		t.Fatalf("audit trail = %v", all)
	}
	if rows, _ = max.list("/contracts/1/file-events?action=deleted"); len(rows) != 1 || rows[0]["actor_name"] != "max" {
		t.Fatalf("deleted entries = %v", rows)
	}
	if rows, _ = max.list(fmt.Sprintf("/contracts/1/file-events?file_id=%d", id2)); len(rows) < 1 {
		t.Fatalf("entries for file %d = %v", id2, rows)
	}
	max.call("GET", "/contracts/1/file-events?file_id=x", "", 400)
	// Refused attempts are on it too (bob's replace above).
	if rows, _ = max.list("/contracts/1/file-events?action=denied"); len(rows) < 1 {
		t.Fatal("denied attempts were not audited")
	}
}

func TestContractFileScanning(t *testing.T) {
	scanner := &e2eScanner{res: port.ScanResult{Clean: true, Engine: "clamav"}}
	e := newApp(t, fx.Decorate(func(port.Scanner) port.Scanner { return scanner }))
	admin := client{t: t, h: e.engine}
	alice := admin.as(newPeople(t, e).alice)
	admin.call("POST", "/accounts", `{"account_name":"Acme"}`, 201)
	admin.call("POST", "/contracts", `{"account_id":1}`, 201)
	pdf := validPDF(t)

	ok := alice.upload(1, map[string]string{"kind": "contract"}, "ok.pdf", pdf, 201)
	if ok["scan_status"] != "clean" || ok["scan_engine"] != "clamav" || ok["scanned_at"] == nil {
		t.Fatalf("clean upload = %v", ok)
	}

	scanner.res = port.ScanResult{Signature: "Win.Test.EICAR_HDB-1", Engine: "clamav"}
	if got := alice.upload(1, map[string]string{"kind": "contract"}, "bad.pdf", pdf, 422); !strings.Contains(got["error"].(string), "EICAR") {
		t.Fatalf("infected upload: %v", got)
	}

	scanner.err = errors.New("connection refused")
	alice.upload(1, map[string]string{"kind": "contract"}, "down.pdf", pdf, 502) // a required scanner that is down
	if rows, total := alice.list("/contracts/1/files"); total != "1" || len(rows) != 1 {
		t.Fatalf("only the clean file should exist: %v", rows)
	}
}

type e2eScanner struct {
	res port.ScanResult
	err error
}

func (s *e2eScanner) Enabled() bool { return true }
func (s *e2eScanner) Scan(context.Context, string, []byte) (port.ScanResult, error) {
	return s.res, s.err
}

func validPDF(t *testing.T) []byte {
	t.Helper()
	return fpdfDocument(t, nil)
}

func pdfWithJavaScript(t *testing.T) []byte {
	t.Helper()
	return fpdfDocument(t, func(p *fpdf.Fpdf) { p.SetJavascript("app.alert('x');") })
}

func fpdfDocument(t *testing.T, mutate func(*fpdf.Fpdf)) []byte {
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

func TestSignedRenewalIsAProtectedContractFile(t *testing.T) {
	e := newApp(t)
	c, future := expiredContract(t, e)

	key := testsupport.ECKey(t)
	cert := testsupport.ParseCert(t, testsupport.Issue(t, e.ca, key, testsupport.Options{OCSPURL: e.rev.OCSPURL()}))
	signingTime := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	issued := c.raw("GET", "/contracts/1/document?signed_by=Ada&end_date="+future+"&signing_time="+signingTime, "", 200, "application/pdf")
	signed := testsupport.SignPDF(t, issued, key, cert, []*x509.Certificate{e.ca.Cert}, e.tsa.URL())
	b, _ := json.Marshal(map[string]any{"signed_by": "Ada", "end_date": future, "signing_time": signingTime, "signed_pdf": base64.StdEncoding.EncodeToString(signed)})
	c.call("POST", "/contracts/1/sign", string(b), 200)

	// It shows up among the contract's files, as a system file, next to what users upload.
	c.upload(1, map[string]string{"kind": "scan"}, "bản scan.pdf", validPDF(t), 201)
	rows, total := c.list("/contracts/1/files")
	if total != "2" {
		t.Fatalf("files = %v", rows)
	}
	renewals, _ := c.list("/contracts/1/files?kind=renewal")
	if len(renewals) != 1 || renewals[0]["source"] != "system" || renewals[0]["uploaded_by"] != "Ada" ||
		renewals[0]["content_type"] != "application/pdf" || !strings.HasPrefix(renewals[0]["name"].(string), "Phu-luc-gia-han-") {
		t.Fatalf("renewal file = %v", renewals)
	}

	// The same bytes through both endpoints; the evidence cannot be deleted.
	id := int(renewals[0]["id"].(float64))
	viaFiles := c.download(fmt.Sprintf("/contracts/1/files/%d/download", id), 200).Body.Bytes()
	viaSigned := c.raw("GET", "/contracts/1/signed-document", "", 200, "application/pdf")
	if !bytes.Equal(viaFiles, viaSigned) || !bytes.HasPrefix(viaFiles, signed) {
		t.Fatal("the renewal file and signed-document differ, or do not extend Bên B's signed file")
	}
	c.call("DELETE", fmt.Sprintf("/contracts/1/files/%d", id), "", 403)
}
