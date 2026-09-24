package digisign

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/digitorus/pdf"
	"github.com/digitorus/pkcs7"
	"github.com/digitorus/timestamp"
)

var (
	oidTimestampToken = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14}
	oidSHA256         = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidSHA384         = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2}
	oidSHA512         = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3}
)

type pdfSignature struct {
	contents  []byte
	byteRange [4]int64
}

func (v *Verifier) VerifyPDF(ctx context.Context, doc []byte, at time.Time) (result *port.VerifiedSignature, err error) {
	defer func() {
		if r := recover(); r != nil {
			result, err = nil, invalid("malformed PDF")
		}
	}()

	sig, err := findCoveringSignature(doc)
	if err != nil {
		return nil, err
	}
	p7, err := pkcs7.Parse(sig.contents)
	if err != nil {
		return nil, invalid("cannot parse the PDF signature: %v", err)
	}
	if len(p7.Signers) != 1 {
		return nil, invalid("expected exactly one signer, found %d", len(p7.Signers))
	}
	p7.Content = append(append([]byte{}, doc[:sig.byteRange[1]]...), doc[sig.byteRange[2]:]...)

	signer := p7.Signers[0]
	if alg := signer.DigestAlgorithm.Algorithm; !alg.Equal(oidSHA256) && !alg.Equal(oidSHA384) && !alg.Equal(oidSHA512) {
		return nil, invalid("PDF signature uses a weak digest algorithm")
	}
	if err := p7.Verify(); err != nil {
		return nil, invalid("PDF signature does not match the document: %v", err)
	}
	leaf := p7.GetOnlySigner()
	if leaf == nil {
		return nil, invalid("the PDF does not include the signer's certificate")
	}
	if err := checkKeyStrength(leaf.PublicKey); err != nil {
		return nil, err
	}
	if err := checkKeyUsage(leaf); err != nil {
		return nil, err
	}

	var token []byte
	for _, a := range signer.UnauthenticatedAttributes {
		if a.Type.Equal(oidTimestampToken) {
			token = a.Value.Bytes
			break
		}
	}
	evidence, judgedAt, err := v.embeddedTimestamp(ctx, token, signer.EncryptedDigest, at)
	if err != nil {
		return nil, err
	}

	intermediates := x509.NewCertPool()
	for _, c := range p7.Certificates {
		if !c.Equal(leaf) {
			intermediates.AddCert(c)
		}
	}
	revocation, err := v.checkTrust(ctx, leaf, intermediates, judgedAt)
	if err != nil {
		return nil, err
	}

	out := describe(leaf, formatPAdES, "PDF-CMS", base64.StdEncoding.EncodeToString(signer.EncryptedDigest), revocation)
	out.PayloadHash = hexSHA256(doc)
	out.Timestamp = evidence
	return out, nil
}

func (v *Verifier) embeddedTimestamp(ctx context.Context, token []byte, encryptedDigest []byte, at time.Time) (*port.TimestampEvidence, time.Time, error) {
	if token == nil {
		return nil, at, nil
	}
	ts, err := timestamp.Parse(token)
	if err != nil {
		return nil, at, invalid("cannot parse the embedded time-stamp: %v", err)
	}
	authority, err := v.validateTimestamp(ctx, ts, encryptedDigest, at)
	if err != nil {
		return nil, at, err
	}
	return &port.TimestampEvidence{
		Token:     base64.StdEncoding.EncodeToString(ts.RawToken),
		Time:      ts.Time.UTC(),
		Authority: authority,
	}, ts.Time, nil
}

func findCoveringSignature(doc []byte) (*pdfSignature, error) {
	size := int64(len(doc))
	r, err := pdf.NewReader(bytes.NewReader(doc), size)
	if err != nil {
		return nil, invalid("cannot read the PDF: %v", err)
	}

	var found *pdfSignature
	sawSignature := false
	for _, x := range r.Xref() {
		val := r.Resolve(x.Ptr(), x.Ptr())
		if val.Key("Filter").Name() != "Adobe.PPKLite" {
			continue
		}
		sawSignature = true

		br := val.Key("ByteRange")
		if br.Len() != 4 {
			continue
		}
		var rng [4]int64
		for i := range rng {
			rng[i] = br.Index(i).Int64()
		}

		if rng[0] != 0 || rng[1] <= 0 || rng[1] >= rng[2] || rng[2]+rng[3] != size {
			continue
		}
		if doc[rng[1]] != '<' || doc[rng[2]-1] != '>' {
			continue
		}
		contents := []byte(val.Key("Contents").RawString())
		if len(contents) == 0 {
			continue
		}
		if found == nil || rng[1] > found.byteRange[1] {
			found = &pdfSignature{contents: contents, byteRange: rng}
		}
	}
	if found == nil {
		if sawSignature {
			return nil, invalid("the PDF signature does not cover the whole document")
		}
		return nil, invalid("the PDF is not signed")
	}
	return found, nil
}
