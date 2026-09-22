package nfcreader

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/asn1"
	"fmt"
	"hash"

	"go.mozilla.org/pkcs7"
)

const tagEFSOD uint32 = 0x77

type algorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

type dataGroupHash struct {
	DataGroupNumber int
	HashValue       []byte
}

type ldsSecurityObject struct {
	Version         int
	HashAlgorithm   algorithmIdentifier
	DataGroupHashes []dataGroupHash
}

type SODResult struct {
	HashAlgorithm           string
	DataGroupHashValid      map[int]bool
	SignatureSelfConsistent bool
}

var hashOIDs = map[string]func() hash.Hash{
	"1.3.14.3.2.26":          sha1.New,
	"2.16.840.1.101.3.4.2.1": sha256.New,
	"2.16.840.1.101.3.4.2.2": sha512.New384,
	"2.16.840.1.101.3.4.2.3": sha512.New,
	"2.16.840.1.101.3.4.2.4": sha256.New224,
}

func ParseAndVerifySOD(efsod []byte, dataGroups map[int][]byte) (*SODResult, error) {
	tag, tagLen, err := parseTag(efsod)
	if err != nil {
		return nil, fmt.Errorf("nfcreader: parse EF.SOD tag: %w", err)
	}
	if tag != tagEFSOD {
		return nil, fmt.Errorf("nfcreader: EF.SOD: expected outer tag %#x, got %#x", tagEFSOD, tag)
	}
	length, lenLen, err := parseLength(efsod[tagLen:])
	if err != nil {
		return nil, fmt.Errorf("nfcreader: parse EF.SOD length: %w", err)
	}
	valueStart := tagLen + lenLen
	cmsData := efsod[valueStart : valueStart+length]

	p7, err := pkcs7.Parse(cmsData)
	if err != nil {
		return nil, fmt.Errorf("nfcreader: EF.SOD is not a valid CMS SignedData: %w", err)
	}

	var lds ldsSecurityObject
	if _, err := asn1.Unmarshal(p7.Content, &lds); err != nil {
		return nil, fmt.Errorf("nfcreader: EF.SOD content is not a valid LDSSecurityObject: %w", err)
	}

	newHash, ok := hashOIDs[lds.HashAlgorithm.Algorithm.String()]
	if !ok {
		return nil, fmt.Errorf("nfcreader: unsupported LDSSecurityObject hash algorithm OID %s", lds.HashAlgorithm.Algorithm.String())
	}

	result := &SODResult{
		HashAlgorithm:      lds.HashAlgorithm.Algorithm.String(),
		DataGroupHashValid: make(map[int]bool, len(dataGroups)),
	}

	declared := make(map[int][]byte, len(lds.DataGroupHashes))
	for _, dgh := range lds.DataGroupHashes {
		declared[dgh.DataGroupNumber] = dgh.HashValue
	}

	for dgNum, raw := range dataGroups {
		want, ok := declared[dgNum]
		if !ok {
			result.DataGroupHashValid[dgNum] = false
			continue
		}
		h := newHash()
		h.Write(raw)
		got := h.Sum(nil)
		result.DataGroupHashValid[dgNum] = constantTimeEqual(got, want)
	}

	result.SignatureSelfConsistent = p7.Verify() == nil

	return result, nil
}

func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
