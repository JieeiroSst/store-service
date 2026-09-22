package model

type ChipReadResult struct {
	DocumentNumber   string
	Surname          string
	GivenNames       string
	Nationality      string
	DateOfBirth      string
	Sex              string
	DateOfExpiry     string
	MRZChecksumValid bool

	FaceImage []byte

	PresentDataGroups []int

	DataGroupHashValid map[int]bool

	SODSignatureSelfConsistent bool
}

func (r ChipReadResult) AllHashesValid() bool {
	if len(r.DataGroupHashValid) == 0 {
		return false
	}
	for _, ok := range r.DataGroupHashValid {
		if !ok {
			return false
		}
	}
	return true
}
