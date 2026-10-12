package application

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
	"go.uber.org/fx"
)

const defaultContentType = "application/octet-stream"

type uploadHost struct {
	serverHost string
	host       string
	putURL     string
	getURL     string
}

type HTTPUploadService struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	storage port.FileStorage
	router  *Router
	key     []byte
	hosts   []uploadHost
	now     func() time.Time
}

type HTTPUploadParams struct {
	fx.In

	Config  *config.Config
	Matcher *domain.ACLMatcher
	Router  *Router
	Storage port.FileStorage `optional:"true"`
}

func NewHTTPUploadService(p HTTPUploadParams) (*HTTPUploadService, error) {
	s := &HTTPUploadService{cfg: p.Config, matcher: p.Matcher, storage: p.Storage, router: p.Router, now: time.Now}
	mod := p.Config.Modules.HTTPUpload
	if !mod.Enabled {
		return s, nil
	}
	s.key = []byte(mod.ExternalSecret)
	if len(s.key) == 0 {
		s.key = make([]byte, 32)
		if _, err := rand.Read(s.key); err != nil {
			return nil, err
		}
		slog.Warn("mod_http_upload: no external_secret configured, upload slots become invalid after a restart")
	}
	for _, h := range p.Config.Hosts {
		get := mod.GetURL
		if get == "" {
			get = mod.PutURL
		}
		s.hosts = append(s.hosts, uploadHost{
			serverHost: h,
			host:       config.ExpandHost(mod.Host, h),
			putURL:     strings.TrimRight(config.ExpandHost(mod.PutURL, h), "/"),
			getURL:     strings.TrimRight(config.ExpandHost(get, h), "/"),
		})
	}
	return s, nil
}

func (s *HTTPUploadService) Enabled() bool {
	return s.cfg.Modules.HTTPUpload.Enabled && s.storage != nil
}

func (s *HTTPUploadService) Register() {
	if !s.Enabled() {
		return
	}
	for _, h := range s.hosts {
		h := h
		s.router.RegisterService(ServiceRoute{
			Host: h.host,
			Name: s.cfg.Modules.HTTPUpload.Name,
			Handler: func(ctx context.Context, stanza *domain.Element) {
				if res := s.processStanza(ctx, h, stanza); res != nil {
					s.router.Route(ctx, res)
				}
			},
		})
		slog.Info("mod_http_upload: service registered", "host", h.host, "put_url", h.putURL, "get_url", h.getURL)
	}
}

func (s *HTTPUploadService) processStanza(ctx context.Context, h uploadHost, stanza *domain.Element) *domain.Element {
	if stanza.Name != domain.KindIQ {
		return nil
	}
	typ := domain.StanzaType(stanza)
	if typ != "get" && typ != "set" {
		return nil
	}
	els := stanza.Elements()
	if typ == "get" && len(els) == 1 {
		q := els[0]
		switch {
		case q.Name == "query" && q.Namespace() == domain.NSDiscoInfo:
			return domain.MakeIQResult(stanza, s.discoInfo())
		case q.Name == "query" && q.Namespace() == domain.NSDiscoItems:
			return domain.MakeIQResult(stanza, domain.NewElement("query", domain.NSDiscoItems))
		case q.Name == "request" && q.Namespace() == domain.NSHTTPUpload:
			return s.slotRequest(stanza, h, q)
		}
	}
	return domain.MakeError(stanza, domain.ErrServiceUnavailable("No module is handling this query"))
}

func (s *HTTPUploadService) discoInfo() *domain.Element {
	q := domain.NewElement("query", domain.NSDiscoInfo)
	q.Append(domain.NewElement("identity", "").SetAttr("category", "store").SetAttr("type", "file").SetAttr("name", s.cfg.Modules.HTTPUpload.Name))
	for _, f := range []string{domain.NSHTTPUpload, domain.NSDiscoInfo, domain.NSDiscoItems} {
		q.Append(domain.NewElement("feature", "").SetAttr("var", f))
	}
	if max := s.cfg.Modules.HTTPUpload.MaxSize; max > 0 {
		x := domain.NewElement("x", domain.NSXData).SetAttr("type", "result")
		ft := domain.NewElement("field", "").SetAttr("var", "FORM_TYPE").SetAttr("type", "hidden")
		ft.Append(domain.NewElement("value", "").AppendText(domain.NSHTTPUpload))
		mf := domain.NewElement("field", "").SetAttr("var", "max-file-size")
		mf.Append(domain.NewElement("value", "").AppendText(strconv.FormatInt(max, 10)))
		x.Append(ft, mf)
		q.Append(x)
	}
	return q
}

func (s *HTTPUploadService) slotRequest(iq *domain.Element, h uploadHost, req *domain.Element) *domain.Element {
	from, ok := domain.StanzaFrom(iq)
	if !ok {
		return domain.MakeError(iq, domain.ErrBadRequest("Missing 'from'"))
	}
	lang := domain.Lang(iq)
	if !s.matcher.MatchRule(s.cfg.Modules.HTTPUpload.Access, domain.MatchJID(from)) {
		e := domain.ErrForbidden("Access denied by service policy")
		e.Lang = lang
		return domain.MakeError(iq, e)
	}
	fileName := req.Attr("filename")
	size, err := strconv.ParseInt(req.Attr("size"), 10, 64)
	if fileName == "" || err != nil || size <= 0 {
		return domain.MakeError(iq, domain.ErrBadRequest("Invalid 'filename' or 'size' attribute"))
	}
	if max := s.cfg.Modules.HTTPUpload.MaxSize; max > 0 && size > max {
		slog.Warn("mod_http_upload: rejecting file, too large", "file", fileName, "jid", from.String(), "size", size)
		return domain.MakeError(iq, domain.FileTooLarge(max, lang))
	}
	slot, err := domain.NewUploadSlot(from, fileName, size, s.now())
	if err != nil {
		return domain.MakeError(iq, domain.ErrInternalServerError(""))
	}
	sig := domain.SignSlot(s.key, slot.Path(), slot.Size, slot.Expires)
	path := escapePath(slot.Path())
	putURL := h.putURL + "/" + path + "?s=" + strconv.FormatInt(slot.Size, 10) + "&e=" + strconv.FormatInt(slot.Expires, 10) + "&v=" + sig
	getURL := h.getURL + "/" + path
	el := domain.NewElement("slot", domain.NSHTTPUpload)
	put := domain.NewElement("put", "").SetAttr("url", putURL)
	el.Append(put, domain.NewElement("get", "").SetAttr("url", getURL))
	slog.Info("mod_http_upload: slot created", "jid", from.String(), "file", fileName, "size", size,
		"content_type", contentType(req.Attr("content-type")))
	return domain.MakeIQResult(iq, el)
}

func contentType(ct string) string {
	if ct == "" {
		return defaultContentType
	}
	return ct
}

func escapePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func parseSlotPath(segments []string) (domain.UploadSlot, bool) {
	if len(segments) != 3 {
		return domain.UploadSlot{}, false
	}
	for _, seg := range segments {
		if seg == "" {
			return domain.UploadSlot{}, false
		}
	}
	return domain.UploadSlot{UserStr: segments[0], Secret: segments[1], FileName: segments[2]}, true
}

func (s *HTTPUploadService) Put(ctx context.Context, segments []string, query url.Values, length int64, body io.Reader) error {
	slot, ok := parseSlotPath(segments)
	if !ok {
		return domain.ErrFileNotFound
	}
	expires, err1 := strconv.ParseInt(query.Get("e"), 10, 64)
	size, err2 := strconv.ParseInt(query.Get("s"), 10, 64)
	if err1 != nil || err2 != nil {
		return domain.ErrInvalidSlot
	}
	if !domain.VerifySlot(s.key, slot.Path(), size, expires, query.Get("v"), s.now()) {
		return domain.ErrInvalidSlot
	}
	if length != size {
		return domain.ErrSlotSizeMismatch
	}
	used, err := s.storage.Exists(ctx, slot.StorageKey())
	if err != nil {
		return err
	}
	if used {
		return domain.ErrInvalidSlot
	}
	return s.storage.Store(ctx, slot.StorageKey(), slot.FileName, length, io.LimitReader(body, length))
}

func (s *HTTPUploadService) Get(ctx context.Context, segments []string) (domain.StoredFile, error) {
	slot, ok := parseSlotPath(segments)
	if !ok {
		return domain.StoredFile{}, domain.ErrFileNotFound
	}
	f, err := s.storage.Fetch(ctx, slot.StorageKey())
	if err != nil {
		return domain.StoredFile{}, err
	}
	if f.Name != "" && f.Name != slot.FileName {
		f.Body.Close()
		return domain.StoredFile{}, domain.ErrFileNotFound
	}
	f.Name = slot.FileName
	return f, nil
}

func IsUploadClientError(err error) bool {
	return errors.Is(err, domain.ErrInvalidSlot) || errors.Is(err, domain.ErrSlotSizeMismatch) ||
		errors.Is(err, domain.ErrFileTooLarge) || errors.Is(err, domain.ErrUnsupportedFile) ||
		errors.Is(err, domain.ErrFileNotFound)
}
