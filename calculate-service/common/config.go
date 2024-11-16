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
