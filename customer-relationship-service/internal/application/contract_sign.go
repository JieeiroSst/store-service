package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.uber.org/fx"
)

type ContractWorkflowDeps struct {
	fx.In

	Repo         port.Repository[model.Contract]
	Files        port.ContractFileKeeper
	Accounts     port.Repository[model.Account]
	Sealer       port.CounterSigner
	Tx           port.TxRunner
	Verifier     port.SignatureVerifier
	PDFVerifier  port.PDFSignatureVerifier
	Timestamper  port.Timestamper
	Renderer     port.ContractRenderer
	Notifier     port.Notifier
	Orchestrator port.ContractOrchestrator
}

type contractWorkflow struct {
	repo         port.Repository[model.Contract]
	files        port.ContractFileKeeper
	accounts     port.Repository[model.Account]
	sealer       port.CounterSigner
	tx           port.TxRunner
	verifier     port.SignatureVerifier
	pdfVerifier  port.PDFSignatureVerifier
	stamper      port.Timestamper
	renderer     port.ContractRenderer
	notifier     port.Notifier
	orchestrator port.ContractOrchestrator
	now          func() time.Time
}

func NewContractWorkflow(d ContractWorkflowDeps) port.ContractWorkflow {
	return &contractWorkflow{
		repo: d.Repo, files: d.Files, accounts: d.Accounts, sealer: d.Sealer, tx: d.Tx, verifier: d.Verifier, pdfVerifier: d.PDFVerifier,
		stamper: d.Timestamper, renderer: d.Renderer, notifier: d.Notifier, orchestrator: d.Orchestrator, now: time.Now,
	}
}

const maxSigningSkew = 5 * time.Minute

func (w *contractWorkflow) Sign(ctx context.Context, id uint, in port.SignContractInput) (*model.Contract, error) {
	now := w.now()
	endDate := in.EndDate.UTC().Truncate(time.Second)
	signingTime := in.SigningTime.UTC().Truncate(time.Second)

	if strings.TrimSpace(in.SignedBy) == "" || !endDate.After(now) {
		return nil, common.ErrInvalidRequest
	}
	if skew := now.Sub(signingTime); skew > maxSigningSkew || skew < -maxSigningSkew {
		return nil, common.ErrInvalidRequest
	}
	detached := in.Algorithm != "" || in.Signature != "" || in.Certificate != ""
	if detached == (in.SignedPDF != "") {
		return nil, common.ErrInvalidRequest
	}
	if detached && (in.Algorithm == "" || in.Signature == "" || in.Certificate == "") {
		return nil, common.ErrInvalidRequest
	}

	c, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != model.ContractExpired {
		return nil, common.ErrInvalidTransition
	}

	var (
		verified *port.VerifiedSignature
		signed   []byte
	)
	if detached {
		verified, err = w.verifyDetached(ctx, c, in, endDate, signingTime)
	} else {
		signed, verified, err = w.verifyPDF(ctx, c, in, endDate, signingTime)
	}
	if err != nil {
		return nil, err
	}

	var seal *port.VerifiedSignature
	if signed != nil && w.sealer.Enabled() {
		signed, seal, err = w.sealer.CounterSign(ctx, signed, now)
		if err != nil {
			return nil, err
		}
	}

	var file *model.ContractFile
	if signed != nil {
		file, err = w.files.Prepare(ctx, port.FileSpec{
			ContractID:  id,
			Kind:        model.FileKindRenewal,
			Name:        renewalFileName(c),
			Description: "Phụ lục gia hạn hợp đồng đã ký số",
			UploadedBy:  in.SignedBy,
			Source:      model.FileSourceSystem,
			Data:        signed,
		})
		if err != nil {
			return nil, err
		}
	}

	c.Status, c.Approval = model.ContractApproved, model.ContractApproved
	c.SignedBy, c.SignedAt, c.EndDate = in.SignedBy, &now, &endDate
	c.SignatureFormat, c.RevocationStatus = verified.Format, verified.Revocation
	c.SignatureAlgorithm, c.Signature = verified.Algorithm, verified.Signature
	c.SignerCertificate, c.SignerSubject = verified.Certificate, verified.Subject
	c.SignerSerial, c.SignerFingerprint = verified.Serial, verified.Fingerprint
	c.SignaturePayloadHash = verified.PayloadHash
	c.CountersignerSubject, c.CountersignerSerial, c.CountersignerFingerprint, c.CountersignedAt = "", "", "", nil
	if seal != nil {
		at := now
		if seal.Timestamp != nil {
			at = seal.Timestamp.Time
		}
		c.CountersignerSubject, c.CountersignerSerial, c.CountersignerFingerprint = seal.Subject, seal.Serial, seal.Fingerprint
		c.CountersignedAt = &at
	}
	c.TimestampToken, c.TimestampAt, c.TimestampAuthority = "", nil, ""
	if ts := verified.Timestamp; ts != nil {
		at := ts.Time
		c.TimestampToken, c.TimestampAt, c.TimestampAuthority = ts.Token, &at, ts.Authority
	}

	err = w.tx.InTx(ctx, func(ctx context.Context) error {
		if err := w.repo.Update(ctx, id, c); err != nil {
			return err
		}
		if file == nil {
			return nil
		}
		return w.files.Persist(ctx, file)
	})
	if err != nil {
		w.files.Discard(ctx, file) // do not leave an orphan
		return nil, err
	}
	syncLifecycle(ctx, w.orchestrator, id)
	notify(ctx, w.notifier, "Contract renewed", "Contract #%d was signed again by %s (%s) until %s", c.ID, in.SignedBy, verified.Subject, endDate.Format("2006-01-02"))
	return w.repo.GetByID(ctx, id)
}

func (w *contractWorkflow) verifyDetached(ctx context.Context, c *model.Contract, in port.SignContractInput, endDate, signingTime time.Time) (*port.VerifiedSignature, error) {
	verified, err := w.verifier.Verify(ctx, model.SigningPayload(c, in.SignedBy, endDate, signingTime), port.SignatureEvidence{
		Algorithm:      in.Algorithm,
		Signature:      in.Signature,
		CertificatePEM: in.Certificate,
	}, signingTime)
	if err != nil {
		return nil, err
	}
	sig, err := base64.StdEncoding.DecodeString(in.Signature)
	if err != nil {
		return nil, common.ErrInvalidRequest
	}
	verified.Timestamp, err = w.stamper.Stamp(ctx, sig)
	if err != nil {
		return nil, err
	}
	return verified, nil
}

func (w *contractWorkflow) verifyPDF(ctx context.Context, c *model.Contract, in port.SignContractInput, endDate, signingTime time.Time) ([]byte, *port.VerifiedSignature, error) {
	signed, err := base64.StdEncoding.DecodeString(in.SignedPDF)
	if err != nil || len(signed) == 0 {
		return nil, nil, common.ErrInvalidRequest
	}
	issued, err := w.render(ctx, c, in.SignedBy, endDate, signingTime)
	if err != nil {
		return nil, nil, err
	}

	if !bytes.HasPrefix(signed, issued) {
		return nil, nil, fmt.Errorf("%w: the signed PDF is not the document that was issued for these parameters", common.ErrInvalidSignature)
	}
	verified, err := w.pdfVerifier.VerifyPDF(ctx, signed, signingTime)
	if err != nil {
		return nil, nil, err
	}
	return signed, verified, nil
}

func (w *contractWorkflow) Document(ctx context.Context, id uint, signedBy string, endDate, signingTime time.Time) ([]byte, error) {
	c, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return w.render(ctx, c, signedBy, endDate.UTC().Truncate(time.Second), signingTime.UTC().Truncate(time.Second))
}

func (w *contractWorkflow) render(ctx context.Context, c *model.Contract, signedBy string, endDate, signingTime time.Time) ([]byte, error) {
	account, err := w.accounts.GetByID(ctx, c.AccountID)
	if err != nil {
		return nil, err
	}
	return w.renderer.Render(port.RenewalDocument{
		Contract: c, Account: account, SignedBy: signedBy, EndDate: endDate, SigningTime: signingTime,
	})
}

func (w *contractWorkflow) SignedDocument(ctx context.Context, id uint) ([]byte, error) {
	f, err := w.files.Latest(ctx, id, model.FileKindRenewal)
	if err != nil {
		return nil, err
	}
	return w.files.Open(ctx, f)
}

func shortHash(h string) string {
	if len(h) > 16 {
		return h[:16]
	}
	return h
}

func renewalFileName(c *model.Contract) string {
	number := c.Number
	if strings.TrimSpace(number) == "" {
		number = fmt.Sprintf("%d", c.ID)
	}
	number = strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(number)
	return "Phu-luc-gia-han-" + number + ".pdf"
}
