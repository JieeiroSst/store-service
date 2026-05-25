package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"go.uber.org/zap"
)

type Config struct {
	Provider  string // "sendgrid" | "smtp"
	APIKey    string
	FromEmail string
	FromName  string
}

type sendGridService struct {
	cfg        Config
	httpClient *http.Client
	logger     *zap.Logger
}

func NewSendGridService(cfg Config, logger *zap.Logger) port.NotificationService {
	return &sendGridService{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		logger:     logger,
	}
}

var emailTemplates = map[string]struct{ Subject, Body string }{
	"application_received": {
		Subject: "Đơn ứng tuyển của bạn đã được nhận",
		Body:    "Cảm ơn bạn đã ứng tuyển. Chúng tôi sẽ liên hệ sớm!",
	},
	"cv_under_review": {
		Subject: "CV của bạn đang được xem xét",
		Body:    "CV của bạn hiện đang được recruiter xem xét.",
	},
	"interview_invite": {
		Subject: "Lịch phỏng vấn của bạn",
		Body:    "Bạn được mời tham gia buổi phỏng vấn.",
	},
	"interview_reminder_24h": {
		Subject: "Nhắc nhở: Phỏng vấn ngày mai",
		Body:    "Phỏng vấn của bạn sẽ diễn ra vào ngày mai. Chúc bạn may mắn!",
	},
	"offer_extended": {
		Subject: "Bạn nhận được offer!",
		Body:    "Chúc mừng! Chúng tôi đã gửi offer letter cho bạn.",
	},
	"referral_commission_ready": {
		Subject: "Hoa hồng của bạn đã sẵn sàng",
		Body:    "Ứng viên bạn giới thiệu đã được tuyển dụng. Hoa hồng đang chờ xử lý.",
	},
	"partner_welcome": {
		Subject: "Chào mừng đến với mạng lưới CTV!",
		Body:    "Tài khoản CTV của bạn đã được kích hoạt.",
	},
	"payout_requested": {
		Subject: "Yêu cầu rút tiền đã được ghi nhận",
		Body:    "Yêu cầu rút hoa hồng của bạn đang được xử lý.",
	},
}

func (s *sendGridService) Send(ctx context.Context, n port.NotificationPayload) error {
	tpl, ok := emailTemplates[n.TemplateID]
	if !ok {
		s.logger.Warn("unknown template", zap.String("template_id", n.TemplateID))
		tpl.Subject = n.TemplateID
		tpl.Body = fmt.Sprintf("Notification: %v", n.Data)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{"to": []map[string]string{{"email": n.RecipientID.String()}}},
		},
		"from":    map[string]string{"email": s.cfg.FromEmail, "name": s.cfg.FromName},
		"subject": tpl.Subject,
		"content": []map[string]string{
			{"type": "text/plain", "value": tpl.Body},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.sendgrid.com/v3/mail/send",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification: sendgrid request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("notification: sendgrid error status %d", resp.StatusCode)
	}

	s.logger.Info("notification sent",
		zap.String("recipient", n.RecipientID.String()),
		zap.String("template", n.TemplateID),
	)
	return nil
}

func (s *sendGridService) SendBulk(ctx context.Context, ns []port.NotificationPayload) error {
	for _, n := range ns {
		if err := s.Send(ctx, n); err != nil {
			s.logger.Error("bulk send failed for recipient",
				zap.String("recipient", n.RecipientID.String()),
				zap.Error(err),
			)
		}
	}
	return nil
}
