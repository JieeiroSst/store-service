package infrastructure

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/adapter/secondary/uploadservice/uploadservicetest"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

var pngBytes = append([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"), bytes.Repeat([]byte{7}, 2000)...)

func startWithUpload(t *testing.T) (testServer, *uploadservicetest.Fake) {
	t.Helper()
	fake := uploadservicetest.NewFake("svc-key")
	us := httptest.NewServer(fake)
	t.Cleanup(us.Close)
	s := startServer(t, chatConfig+fmt.Sprintf(`
upload_service:
  url: %s
  service_key: svc-key
modules:
  mod_http_upload:
    max_size: 1048576
    put_url: "http://127.0.0.1:@HTTP_PORT@/upload"
    external_secret: test-secret
`, us.URL))
	return s, fake
}

func httpDo(t *testing.T, method, url string, body []byte) (int, []byte, http.Header) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, out, resp.Header
}

func requestSlot(c *client, id, file string, size int, ctype string) *domain.Element {
	return c.iq(id, "get", "upload.localhost", fmt.Sprintf(
		"<request xmlns='urn:xmpp:http:upload:0' filename='%s' size='%d' content-type='%s'/>", file, size, ctype))
}

func TestSendFileThroughUploadService(t *testing.T) {
	s, fake := startWithUpload(t)
	provision(t, s, "alice", "bob")
	alice := login(t, s.c2s, "alice", "alicepw", "phone", "PLAIN")
	bob := login(t, s.c2s, "bob", "bobpw", "phone", "PLAIN")
	alice.send("<presence/>")
	bob.send("<presence/>")
	bob.expect("alice online", isPresence("alice@localhost/phone", "available"))

	t.Log("1. discover the upload service on the server")
	items := alice.iq("di", "get", "localhost", "<query xmlns='http://jabber.org/protocol/disco#items'/>")
	item := items.Child("query", domain.NSDiscoItems).Child("item", "")
	if item == nil || item.Attr("jid") != "upload.localhost" {
		t.Fatalf("disco#items: %s", items)
	}
	info := alice.iq("dinfo", "get", "upload.localhost", "<query xmlns='http://jabber.org/protocol/disco#info'/>")
	q := info.Child("query", domain.NSDiscoInfo)
	if q.Child("identity", "").Attr("category") != "store" || !strings.Contains(info.String(), "1048576") {
		t.Fatalf("disco#info: %s", info)
	}

	t.Log("2. request a slot")
	res := requestSlot(alice, "slot1", "ảnh chụp.png", len(pngBytes), "image/png")
	slot := res.Child("slot", domain.NSHTTPUpload)
	if slot == nil {
		t.Fatalf("slot: %s", res)
	}
	putURL := slot.Child("put", "").Attr("url")
	getURL := slot.Child("get", "").Attr("url")
	t.Logf("   put %s", putURL)
	t.Logf("   get %s", getURL)

	t.Log("3. tampered or wrong-size uploads are refused")
	if code, _, _ := httpDo(t, http.MethodPut, strings.Replace(putURL, "&v=", "&v=00", 1), pngBytes); code != http.StatusForbidden {
		t.Fatalf("tampered signature = %d", code)
	}
	if code, _, _ := httpDo(t, http.MethodPut, putURL, pngBytes[:100]); code != http.StatusRequestEntityTooLarge {
		t.Fatalf("size mismatch = %d", code)
	}
	if fake.Uploads.Load() != 0 {
		t.Fatal("rejected uploads must not reach upload_service")
	}

	t.Log("4. HTTP PUT stores the file in upload_service")
	if code, body, _ := httpDo(t, http.MethodPut, putURL, pngBytes); code != http.StatusCreated {
		t.Fatalf("PUT = %d %s", code, body)
	}
	files := fake.Files()
	if len(files) != 1 || !strings.HasPrefix(files[0].ReceiverID, "ejabberd:") || !bytes.Equal(files[0].Data, pngBytes) {
		t.Fatalf("upload_service content: %+v", files)
	}
	if code, _, _ := httpDo(t, http.MethodPut, putURL, pngBytes); code != http.StatusForbidden {
		t.Fatalf("slot reuse = %d", code)
	}

	t.Log("5. alice sends the link in a chat message, bob receives it in real time")
	alice.send(fmt.Sprintf("<message to='bob@localhost' type='chat' id='f1'><body>%s</body><x xmlns='jabber:x:oob'><url>%s</url></x></message>",
		domain.EscapeText(getURL), domain.EscapeText(getURL)))
	msg := bob.expect("file message", func(e *domain.Element) bool { return e.Name == "message" && e.Attr("id") == "f1" })
	link := msg.Child("x", domain.NSOOB).ChildText("url", "")
	if link != getURL {
		t.Fatalf("oob url = %q", link)
	}

	t.Log("6. bob downloads the file from the link")
	code, body, hdr := httpDo(t, http.MethodGet, link, nil)
	if code != http.StatusOK || !bytes.Equal(body, pngBytes) || hdr.Get("Content-Type") != "image/png" {
		t.Fatalf("GET = %d %s (%d bytes)", code, hdr.Get("Content-Type"), len(body))
	}
	if !strings.Contains(hdr.Get("Content-Disposition"), "inline") {
		t.Fatalf("images should be displayed inline: %s", hdr.Get("Content-Disposition"))
	}
	if code, _, _ := httpDo(t, http.MethodHead, link, nil); code != http.StatusOK {
		t.Fatalf("HEAD = %d", code)
	}
	if code, _, _ := httpDo(t, http.MethodGet, strings.Replace(link, "/upload/", "/upload/x", 1), nil); code != http.StatusNotFound {
		t.Fatalf("other user path = %d", code)
	}
	if code, _, _ := httpDo(t, http.MethodGet, strings.TrimSuffix(link, ".png")+".jpg", nil); code != http.StatusNotFound {
		t.Fatalf("other file name = %d", code)
	}

	t.Log("7. the link also works when bob was offline")
	bob.send("<presence type='unavailable'/></stream:stream>")
	alice.expect("bob offline", isPresence("bob@localhost/phone", "unavailable"))
	alice.send(fmt.Sprintf("<message to='bob@localhost' type='chat' id='f2'><body>%s</body><x xmlns='jabber:x:oob'><url>%s</url></x></message>",
		domain.EscapeText(getURL), domain.EscapeText(getURL)))
	time.Sleep(200 * time.Millisecond)
	bob2 := login(t, s.c2s, "bob", "bobpw", "phone", "PLAIN")
	bob2.send("<presence/>")
	off := bob2.expect("offline file message", func(e *domain.Element) bool { return e.Name == "message" && e.Attr("id") == "f2" })
	if off.Child("x", domain.NSOOB).ChildText("url", "") != getURL || off.Child("delay", domain.NSDelay) == nil {
		t.Fatalf("offline file message: %s", off)
	}
}

func TestUploadRejections(t *testing.T) {
	s, _ := startWithUpload(t)
	provision(t, s, "carol")
	carol := login(t, s.c2s, "carol", "carolpw", "phone", "PLAIN")

	t.Log("files above max_size get file-too-large with the limit")
	res := requestSlot(carol, "big", "movie.mp4", 5<<20, "video/mp4")
	tl := res.Child("error", "").Child("file-too-large", domain.NSHTTPUpload)
	if res.Attr("type") != "error" || res.Child("error", "").Child("not-acceptable", domain.NSStanzas) == nil ||
		tl == nil || tl.ChildText("max-file-size", "") != "1048576" {
		t.Fatalf("too large: %s", res)
	}

	t.Log("invalid requests")
	if r := requestSlot(carol, "bad", "", 10, "text/plain"); r.Child("error", "").Child("bad-request", domain.NSStanzas) == nil {
		t.Fatalf("empty filename: %s", r)
	}

	t.Log("upload_service decides which file types are allowed")
	exe := []byte("MZ\x90\x00 not allowed!!")
	res = requestSlot(carol, "exe", "tool.exe", len(exe), "application/octet-stream")
	put := res.Child("slot", domain.NSHTTPUpload).Child("put", "").Attr("url")
	if code, _, _ := httpDo(t, http.MethodPut, put, exe); code != http.StatusUnsupportedMediaType {
		t.Fatalf("exe upload = %d", code)
	}

}
