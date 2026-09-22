package nfcreader

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type reader struct{}

func NewNFCReader() port.NFCReader {
	return &reader{}
}

func (r *reader) ReadChip(ctx context.Context, dump port.ChipDump) (*model.ChipReadResult, error) {
	result := &model.ChipReadResult{}
	dataGroups := make(map[int][]byte)

	if len(dump.EFCOM) > 0 {
		com, err := ParseEFCOM(dump.EFCOM)
		if err != nil {
			return nil, err
		}
		result.PresentDataGroups = com.PresentDataGroups
	}

	if len(dump.DG1) > 0 {
		fields, _, err := ParseDG1(dump.DG1)
		if err != nil {
			return nil, err
		}
		result.DocumentNumber = fields.DocumentNumber
		result.Surname = fields.Surname
		result.GivenNames = fields.GivenNames
		result.Nationality = fields.Nationality
		result.DateOfBirth = fields.DateOfBirth
		result.Sex = fields.Sex
		result.DateOfExpiry = fields.DateOfExpiry
		result.MRZChecksumValid = fields.DocNumberValid && fields.DOBValid && fields.ExpiryValid && fields.CompositeValid
		dataGroups[1] = dump.DG1
	}

	if len(dump.DG2) > 0 {
		photo, err := ParseDG2Photo(dump.DG2)
		if err != nil {
			return nil, err
		}
		result.FaceImage = photo
		dataGroups[2] = dump.DG2
	}

	if len(dump.EFSOD) > 0 {
		if len(dataGroups) == 0 {
			return nil, fmt.Errorf("nfcreader: EF.SOD submitted with no data groups to verify against it")
		}
		sod, err := ParseAndVerifySOD(dump.EFSOD, dataGroups)
		if err != nil {
			return nil, err
		}
		result.DataGroupHashValid = sod.DataGroupHashValid
		result.SODSignatureSelfConsistent = sod.SignatureSelfConsistent
	}

	return result, nil
}
