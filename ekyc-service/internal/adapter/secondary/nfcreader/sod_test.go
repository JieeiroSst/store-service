package nfcreader

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"

	"go.mozilla.org/pkcs7"
)

// oidSHA256 is the ASN.1 OID ICAO 9303 LDSSecurityObject uses to declare
// SHA-256 as its data-group hash algorithm.
var oidSHA256 = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}

// selfSignedTestCert generates a throwaway self-signed EC certificate,
// standing in for a real Document Signer Certificate (which would chain
// to a national CSCA this test has no access to).
func selfSignedTestCert(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test Document Signer"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return cert, key
}

// buildTestEFSOD builds a real CMS SignedData wrapping an
// ICAO-9303-shaped LDSSecurityObject over the given data groups, signed
// with a throwaway self-signed cert, and wraps it in the EF.SOD BER-TLV
// outer tag (0x77) the same way a real chip dump would.
func buildTestEFSOD(t *testing.T, dataGroups map[int][]byte) []byte {
	t.Helper()

	hashes := make([]dataGroupHash, 0, len(dataGroups))
	for num, raw := range dataGroups {
		sum := sha256.Sum256(raw)
		hashes = append(hashes, dataGroupHash{DataGroupNumber: num, HashValue: sum[:]})
	}
	lds := ldsSecurityObject{
		Version:         0,
		HashAlgorithm:   algorithmIdentifier{Algorithm: oidSHA256},
		DataGroupHashes: hashes,
	}
	content, err := asn1.Marshal(lds)
	if err != nil {
		t.Fatalf("marshal LDSSecurityObject: %v", err)
	}

	cert, key := selfSignedTestCert(t)
	sd, err := pkcs7.NewSignedData(content)
	if err != nil {
		t.Fatalf("NewSignedData: %v", err)
	}
	if err := sd.AddSigner(cert, key, pkcs7.SignerInfoConfig{}); err != nil {
		t.Fatalf("AddSigner: %v", err)
	}
	cms, err := sd.Finish()
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}

	return encodeTLV(tagEFSOD, cms)
}

func TestParseAndVerifySOD_ValidDataGroups(t *testing.T) {
	dg1 := []byte("fake DG1 content for hashing")
	dg2 := []byte("fake DG2 content for hashing")
	dataGroups := map[int][]byte{1: dg1, 2: dg2}

	efsod := buildTestEFSOD(t, dataGroups)

	result, err := ParseAndVerifySOD(efsod, dataGroups)
	if err != nil {
		t.Fatalf("ParseAndVerifySOD: %v", err)
	}
	if !result.SignatureSelfConsistent {
		t.Errorf("expected SignatureSelfConsistent = true for an untampered SOD")
	}
	if !result.DataGroupHashValid[1] || !result.DataGroupHashValid[2] {
		t.Errorf("expected both DG1 and DG2 hashes valid, got %+v", result.DataGroupHashValid)
	}
}

func TestParseAndVerifySOD_TamperedDataGroup(t *testing.T) {
	dg1 := []byte("fake DG1 content for hashing")
	dg2 := []byte("fake DG2 content for hashing")
	dataGroups := map[int][]byte{1: dg1, 2: dg2}

	efsod := buildTestEFSOD(t, dataGroups)

	tampered := map[int][]byte{1: dg1, 2: []byte("this DG2 was altered after signing")}
	result, err := ParseAndVerifySOD(efsod, tampered)
	if err != nil {
		t.Fatalf("ParseAndVerifySOD: %v", err)
	}
	if !result.DataGroupHashValid[1] {
		t.Errorf("expected untouched DG1 hash to still be valid")
	}
	if result.DataGroupHashValid[2] {
		t.Errorf("expected tampered DG2 hash to be invalid")
	}
	// The CMS signature itself is still self-consistent - it only covers
	// the LDSSecurityObject's declared hash list, not the raw DGs
	// themselves, which is exactly why per-DG hash comparison above is
	// the check that catches tampering with a data group's content.
	if !result.SignatureSelfConsistent {
		t.Errorf("expected SignatureSelfConsistent = true (SOD itself wasn't altered)")
	}
}
