package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/auth"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type ContractFilesDeps struct {
	fx.In

	Files     port.ContractFileRepository
	Events    port.Repository[model.ContractFileEvent]
	Contracts port.Repository[model.Contract]
	Store     port.DocumentStore
	Scanner   port.Scanner
	Tx        port.TxRunner
	Config    *config.Config
}

// ContractFiles manages the files attached to contracts: who may do what,
// what an upload must pass, versions, and the audit trail. It implements both
// the user-facing use case and the keeper the contract workflows use.
type ContractFiles struct {
	files     port.ContractFileRepository
	events    port.Repository[model.ContractFileEvent]
	contracts port.Repository[model.Contract]
	store     port.DocumentStore
	scanner   port.Scanner
	tx        port.TxRunner
	cfg       config.FilesConfig
	now       func() time.Time
}

func NewContractFiles(d ContractFilesDeps) *ContractFiles {
	cfg := d.Config.Files
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 25 << 20
	}
	if cfg.PDFPolicy == "" {
		cfg.PDFPolicy = policyStrict
	}
	return &ContractFiles{
		files: d.Files, events: d.Events, contracts: d.Contracts, store: d.Store,
		scanner: d.Scanner, tx: d.Tx, cfg: cfg, now: time.Now,
	}
}

func invalidFile(format string, args ...any) error {
	return fmt.Errorf("%w: %s", common.ErrInvalidRequest, fmt.Sprintf(format, args...))
}

func forbidden(format string, args ...any) error {
	return fmt.Errorf("%w: %s", common.ErrForbidden, fmt.Sprintf(format, args...))
}

// ---- audit trail -----------------------------------------------------------

// record adds an entry to the audit trail on behalf of the caller in ctx (or
// "system" for work the service starts itself).
func (s *ContractFiles) record(ctx context.Context, f *model.ContractFile, contractID uint, action, detail string) error {
	ev := &model.ContractFileEvent{At: s.now().UTC(), ContractID: contractID, Action: action, Detail: detail}
	if f != nil {
		ev.FileID, ev.LineageID, ev.Version, ev.FileName, ev.Kind = f.ID, f.Root(), f.Version, f.Name, f.Kind
	}
	if p, ok := auth.FromContext(ctx); ok {
		ev.ActorID, ev.ActorName, ev.ActorService, ev.IP = p.Subject, p.Display(), p.Service, p.IP
		roles := make([]string, len(p.Roles))
		for i, r := range p.Roles {
			roles[i] = string(r)
		}
		ev.ActorRoles = strings.Join(roles, ",")
	} else {
		ev.ActorID, ev.ActorName, ev.ActorService = "system", "system", true
	}
	if len(ev.Detail) > 1000 {
		ev.Detail = ev.Detail[:1000]
	}
	return s.events.Create(ctx, ev)
}

// note records an entry that must not fail the request it describes.
func (s *ContractFiles) note(ctx context.Context, f *model.ContractFile, contractID uint, action, detail string) {
	if err := s.record(ctx, f, contractID, action, detail); err != nil {
		logrus.WithError(err).WithField("action", action).Warn("file audit entry not recorded")
	}
}

// ---- storing files ---------------------------------------------------------

// Prepare validates a file and uploads its bytes (or keeps them in the row
// when there is no object store). Files a user uploads are also checked for
// active content and scanned for malware. It does not record the file: see
// Persist.
func (s *ContractFiles) Prepare(ctx context.Context, spec port.FileSpec) (*model.ContractFile, error) {
	kind, ok := model.LookupFileKind(spec.Kind)
	if !ok {
		return nil, invalidFile("unknown kind %q", spec.Kind)
	}
	system := spec.Source == model.FileSourceSystem
	if !system && kind.System {
		return nil, forbidden("files of kind %q are kept by the service and cannot be uploaded", kind.Name)
	}
	if len(spec.Data) == 0 {
		return nil, invalidFile("the file is empty")
	}
	// The limit is for uploads: a signed renewal the service produced must never
	// be refused for its size.
	if !system && int64(len(spec.Data)) > s.cfg.MaxBytes {
		return nil, fmt.Errorf("%w: the limit is %d MB", common.ErrTooLarge, s.cfg.MaxBytes>>20)
	}

	name := cleanName(spec.Name)
	if name == "" {
		return nil, invalidFile("a file name is required")
	}
	ext := strings.ToLower(path.Ext(name))
	byExt, ok := model.FormatForExtension(ext)
	if !ok {
		return nil, fmt.Errorf("%w: extension %q is not accepted", common.ErrUnsupportedMedia, ext)
	}
	detected, ok := detectFormat(spec.Data)
	if !ok || !matches(detected, byExt) {
		return nil, fmt.Errorf("%w: the content of %q is not a valid %s file", common.ErrUnsupportedMedia, name, strings.ToUpper(byExt.Name))
	}
	if !kind.Allows(byExt.Name) {
		return nil, fmt.Errorf("%w: files of kind %q cannot be %s", common.ErrUnsupportedMedia, kind.Name, strings.ToUpper(byExt.Name))
	}

	f := &model.ContractFile{
		ContractID: spec.ContractID, Kind: kind.Name, Name: name, ContentType: byExt.ContentType, Format: byExt.Name,
		Size: int64(len(spec.Data)), Description: spec.Description, UploadedBy: spec.UploadedBy, UploaderID: spec.UploaderID,
		Source: spec.Source, Version: 1, Latest: true, Data: spec.Data,
	}
	sum := sha256.Sum256(spec.Data)
	f.SHA256 = hex.EncodeToString(sum[:])
	if f.Source == "" {
		f.Source = model.FileSourceUpload
	}

	if system {
		f.ScanStatus = model.ScanSystem
		if f.UploaderID == "" {
			f.UploaderID = "system"
		}
	} else if err := s.check(ctx, f, byExt.Name, spec.Data); err != nil {
		return nil, err
	}

	if s.store.Enabled() {
		var err error
		f.ObjectKey, err = s.newKey(spec.ContractID, kind.Name, ext)
		if err != nil {
			return nil, err
		}
		if err := s.store.Put(ctx, f.ObjectKey, spec.Data, f.ContentType); err != nil {
			return nil, err
		}
		f.Data = nil
	}
	return f, nil
}

// check runs the safety checks on an uploaded file and records their outcome
// on it.
func (s *ContractFiles) check(ctx context.Context, f *model.ContractFile, format string, data []byte) error {
	if reason := inspect(format, data, s.cfg.PDFPolicy); reason != "" {
		return fmt.Errorf("%w: %s", common.ErrMalicious, reason)
	}
	if !s.scanner.Enabled() {
		f.ScanStatus = model.ScanSkipped
		return nil
	}
	res, err := s.scanner.Scan(ctx, f.Name, data)
	switch {
	case err != nil && s.cfg.ScanRequired:
		logrus.WithError(err).Error("virus scan failed")
		return fmt.Errorf("%w: the virus scanner is unavailable, try again later", common.ErrUpstream)
	case err != nil:
		logrus.WithError(err).Warn("virus scan failed; accepting the file unscanned")
		f.ScanStatus = model.ScanSkipped
	case !res.Clean:
		return fmt.Errorf("%w: a virus was found (%s)", common.ErrMalicious, res.Signature)
	default:
		now := s.now().UTC()
		f.ScanStatus, f.ScanEngine, f.ScannedAt = model.ScanClean, res.Engine, &now
	}
	return nil
}

// newKey makes a unique object key: unlike a content hash it never collides
// with another file's, so discarding one upload cannot remove another's bytes.
func (s *ContractFiles) newKey(contractID uint, kind, ext string) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("contracts/%d/%s/%s-%s%s", contractID, kind, s.now().UTC().Format("20060102"), hex.EncodeToString(b[:]), ext), nil
}

// Persist records a prepared file and its audit entry, and must run inside the
// caller's transaction. A signed renewal automatically becomes the next
// version of the contract's previous one.
func (s *ContractFiles) Persist(ctx context.Context, f *model.ContractFile) error {
	f.Latest = true
	if f.Version == 0 {
		f.Version = 1
	}
	if f.Source == model.FileSourceSystem && f.Kind == model.FileKindRenewal && f.ReplacesID == nil {
		prev, err := s.Latest(ctx, f.ContractID, f.Kind)
		switch {
		case err == nil:
			f.ReplacesID, f.LineageID, f.Version = &prev.ID, prev.Root(), prev.Version+1
		case !errors.Is(err, common.ErrNotFound):
			return err
		}
	}

	if err := s.files.Create(ctx, f); err != nil {
		return err
	}
	action, detail := model.FileEventUploaded, fmt.Sprintf("%s, %d bytes, sha256 %s, scan %s", f.Name, f.Size, f.SHA256, f.ScanStatus)
	if f.ReplacesID != nil {
		superseded, err := s.files.SetLatest(ctx, *f.ReplacesID, false)
		if err != nil {
			return err
		}
		if !superseded {
			return fmt.Errorf("%w: the file was replaced by someone else in the meantime", common.ErrInvalidTransition)
		}
		action, detail = model.FileEventNewVersion, fmt.Sprintf("version %d replaces file %d; %s", f.Version, *f.ReplacesID, detail)
	}
	return s.record(ctx, f, f.ContractID, action, detail)
}

func (s *ContractFiles) Discard(ctx context.Context, f *model.ContractFile) {
	if f != nil && f.ObjectKey != "" {
		_ = s.store.Delete(ctx, f.ObjectKey)
	}
}

// Latest returns the current version of the newest file of a kind.
func (s *ContractFiles) Latest(ctx context.Context, contractID uint, kind string) (*model.ContractFile, error) {
	q := port.ListQuery{Limit: 1, Equals: map[string]any{"contract_id": contractID, "kind": kind, "latest": true}}
	_, total, err := s.files.List(ctx, q)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, common.ErrNotFound
	}
	q.Offset = int(total) - 1 // rows come back oldest first
	rows, _, err := s.files.List(ctx, q)
	if err != nil || len(rows) == 0 {
		return nil, common.ErrDBFailed
	}
	return &rows[0], nil
}

func (s *ContractFiles) Open(ctx context.Context, f *model.ContractFile) ([]byte, error) {
	if f.ObjectKey == "" {
		return f.Data, nil
	}
	data, err := s.store.Get(ctx, f.ObjectKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != f.SHA256 {
		return nil, common.ErrIntegrity
	}
	return data, nil
}

// ---- use cases -------------------------------------------------------------

func caller(ctx context.Context) (auth.Principal, error) {
	p, ok := auth.FromContext(ctx)
	if !ok {
		return auth.Principal{}, forbidden("authentication is required")
	}
	return p, nil
}

func (s *ContractFiles) Upload(ctx context.Context, in port.FileUpload) (*model.ContractFile, error) {
	p, err := caller(ctx)
	if err != nil {
		return nil, err
	}
	if _, ok := model.LookupFileKind(in.Kind); !ok {
		return nil, invalidFile("unknown kind %q", in.Kind)
	}
	if !p.Can(auth.FileUpload, in.Kind) {
		s.note(ctx, nil, in.ContractID, model.FileEventDenied, fmt.Sprintf("upload of %q as kind %s", in.Name, in.Kind))
		return nil, forbidden("your role may not upload files of kind %q", in.Kind)
	}
	if _, err := s.contracts.GetByID(ctx, in.ContractID); err != nil {
		return nil, err
	}

	var old *model.ContractFile
	if in.ReplacesID != 0 {
		if old, err = s.replaceable(ctx, p, in); err != nil {
			return nil, err
		}
	}

	f, err := s.Prepare(ctx, port.FileSpec{
		ContractID: in.ContractID, Kind: in.Kind, Name: in.Name, Description: in.Description,
		UploadedBy: p.Display(), UploaderID: p.Subject, Source: model.FileSourceUpload, Data: in.Data,
	})
	if err != nil {
		s.note(ctx, nil, in.ContractID, model.FileEventRejected, fmt.Sprintf("%q as kind %s: %v", in.Name, in.Kind, err))
		return nil, err
	}
	if old != nil {
		f.ReplacesID, f.LineageID, f.Version = &old.ID, old.Root(), old.Version+1
	}

	if err := s.tx.InTx(ctx, func(ctx context.Context) error { return s.Persist(ctx, f) }); err != nil {
		s.Discard(ctx, f)
		return nil, err
	}
	return f, nil
}

// replaceable finds the file an upload is a new version of and checks the
// caller may replace it.
func (s *ContractFiles) replaceable(ctx context.Context, p auth.Principal, in port.FileUpload) (*model.ContractFile, error) {
	old, err := s.files.GetByID(ctx, in.ReplacesID)
	if err != nil {
		return nil, err
	}
	switch {
	case old.ContractID != in.ContractID:
		return nil, common.ErrNotFound
	case old.Kind != in.Kind:
		return nil, invalidFile("a new version must have the same kind (%s)", old.Kind)
	case !old.Latest:
		return nil, fmt.Errorf("%w: only the latest version can be replaced", common.ErrInvalidTransition)
	case !p.Can(auth.FileDeleteAny, in.Kind) && old.UploaderID != p.Subject:
		s.note(ctx, old, in.ContractID, model.FileEventDenied, "replace a file uploaded by someone else")
		return nil, forbidden("only the uploader or a manager may replace this file")
	}
	if kind, _ := model.LookupFileKind(old.Kind); kind.System {
		return nil, forbidden("files of kind %q cannot be replaced", old.Kind)
	}
	return old, nil
}

func (s *ContractFiles) List(ctx context.Context, contractID uint, f port.FileListFilter) ([]model.ContractFile, int64, error) {
	p, err := caller(ctx)
	if err != nil {
		return nil, 0, err
	}
	if p.Highest() == "" {
		return nil, 0, forbidden("your account has no role for contract files")
	}
	if _, err := s.contracts.GetByID(ctx, contractID); err != nil {
		return nil, 0, err
	}

	q := port.ListQuery{Offset: f.Offset, Limit: f.Limit, Equals: map[string]any{"contract_id": contractID}}
	if !f.AllVersions {
		q.Equals["latest"] = true
	}
	if f.Kind != "" {
		if _, ok := model.LookupFileKind(f.Kind); !ok {
			return nil, 0, invalidFile("unknown kind %q", f.Kind)
		}
		q.Equals["kind"] = f.Kind
	}
	if hidden := p.HiddenKinds(); len(hidden) > 0 {
		values := make([]any, len(hidden))
		for i, h := range hidden {
			values[i] = h
		}
		q.Exclude = map[string][]any{"kind": values}
	}
	return s.files.List(ctx, q)
}

// load returns a file of the given contract that the caller may read. A file
// they may not read is reported as missing rather than forbidden, so its
// existence is not revealed.
func (s *ContractFiles) load(ctx context.Context, p auth.Principal, contractID, fileID uint) (*model.ContractFile, error) {
	f, err := s.files.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if f.ContractID != contractID || !p.Can(auth.FileRead, f.Kind) {
		return nil, common.ErrNotFound
	}
	return f, nil
}

func (s *ContractFiles) Get(ctx context.Context, contractID, fileID uint) (*model.ContractFile, error) {
	p, err := caller(ctx)
	if err != nil {
		return nil, err
	}
	return s.load(ctx, p, contractID, fileID)
}

func (s *ContractFiles) Download(ctx context.Context, contractID, fileID uint) (*model.ContractFile, []byte, error) {
	p, err := caller(ctx)
	if err != nil {
		return nil, nil, err
	}
	f, err := s.load(ctx, p, contractID, fileID)
	if err != nil {
		return nil, nil, err
	}
	data, err := s.Open(ctx, f)
	if err != nil {
		return nil, nil, err
	}
	s.note(ctx, f, contractID, model.FileEventDownloaded, "")
	return f, data, nil
}

func (s *ContractFiles) Delete(ctx context.Context, contractID, fileID uint) error {
	p, err := caller(ctx)
	if err != nil {
		return err
	}
	f, err := s.load(ctx, p, contractID, fileID)
	if err != nil {
		return err
	}

	kind, _ := model.LookupFileKind(f.Kind)
	switch {
	case kind.System:
		s.note(ctx, f, contractID, model.FileEventDenied, "delete evidence")
		return forbidden("files of kind %q are evidence and cannot be deleted", f.Kind)
	case p.Can(auth.FileDeleteAny, f.Kind):
	case p.Can(auth.FileDeleteOwn, f.Kind) && f.UploaderID == p.Subject:
	default:
		s.note(ctx, f, contractID, model.FileEventDenied, "delete")
		return forbidden("only the uploader or a manager may delete this file")
	}

	return s.tx.InTx(ctx, func(ctx context.Context) error {
		if err := s.files.Delete(ctx, fileID); err != nil {
			return err
		}
		// Deleting the current version brings the previous one back.
		if f.Latest && f.ReplacesID != nil {
			if _, err := s.files.SetLatest(ctx, *f.ReplacesID, true); err != nil {
				return err
			}
		}
		return s.record(ctx, f, contractID, model.FileEventDeleted, "")
	})
}

func (s *ContractFiles) History(ctx context.Context, contractID, fileID uint) (*port.FileHistory, error) {
	p, err := caller(ctx)
	if err != nil {
		return nil, err
	}
	f, err := s.load(ctx, p, contractID, fileID)
	if err != nil {
		return nil, err
	}
	versions, err := s.files.Lineage(ctx, f.Root())
	if err != nil {
		return nil, err
	}
	events, _, err := s.events.List(ctx, port.ListQuery{Limit: 1000, Equals: map[string]any{"lineage_id": f.Root()}})
	if err != nil {
		return nil, err
	}
	return &port.FileHistory{Versions: versions, Events: events}, nil
}

func (s *ContractFiles) Audit(ctx context.Context, contractID uint, f port.FileAuditFilter) ([]model.ContractFileEvent, int64, error) {
	p, err := caller(ctx)
	if err != nil {
		return nil, 0, err
	}
	if !p.Can(auth.FileAudit, "") {
		s.note(ctx, nil, contractID, model.FileEventDenied, "read the audit trail")
		return nil, 0, forbidden("only managers may read the audit trail")
	}
	if _, err := s.contracts.GetByID(ctx, contractID); err != nil {
		return nil, 0, err
	}
	q := port.ListQuery{Offset: f.Offset, Limit: f.Limit, Equals: map[string]any{"contract_id": contractID}}
	if f.FileID != 0 {
		q.Equals["file_id"] = f.FileID
	}
	if f.Action != "" {
		q.Equals["action"] = f.Action
	}
	return s.events.List(ctx, q)
}
