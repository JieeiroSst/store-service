package http

import "github.com/referral/service/internal/core/domain"

type utmParams struct {
	UTMSource   string `json:"utm_source"   form:"utm_source"`
	UTMMedium   string `json:"utm_medium"   form:"utm_medium"`
	UTMCampaign string `json:"utm_campaign" form:"utm_campaign"`
	UTMTerm     string `json:"utm_term"     form:"utm_term"`
	UTMContent  string `json:"utm_content"  form:"utm_content"`
}

func (p utmParams) toDomain() domain.UTM {
	return domain.UTM{
		Source:   p.UTMSource,
		Medium:   p.UTMMedium,
		Campaign: p.UTMCampaign,
		Term:     p.UTMTerm,
		Content:  p.UTMContent,
	}.Normalize()
}

type generateLinkRequest struct {
	OwnerUserID string `json:"owner_user_id" binding:"required"`
	Channel     string `json:"channel"`  // "copy" | "whatsapp" | "facebook" | "instagram"
	Platform    string `json:"platform"` // "ios" | "android" | "universal"
	utmParams
}

type trackEventRequest struct {
	RefCode            string `json:"ref_code"   binding:"required"`
	EventType          string `json:"event_type" binding:"required"`
	Platform           string `json:"platform"`
	Channel            string `json:"channel"`
	NewUserID          string `json:"new_user_id"`
	IPAddress          string `json:"ip_address"`
	DeviceID           string `json:"device_id"`
	UserAgent          string `json:"user_agent"`
	FirebaseInstanceID string `json:"firebase_instance_id"`
	utmParams
}

type confirmInstallRequest struct {
	RefCode            string `json:"ref_code"`
	ClickID            string `json:"click_id"`
	NewUserID          string `json:"new_user_id" binding:"required"`
	Platform           string `json:"platform"`
	DeviceID           string `json:"device_id"`
	FirebaseInstanceID string `json:"firebase_instance_id"`
	utmParams
}

type activateReferralRequest struct {
	RefCode            string `json:"ref_code"  binding:"required"`
	ClickID            string `json:"click_id"`
	UserID             string `json:"user_id"   binding:"required"`
	Platform           string `json:"platform"`
	DeviceID           string `json:"device_id"`
	FirebaseInstanceID string `json:"firebase_instance_id"`
	utmParams
}

type rewardTierInput struct {
	MinCount    int64   `json:"min_count"    binding:"required"`
	MaxCount    int64   `json:"max_count"`
	RewardValue float64 `json:"reward_value" binding:"required"`
}

type createRewardProgramRequest struct {
	Name     string            `json:"name"     binding:"required"`
	Activate bool              `json:"activate"`
	Tiers    []rewardTierInput `json:"tiers" binding:"required,min=1"`
}

type setRewardProgramStatusRequest struct {
	Active bool `json:"active"`
}

var validChannels = map[domain.Channel]bool{
	domain.ChannelCopy:      true,
	domain.ChannelWhatsApp:  true,
	domain.ChannelFacebook:  true,
	domain.ChannelInstagram: true,
	domain.ChannelZalo:      true,
	domain.ChannelOther:     true,
}

// shareChannels lists the channels surfaced by GetShareTargets, in display order.
var shareChannels = []struct {
	Channel domain.Channel
	Label   string
}{
	{domain.ChannelZalo, "Zalo"},
	{domain.ChannelFacebook, "Facebook"},
	{domain.ChannelWhatsApp, "WhatsApp"},
	{domain.ChannelInstagram, "Instagram"},
	{domain.ChannelCopy, "Copy link"},
}

var validEventTypes = map[domain.EventType]bool{
	domain.EventLinkCopied:   true,
	domain.EventLinkClicked:  true,
	domain.EventAppInstalled: true,
	domain.EventRegistered:   true,
	domain.EventRewardGiven:  true,
}

// --- response structs (adapter-layer only, not domain types) ---

type generateLinkResponse struct {
	RefCode   string `json:"ref_code"`
	DeepLink  string `json:"deep_link"`
	ShareURL  string `json:"share_url"`
	ExpiresAt int64  `json:"expires_at"`
	domain.UTM
}

type utmReportQuery struct {
	OwnerUserID string `form:"owner_user_id"`
	From        int64  `form:"from"`
	To          int64  `form:"to"`
	utmParams
}

type attributionResponse struct {
	Attributed  bool              `json:"attributed"`
	OwnerUserID string            `json:"owner_user_id,omitempty"`
	RewardType  domain.RewardType `json:"reward_type,omitempty"`
}

type shareTarget struct {
	Channel  string `json:"channel"`
	Label    string `json:"label"`
	ShareURL string `json:"share_url"`
	QRURL    string `json:"qr_url"`
	domain.UTM
}
