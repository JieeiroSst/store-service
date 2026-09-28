package application

import (
	"context"
	"fmt"
	"net/mail"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

type campaignService struct {
	campaigns port.CampaignRepository
	batches   port.CampaignBatchRepository
	template  port.TemplateRenderer
	settings  CampaignSettings
}

func NewCampaignService(
	campaigns port.CampaignRepository,
	batches port.CampaignBatchRepository,
	template port.TemplateRenderer,
	settings CampaignSettings,
) port.CampaignUsecase {
	return &campaignService{campaigns: campaigns, batches: batches, template: template, settings: settings}
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", common.ErrInvalidRequest, fmt.Sprintf(format, args...))
}

func (s *campaignService) validate(in *port.CreateCampaignInput) error {
	in.Title = strings.TrimSpace(in.Title)
	in.Topic = strings.TrimSpace(in.Topic)
	switch in.Channel {
	case model.CampaignChannelPush:
		if in.Audience != model.AudienceAllDevices && in.Audience != model.AudienceUsers && in.Audience != model.AudienceTopic {
			return invalid("push audience must be all_devices, users or topic")
		}
		if in.Title == "" || strings.TrimSpace(in.Message) == "" {
			return invalid("title and message are required")
		}
	case model.CampaignChannelEmail:
		if in.Audience != model.AudienceEmails {
			return invalid("email audience must be emails")
		}
		if in.TemplateType == "" && (in.Title == "" || strings.TrimSpace(in.Message) == "") {
			return invalid("email needs template_type, or title and message")
		}
		if in.TemplateType != "" {
			if _, _, err := s.template.Render(in.TemplateType, in.TemplateData); err != nil {
				return invalid("template: %v", err)
			}
		}
	default:
		return invalid("channel must be push or email")
	}

	switch in.Audience {
	case model.AudienceUsers:
		if len(in.UserIDs) == 0 || len(in.UserIDs) > s.settings.MaxRecipients {
			return invalid("user_ids must contain 1 to %d ids", s.settings.MaxRecipients)
		}
		for _, id := range in.UserIDs {
			if id == 0 {
				return invalid("user_ids must not contain 0")
			}
		}
	case model.AudienceEmails:
		if len(in.Emails) == 0 || len(in.Emails) > s.settings.MaxRecipients {
			return invalid("emails must contain 1 to %d addresses", s.settings.MaxRecipients)
		}
		for i, e := range in.Emails {
			addr, err := mail.ParseAddress(strings.TrimSpace(e))
			if err != nil {
				return invalid("emails[%d] %q is not a valid address", i, e)
			}
			in.Emails[i] = strings.ToLower(addr.Address)
		}
	case model.AudienceTopic:
		if !model.ValidTopic(in.Topic) {
			return invalid("topic must match [a-zA-Z0-9-_.~%%]+")
		}
	}
	return nil
}

func (s *campaignService) Create(ctx context.Context, in port.CreateCampaignInput) (*model.Campaign, error) {
	if err := s.validate(&in); err != nil {
		return nil, err
	}
	c := &model.Campaign{
		Channel:      in.Channel,
		Audience:     in.Audience,
		Topic:        in.Topic,
		Title:        in.Title,
		Message:      in.Message,
		Data:         in.Data,
		TemplateType: in.TemplateType,
		TemplateData: in.TemplateData,
		Status:       model.CampaignDraft,
		RequestedBy:  requester(in.RequestedBy),
	}
	if err := s.campaigns.Create(ctx, c); err != nil {
		return nil, err
	}

	var recipients []model.CampaignRecipient
	switch in.Audience {
	case model.AudienceUsers:
		recipients = make([]model.CampaignRecipient, 0, len(in.UserIDs))
		for _, id := range in.UserIDs {
			recipients = append(recipients, model.CampaignRecipient{CampaignID: c.ID, RecipientKey: "u:" + strconv.FormatUint(uint64(id), 10), UserID: id})
		}
	case model.AudienceEmails:
		recipients = make([]model.CampaignRecipient, 0, len(in.Emails))
		for _, e := range in.Emails {
			recipients = append(recipients, model.CampaignRecipient{CampaignID: c.ID, RecipientKey: "e:" + e, Email: e})
		}
	}
	inserted, err := s.campaigns.AddRecipients(ctx, recipients)
	if err != nil {
		return nil, err
	}
	if err := s.campaigns.Activate(ctx, c.ID, int(inserted)); err != nil {
		return nil, err
	}
	return s.campaigns.GetByID(ctx, c.ID)
}

func (s *campaignService) Get(ctx context.Context, id uint) (*port.CampaignView, error) {
	c, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p, err := s.campaigns.Progress(ctx, id)
	if err != nil {
		return nil, err
	}
	return &port.CampaignView{Campaign: *c, Progress: p}, nil
}

func (s *campaignService) List(ctx context.Context, limit, offset int) ([]model.Campaign, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.campaigns.List(ctx, limit, max(offset, 0))
}

func (s *campaignService) Cancel(ctx context.Context, id uint) (*port.CampaignView, error) {
	c, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status.Finished() {
		return nil, invalid("campaign %d is already %s", id, c.Status)
	}
	if _, err := s.campaigns.Cancel(ctx, id); err != nil {
		return nil, err
	}
	if err := s.batches.CancelPending(ctx, id); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func requester(v string) string {
	if v = strings.TrimSpace(v); v != "" {
		return v
	}
	return model.DefaultRequester
}
