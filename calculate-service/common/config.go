package common

type CampaignConfigType int

const (
	Draft CampaignConfigType = iota + 1
	Active
	Cancel
)

func (c CampaignConfigType) Value() int {
	return int(c)
}

type CampaignType string

const (
	V1 CampaignType = "V1"
	V2 CampaignType = "V2"
)

func (c CampaignType) Value() string {
	return string(c)
}
